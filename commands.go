package main

import (
	"fmt"
	"go/build"
	"os"

	"github.com/urfave/cli"
)

type Options struct {
	initSQLFilePath  string
	checkSQLFilePath string
	host             string
	user             string
	passwd           string
	db               string
	port             int
}

var Opt Options
var mytxPath = build.Default.GOPATH + "/src/github.com/tom--bo/mytx/"

var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "initsql, i",
		Value:       mytxPath + "samples/sql/init.sql",
		Usage:       "",
		Destination: &Opt.initSQLFilePath,
	},
	cli.StringFlag{
		Name:        "checksql, c",
		Value:       mytxPath + "samples/sql/check.sql",
		Usage:       "",
		Destination: &Opt.checkSQLFilePath,
	},
	cli.StringFlag{
		Name:        "host, H",
		Value:       "localhost",
		Usage:       "",
		Destination: &Opt.host,
	},
	cli.StringFlag{
		Name:        "user, u",
		Value:       "mysql",
		Usage:       "",
		Destination: &Opt.user,
	},
	cli.StringFlag{
		Name:        "pasword, p",
		Value:       "mysql",
		Usage:       "",
		Destination: &Opt.passwd,
	},
	cli.StringFlag{
		Name:        "database, db, d",
		Value:       "sample",
		Usage:       "",
		Destination: &Opt.db,
	},
	cli.IntFlag{
		Name:        "port, P",
		Value:       3306,
		Usage:       "",
		Destination: &Opt.port,
	},
}

var Commands = []cli.Command{}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
