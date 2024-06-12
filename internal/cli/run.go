package cli

import (
	"flag"
)

// Run run command
func RunCmd(args []string) error {
	// Flags and help for the command
	var verbose bool
	var configFile string
	var workingDirectory string
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.BoolVar(&verbose, "verbose", false, "The flag to set in order to get verbose output")
	flagSet.StringVar(&configFile, "config", "gonovate.json", "The path to the config file to read")
	flagSet.StringVar(&workingDirectory, "workDir", "", "The path to the working directory")
	flagSet.Usage = func() { printCmdUsage(flagSet, "run", "") }
	flagSet.Parse(args)

	return nil
}
