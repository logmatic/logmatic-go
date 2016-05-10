package logmatic

import (
	"crypto/tls"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"errors"
	"sync"
)

const network = "tcp"
const raddr = "api.logmatic.io:10515"
const maxRetries = 3

// LogmaticHook to send logs via syslog protocol.
type LogmaticHook struct {
	LogmaticEndpoint *tls.Conn
	LogmaticNetwork  string
	LogmaticRaddr    string
	LogmaticApiKey   string

	mu sync.Mutex
}

// Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewSyslogHook("udp", "localhost:514", syslog.LOG_DEBUG, "")`
// `if err == nil { log.Hooks.Add(hook) }`
func NewLogmaticHook(apiKey string) *LogmaticHook {

	// connect to this socket
	conn, _ := tls.Dial(network, raddr, &tls.Config{})
	return &LogmaticHook{conn, network, raddr, apiKey, sync.Mutex{}}
}

func (hook *LogmaticHook) Fire(entry *logrus.Entry) error {

	msg, _ := entry.String()
	payload := fmt.Sprintf("%s %s", hook.LogmaticApiKey, msg)

	_, err := hook.writeAndRetry([]byte(payload))
	if err != nil {
		return err
	}

	return nil
}

func (hook *LogmaticHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *LogmaticHook) writeAndRetry(b []byte) (int, error) {

	for i := 0; i < maxRetries; i++ {

		if hook.LogmaticEndpoint == nil {

			hook.mu.Lock()

			// reconnect
			conn, err := tls.Dial(hook.LogmaticNetwork, hook.LogmaticRaddr, &tls.Config{})
			hook.LogmaticEndpoint = conn
			if err != nil {
				hook.LogmaticEndpoint = nil
				continue
			}

			hook.mu.Unlock();

		}
		n, err := hook.LogmaticEndpoint.Write(b)
		if err == nil {
			return n, err
		} else {
			fmt.Fprintf(os.Stderr, "Unable to send log line. Wrote %d bytes before error: %v\n", n, err)
			fmt.Fprintf(os.Stderr, "Making a new attempt\n")
			hook.LogmaticEndpoint = nil

		}
	}

	return 0, errors.New("Failed to connect to Logmatic.io")
}
