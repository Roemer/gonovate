package config

// This type represents the root config object.
type Config struct {
	Extends []string `json:"extends"`
}
