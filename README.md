# go-slog

Package slog is a fork of the standard go log and syslog packages. It is a
drop in replacement for the default log package and consolidates logging to
stdout/syslog/file with log levels. However, it does not instantiate a
standard logger. File logger supports log file rotation on SIGHUP.

Work in progress...

## Usage

```go
import slog "github.com/bartke/go-slog"
```

Setup and use a standard logger by defining a global:

```go
var log = logger.New(os.Stderr, "", LstdFlags, LOG_INFO)

// and use it as usual
log.Error("something went wrong")
```

### Example

Simple example that shows how to enable syslog/logfile and setting the log
level with flags.

```go
package main

import (
	"flag"
	"fmt"
	"os"

	slog "github.com/bartke/go-slog"
)

// setup our new standard logger
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
```

