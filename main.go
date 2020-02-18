package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sslhound/sigsink/server"
)

var ReleaseCode string
var GitCommit string
var BuildTime string

func main() {
	compiledAt, err := time.Parse(time.RFC822Z, BuildTime)
	if err != nil {
		compiledAt = time.Now()
	}
	if ReleaseCode == "" {
		ReleaseCode = "na"
	}
	if GitCommit == "" {
		GitCommit = "na"
	}

	app := cli.NewApp()
	app.Name = "sigsink"
	app.Usage = "An http signature verification tool."
	app.Version = fmt.Sprintf("%s-%s", ReleaseCode, GitCommit)
	app.Compiled = compiledAt
	app.Copyright = "(c) 2020 SSL Hound, LLC"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen",
			Usage:   "Configure the server to listen to this interface.",
			EnvVars: []string{"LISTEN"},
			Value:   "0.0.0.0:7000",
		},
		&cli.StringFlag{
			Name:    "domain",
			Usage:   "Set the website domain.",
			Value:   "sigsink.sslhound.com",
			EnvVars: []string{"DOMAIN"},
		},
		&cli.StringFlag{
			Name:    "environment",
			Usage:   "Set the environment the application is running in.",
			EnvVars: []string{"ENVIRONMENT"},
			Value:   "development",
		},
		&cli.BoolFlag{
			Name:    "enable-keyfetch",
			Usage:   "Enable fetching remote keys",
			EnvVars: []string{"ENABLE_KEYFETCH"},
			Value:   false,
		},
		&cli.StringSliceFlag{
			Name:    "key-source",
			Usage:   "A location to load keys from",
			EnvVars: []string{"KEY_SOURCE"},
			Value:   cli.NewStringSlice("./keys"),
		},
	}

	app.Action = server.Action

	sort.Sort(cli.FlagsByName(app.Flags))

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
