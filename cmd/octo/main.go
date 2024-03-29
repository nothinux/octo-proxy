package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nothinux/octo-proxy/pkg/config"
	"github.com/nothinux/octo-proxy/pkg/runner"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var banner = `         _                           
 ___ ___| |_ ___ ___ ___ ___ _ _ _ _ 
| . |  _|  _| . | . |  _| . |_'_| | |
|___|___|_| |___|  _|_| |___|_,_|_  |
                |_|             |___| v%s

`
var usage = `Usage of octo:
octo [flag] arguments...

Flags:
  -config
    Specify config location path (default: ./config.yaml)
  -listener
    Specify listener for running octo-proxy (default: 0.0.0.0:5000)
  -target
    Specify target backend which traffic will be forwarded
  -metrics
    Specify address and port to run the metrics server
  -debug
    Enable debug log messages
  -version
    Print octo-proxy version

`

var (
	Version    = "x.X"
	showBanner = fmt.Sprintf(banner, Version)
)

func main() {
	if err := runMain(); err != nil {
		log.Fatal().Err(err).Msg("failed to start")
	}
}

func setupLogger(debug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro

	consoleWriter := zerolog.ConsoleWriter{
		TimeFormat: time.StampMicro,
		Out:        os.Stdout,
	}

	log.Logger = log.Output(consoleWriter)

	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

func runMain() error {
	var (
		configPath = flag.String("config", "config.yaml", "Specify config location path")
		ver        = flag.Bool("version", false, "Print octo-proxy version")
		listener   = flag.String("listener", "127.0.0.1:5000", "Specify listener for running octo-proxy")
		target     = flag.String("target", "", "Specify comma-separated list of targets for running octo-proxy")
		metrics    = flag.String("metrics", "0.0.0.0:9123", "Address and port to run the metrics server on")
		debug      = flag.Bool("debug", false, "Enable debug messages")
	)

	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), showBanner, usage)
	}
	flag.Parse()

	fmt.Fprintf(os.Stdout, showBanner)

	setupLogger(*debug)

	// run with flag
	if *target != "" {
		targets := strings.Split(*target, ",")

		c, err := config.GenerateConfig(*listener, targets, *metrics)
		if err != nil {
			return err
		}

		if err := runner.Run(c, ""); err != nil {
			return err
		}

		return nil
	}

	if *ver {
		fmt.Printf("octo-proxy version v%s\n", Version)
		return nil
	}

	// running with configuration, first read configuration and then
	// run the runner
	c, err := config.New(*configPath)
	if err != nil {
		return err
	}

	if err := runner.Run(c, *configPath); err != nil {
		return err
	}

	return nil
}
