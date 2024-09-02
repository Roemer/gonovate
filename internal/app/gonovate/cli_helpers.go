package gonovate

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type stringSliceFlag []string

func (i *stringSliceFlag) String() string {
	return strings.Join(*i, "; ")
}

func (i *stringSliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Prints the help for a command
func printCmdUsage(flagSet *flag.FlagSet, commandName, nonFlagArgs string) {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "  gonovate %s [flags]", commandName)
	if nonFlagArgs != "" {
		fmt.Fprint(os.Stderr, " "+nonFlagArgs)
	}
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Flags:")
	flagSet.PrintDefaults()
}
