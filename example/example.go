package main

import (
	"flag"
	"fmt"
	"os"

	slog "github.com/bartke/go-slog"
)

var log *slog.Logger

func main() {
	var err error
	var logLevel int
	var logSyslog bool
	var logFile string

	flag.IntVar(&logLevel, "log-level", slog.LOG_INFO, "sets the log level")
	flag.BoolVar(&logSyslog, "log-syslog", false, "enable logging to syslog")
	flag.StringVar(&logFile, "log-file", "", "enable logging to a log file")
	flag.Parse()

	// Logging to stdout
	log = slog.New(os.Stderr, "my_app", slog.LstdFlags|slog.Llevel)
	log.SetLogLevel(slog.Priority(logLevel))

	if logSyslog {
		// Logging to syslog
		log, err = slog.NewSyslog(slog.LOG_LOCAL2|slog.Priority(logLevel), "my_app")
		if err != nil {
			fmt.Print(err)
		}
	} else if logFile != "" {
		// Logging to file
		log, err = slog.NewLogfile(logFile, 0666, "my_app", slog.LstdFlags|slog.Llevel)
		if err != nil {
			fmt.Print(err)
		}
		// we can always change the log level and enable/disable messages
		log.SetLogLevel(slog.Priority(logLevel))
	}

	// example logging
	log.Println("println at default log level")

	//log.Emerg("system is unusable")
	log.Alert("action must be taken immediately")
	log.Crit("critical conditions")
	log.Err("error conditions")
	log.Warning("warning conditions")
	log.Notice("normal but significant condition")
	log.Info("informational")
	log.Debug("debug-level messages")

	log.Fatal("fatal at fatal log-level")
}
