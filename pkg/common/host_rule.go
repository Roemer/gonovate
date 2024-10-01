package common

import "os"

type HostRule struct {
	MatchHost string
	Username  string
	Password  string
	Token     string
}

func (hr *HostRule) UsernameExpanded() string {
	return os.ExpandEnv(hr.Username)
}

func (hr *HostRule) PasswordExpanded() string {
	return os.ExpandEnv(hr.Password)
}

func (hr *HostRule) TokendExpanded() string {
	return os.ExpandEnv(hr.Token)
}
