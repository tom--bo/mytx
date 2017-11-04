package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/tom--bo/mytx/command"
)

var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		EnvVar: "ENV_C",
		Name:   "c",
		Value:  "",
		Usage:  "",
	},
	cli.StringFlag{
		EnvVar: "ENV_W",
		Name:   "w",
		Value:  "",
		Usage:  "",
	},
}

var Commands = []cli.Command{}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
