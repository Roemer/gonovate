package gonovate

import (
	"flag"
	"fmt"
	"os"
)

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
