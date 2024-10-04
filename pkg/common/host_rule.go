package common

import "os"

// A host rule that is applied when using a certain host.
type HostRule struct {
	// The host that needs to match in order to use this rule.
	MatchHost string `json:"matchHost" yaml:"matchHost"`
	// The username to authenticate with the host.
	Username string `json:"username" yaml:"username"`
	// The password to authenticate with the host.
	Password string `json:"password" yaml:"password"`
	// A token to authenticate with the host.
	Token string `json:"token" yaml:"token"`
}

// Expands the username with environment variables.
func (hr *HostRule) UsernameExpanded() string {
	return os.ExpandEnv(hr.Username)
}

// Expands the password with environment variables.
func (hr *HostRule) PasswordExpanded() string {
	return os.ExpandEnv(hr.Password)
}

// Expands the token with environment variables.
func (hr *HostRule) TokendExpanded() string {
	return os.ExpandEnv(hr.Token)
}
