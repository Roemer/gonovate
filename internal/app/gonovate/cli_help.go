package gonovate

import "flag"

func HelpCmd(_ []string) error {
	flag.Usage()
	return nil
}
