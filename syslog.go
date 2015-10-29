package slog

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Facilities from /usr/include/sys/syslog.h.
// These are the same up to Lfpp on Linux, BSD, and OS X.
const (
	LOG_KERN     Priority = iota << 3 // kernel messages
	LOG_USER                          // random user-level messages
	LOG_MAIL                          // mail system
	LOG_DAEMON                        // system daemons
	LOG_AUTH                          // security/authorization messages
	LOG_SYSLOG                        // messages generated internally by syslogd
	LOG_LPR                           // line printer subsystem
	LOG_NEWS                          // network news subsystem
	LOG_UUCP                          // UUCP subsystem
	LOG_CRON                          // clock daemon
	LOG_AUTHPRIV                      // security/authorization messages (private)
	LOG_FTP                           // FTP server
	_                                 // other codes through 15 reserved for system use
	_
	_
	_
	LOG_LOCAL0 // reserved for local use
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

const severityMask = 0x07
const facilityMask = 0xf8

// This interface and the separate syslog_unix.go file exist for
// Solaris support as implemented by gccgo.  On Solaris you can not
// simply open a TCP connection to the syslog daemon.  The gccgo
// sources have a syslog_solaris.go file that implements unixSyslog to
// return a type that satisfies this interface and simply calls the C
// library syslog function.
type serverConn interface {
	writeString(p Priority, hostname, tag, s, nl string) error
	close() error
}

type netConn struct {
	local bool
	conn  net.Conn
}

func (n *netConn) close() error {
	return n.conn.Close()
}

func (n *netConn) writeString(p Priority, hostname, tag, msg, nl string) error {
	if n.local {
		// Compared to the network form below, the changes are:
		//	1. Use time.Stamp instead of time.RFC3339.
		//	2. Drop the hostname field from the Fprintf.
		timestamp := time.Now().Format(time.Stamp)
		_, err := fmt.Fprintf(n.conn, "<%d>%s %s[%d]: %s%s",
			p, timestamp,
			tag, os.Getpid(), msg, nl)
		return err
	}
	timestamp := time.Now().Format(time.RFC3339)
	_, err := fmt.Fprintf(n.conn, "<%d>%s %s %s[%d]: %s%s",
		p, timestamp, hostname,
		tag, os.Getpid(), msg, nl)
	return err
}

// NewSyslog for compatibility with standard syslog package.
func NewSyslog(priority Priority, tag string) (*Logger, error) {
	return Dial("", "", priority, tag)
}

// Dial creates a new Logger setup for syslog.
func Dial(network, raddr string, priority Priority, tag string) (*Logger, error) {
	if priority < 0 || priority > LOG_LOCAL7|LOG_DEBUG {
		return nil, errors.New("slog: invalid priority")
	}

	if tag == "" {
		tag = os.Args[0]
	}
	hostname, _ := os.Hostname()

	l := &Logger{
		network:         network,
		raddr:           raddr,
		tag:             tag,
		facility:        priority & facilityMask,
		level:           priority & severityMask,
		hostname:        hostname,
		logType:         tsyslog,
		DefaultLogLevel: priority & severityMask,
		FatalLogLevel:   LOG_CRIT,
		PanicLogLevel:   LOG_ALERT,
	}

	l.Lock()
	defer l.Unlock()

	err := l.connect()
	if err != nil {
		return nil, err
	}
	return l, err
}

// connect makes a connection to the syslog server.
// It must be called with w.mu held.
func (l *Logger) connect() (err error) {
	if l.conn != nil {
		// ignore err from close, it makes sense to continue anyway
		l.conn.close()
		l.conn = nil
	}

	if l.network == "" {
		l.conn, err = unixSyslog()
		if l.hostname == "" {
			l.hostname = "localhost"
		}
	} else {
		var c net.Conn
		c, err = net.Dial(l.network, l.raddr)
		if err == nil {
			l.conn = &netConn{conn: c}
			if l.hostname == "" {
				l.hostname = c.LocalAddr().String()
			}
		}
	}
	return
}

func (l *Logger) syslogWriteAndRetry(p Priority, s string) (int, error) {
	pr := (l.facility & facilityMask) | (p & severityMask)

	l.Lock()
	defer l.Unlock()

	if l.conn != nil {
		if n, err := l.syslogWrite(pr, s); err == nil {
			return n, err
		}
	}
	if err := l.connect(); err != nil {
		return 0, err
	}
	return l.syslogWrite(pr, s)
}

// write generates and writes a syslog formatted string. The
// format is as follows: <PRI>TIMESTAMP HOSTNAME TAG[PID]: MSG
func (l *Logger) syslogWrite(p Priority, msg string) (int, error) {
	// ensure it ends in a \n
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}

	err := l.conn.writeString(p, l.hostname, l.tag, msg, nl)
	if err != nil {
		return 0, err
	}
	// Note: return the length of the input, not the number of
	// bytes printed by Fprintf, because this must behave like
	// an io.Writer.
	return len(msg), nil
}
