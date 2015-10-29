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

