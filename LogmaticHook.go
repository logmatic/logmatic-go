package logmatic

import (
	"crypto/tls"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
)

const network = "tcp"
const raddr = "api.logmatic.io:10515"

// LogmaticHook to send logs via syslog protocol.
type LogmaticHook struct {
	LogmaticEndpoint *tls.Conn
	LogmaticNetwork  string
	LogmaticRaddr    string
	LogmaticApiKey   string
}

// Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewSyslogHook("udp", "localhost:514", syslog.LOG_DEBUG, "")`
// `if err == nil { log.Hooks.Add(hook) }`
func NewLogmaticHook(apiKey string) *LogmaticHook {

	// connect to this socket
	conn, _ := tls.Dial(network, raddr, &tls.Config{})
	return &LogmaticHook{conn, network, raddr, apiKey}
}

func (hook *LogmaticHook) Fire(entry *logrus.Entry) error {

	msg, _ := entry.String()
	payload := fmt.Sprintf("%s %s", hook.LogmaticApiKey, msg)

	bytesWritten, err := hook.writeAndRetry([]byte(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to send log line. Wrote %d bytes before error: %v", bytesWritten, err)
		return err
	}

	return nil
}

func (hook *LogmaticHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *LogmaticHook) writeAndRetry(b []byte) (int, error) {

	if hook.LogmaticEndpoint != nil {

		n, err := hook.LogmaticEndpoint.Write(b)
		if err == nil {
			return n, err
		} else {
			fmt.Fprintf(os.Stderr, "Unable to send log line. Wrote %d bytes before error: %v\n", n, err)
 			fmt.Fprintf(os.Stderr, "Making a new attempt\n")
		}
	}
	// reconnect
	conn, err := tls.Dial(hook.LogmaticNetwork, hook.LogmaticRaddr, &tls.Config{})
	if err != nil {
		return 0, err
	}
	hook.LogmaticEndpoint = conn

	return hook.LogmaticEndpoint.Write(b)

}
