package syslog

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Priority int

const serverityMask = 0x07
const facilityMask = 0x08

const (
	// Serverity 严重程度
	LOG_EMERG	0
	LOG_ALERT	1
	LOG_CRIT	2
	LOG_ERR		3
	LOG_WARNING	4
	LOG_NOTICE	5
	LOG_INFO	6
	LOG_DEBUG	7
)

const (
	LOG_KERN Priority = iota << 3
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_
	_
	_
	_
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

type Writer struct {
	priority Priority
	tag	string
	hostname string
	network string
	raddr string

	mu sync.Mutex
	conn serverConn
}

type serverConn interface {
	writeString(p Priority, hostname, tag, s, nl string) error
	close() error
}

type netConn struct {
	local bool
	conn net.Conn
}

func New(priority Priority, tag string) (* Writer, error) {
	return Dial("","", priority, tag)
}

func Dial(network, raddr string, priority Priority, tag string) (*Writer, error) {
	if priority < 0 || priority > LOG_LOCAL7 | LOG_DEBUG {
		return nil, errors.New("log/syslog: invalid priority")
	}

	if tag == "" {
		tag = os.Args[0]
	}
	hostname, _ := os.Hostname()

	w:=&Writer{
		priority: priority,
		tag: tag,
		hostname: hostname,
		network: network,
		raddr:raddr,
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.connect()
	if err != nil {
		return nil, err
	}
	return w, err
}

func (w *Writer) connect() (err error) {
	if w.conn != nil {
		w.conn.close()
		w.conn = nil
	}

	if w.network == "" {
		w.conn, err = unixSyslog()
		if w.hostname == "" {
			w.hostname = "localhost"
		}
	} else {
		var c net.Conn
		c, err = net.Dial(w.network, w.raddr)
		if err == nil {
			w.conn = &netConn{conn: c}
			if w.hostname == "" {
				w.hostname = c.LocalAddr().String()
			}
		}
	}
	return
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.writeAndRetry(w.priority, string(b))
}


