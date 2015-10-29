package slog

import (
	"os"
	"os/signal"
	"syscall"
)

// NewLogfile creates a new Logger setup for logfile.
func NewLogfile(filename string, perm os.FileMode, tag string, flag int) (*Logger, error) {
	l := New(nil, tag, flag)

	var err error
	l.file, err = os.OpenFile(filename,
		os.O_RDWR|os.O_APPEND|os.O_CREATE,
		perm,
	)
	if err != nil {
		return nil, err
	}

	l.SetOutput(l.file)

	// rotate on SIGHUP
	go l.logRotate(filename, perm)

	return l, nil
}

func (l *Logger) logRotate(filename string, perm os.FileMode) {
	var err error
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	for {
		<-signalChan
		l.Lock()
		// first close if open
		if l.file != nil {
			err := l.file.Close()
			if err != nil {
				l.Fatal(err)
			}
		}

		// reopen and append if existing or create
		l.file, err = os.OpenFile(filename,
			os.O_RDWR|os.O_APPEND|os.O_CREATE,
			perm,
		)
		if err != nil {
			l.Fatal(err)
		}
		l.Unlock()
	}
}
