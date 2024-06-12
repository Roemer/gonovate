package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

type ReadableTextHandler struct {
	options ReadableTextHandlerOptions
	mu      *sync.Mutex
	out     io.Writer
	groups  []handlerGroup
}

type ReadableTextHandlerOptions struct {
	Level slog.Leveler
}

type handlerGroup struct {
	name  string
	attrs []slog.Attr
}

func NewReadableTextHandler(out io.Writer, options *ReadableTextHandlerOptions) *ReadableTextHandler {
	handler := &ReadableTextHandler{out: out, mu: &sync.Mutex{}}
	if options == nil {
		options = &ReadableTextHandlerOptions{}
	}
	handler.options = *options
	if handler.options.Level == nil {
		handler.options.Level = slog.LevelInfo
	}
	// Create the root group
	handler.groups = []handlerGroup{{name: ""}}
	return handler
}

func (h *ReadableTextHandler) Handle(ctx context.Context, record slog.Record) error {
	// Prepare the log entry
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s|%s|%s|%s", record.Time.Format("2006.01.02"), record.Time.Format("15:04:05.000"), record.Level.String(), record.Message))

	// Process the groups and attributes added by "With/WithGroup" methods
	attrStrings := []string{}
	groupPrefix := ""
	for _, g := range h.groups {
		if g.name != "" {
			groupPrefix += g.name + "."
		}
		for _, a := range g.attrs {
			attrStrings = append(attrStrings, buildAttributes(a, groupPrefix)...)
		}
	}
	// Append the remaining attributes from this record
	record.Attrs(func(a slog.Attr) bool {
		attrStrings = append(attrStrings, buildAttributes(a, groupPrefix)...)
		return true
	})
	if len(attrStrings) > 0 {
		sb.WriteString("|")
		sb.WriteString(strings.Join(attrStrings, ", "))
	}
	sb.WriteString("\n")

	// Lock and write the log entry
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write([]byte(sb.String()))
	return err
}

func (h *ReadableTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

func (h *ReadableTextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groups = append(h2.groups, handlerGroup{name: name})
	return h2
}

func (h *ReadableTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := h.clone()
	groupToAdd := &h2.groups[len(h2.groups)-1]
	groupToAdd.attrs = append(groupToAdd.attrs, attrs...)
	return h2
}

func (h *ReadableTextHandler) clone() *ReadableTextHandler {
	h2 := *h
	h2.groups = make([]handlerGroup, len(h.groups))
	copy(h2.groups, h.groups)
	return &h2
}

func buildAttributes(a slog.Attr, groupPrefix string) []string {
	// Resolve the value of the attribute
	a.Value = a.Value.Resolve()
	// Ignore empty attributes
	if a.Equal(slog.Attr{}) {
		return nil
	}
	// Handle different attribute types
	switch a.Value.Kind() {
	case slog.KindGroup:
		attrs := a.Value.Group()
		// Ignore empty groups
		if len(attrs) == 0 {
			return nil
		}
		if a.Key != "" {
			groupPrefix += a.Key + "."
		}
		attrStrings := []string{}
		for _, a := range attrs {
			attrStrings = append(attrStrings, buildAttributes(a, groupPrefix)...)
		}
		return attrStrings
	default:
		// Simply add the key/value pair
		return []string{fmt.Sprintf("%s%s=%s", groupPrefix, a.Key, a.Value.String())}
	}
}
