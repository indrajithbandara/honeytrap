package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/honeytrap/honeytrap/config"
	"github.com/honeytrap/honeytrap/server"
	"github.com/minio/cli"
	"github.com/op/go-logging"
	"github.com/pkg/profile"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var Version = "0.1"

var helpTemplate = `NAME:
{{.Name}} - {{.Usage}}

DESCRIPTION:
{{.Description}}

USAGE:
{{.Name}} {{if .Flags}}[flags] {{end}}command{{if .Flags}}{{end}} [arguments...]

COMMANDS:
{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
{{end}}{{if .Flags}}
FLAGS:
{{range .Flags}}{{.}}
{{end}}{{end}}
VERSION:
` + Version +
	`{{ "\n"}}`

var log = logging.MustGetLogger("honeytrap/cmd")

var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "c,config",
		Usage: "config file",
		Value: "config.toml",
	},
	/*
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Show help.",
		},
	*/
	cli.BoolFlag{Name: "cpu-profile", Usage: "Enable cpu profiler"},
	cli.BoolFlag{Name: "mem-profile", Usage: "Enable memory profiler"},
	cli.BoolFlag{Name: "profiler", Usage: "Enable web profiler"},
}

type Cmd struct {
	*cli.App
}

func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("Honeytrap: The ultimate honeypot framework.")))
}

func serve(c *cli.Context) {
	conf, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stdout, err.Error())
		return
	}

	configFile := c.GlobalString("config")
	if err := conf.Load(configFile); err != nil {
		fmt.Fprintf(os.Stdout, err.Error())
		return
	}

	var profiler interface {
		Stop()
	} = nil

	if c.GlobalBool("cpu-profile") {
		log.Info("CPU profiler started.")
		profiler = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	} else if c.GlobalBool("mem-profile") {
		log.Info("Memory profiler started.")
		profiler = profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	}

	if c.GlobalBool("profiler") {
		log.Info("Profiler listening.")

		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	var server = server.New(conf)
	server.Serve()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)

	<-s

	if profiler != nil {
		profiler.Stop()
	}

	log.Info("Stopping honeytrap....")

	os.Exit(0)
}

func New() *Cmd {
	app := cli.NewApp()
	app.Name = "honeytrap"
	app.Author = ""
	app.Usage = "honeytrap"
	app.Description = `The ultimate honeypot framework.`
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = serve

	return &Cmd{
		App: app,
	}
}
