package logmatic

import (
	"crypto/tls"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"errors"
	"sync"
	"time"
)

const protocol = "tcp"
const endpoint = "api.logmatic.io:10515"
const maxRetries = 5
const sleepTime = 2

// LogmaticHook to send logs via syslog protocol.
type LogmaticHook struct {
	conn           *tls.Conn
	LogmaticApiKey string
	endpoint       string
	maxRetries     int
	maxSleepTime   int

	mu             sync.Mutex
}

// Creates a hook to be added to an instance of logger.
// If you want to use a custom endpoint, you have to just
// create the desired structure like for example:
// 	conn, _ = tls.Dial("tcp","my.hostname.example:1337", &tls.config{})
// 	hook = &LogmaticHook{conn, "<YOUR_API_KEY>", "my.hostname.example:1337", 10, 2, sync.Mutec{}}
// and add the hook to Logrus as:
//	logrus.AddHook(hook)
func NewLogmaticHook(apiKey string) *LogmaticHook {

	// connect to this socket
	conn, _ := tls.Dial(protocol, endpoint, &tls.Config{})
	return &LogmaticHook{conn, apiKey, endpoint, maxRetries, sleepTime, sync.Mutex{}}
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

	for i := 0; i < hook.maxRetries; i++ {

		// sleep between 2 attempts
		if (i > 0) {
			time.Sleep(sleepTime * time.Second)
		}

		if hook.conn == nil {

			hook.mu.Lock()

			// reconnect
			conn, err := tls.Dial(protocol, hook.endpoint, &tls.Config{})
			hook.conn = conn

			hook.mu.Unlock();
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to connect, error: %v\n", err)
				hook.conn = nil
				continue
			}

		}

		n, err := hook.conn.Write(b)

		if err == nil {
			return n, err
		} else {
			fmt.Fprintf(os.Stderr, "Unable to send log line. Wrote %d bytes before error: %v\n", n, err)
			fmt.Fprintf(os.Stderr, "Making a new attempt\n")
			hook.conn = nil

		}
	}

	return 0, errors.New("Failed to connect to Logmatic.io")
}
