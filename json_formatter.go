package logmatic

import (
	"encoding/json"
	"fmt"
        "time"
        log "github.com/Sirupsen/logrus"
)

const defaultTimestampFormat = time.RFC3339


type JSONFormatter struct {
}


func (f *JSONFormatter) Format(entry *log.Entry) ([]byte, error) {

	data := make(log.Fields, len(entry.Data)+3)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	//prefixFieldClashes(data)

	data["date"] = entry.Time.Format(defaultTimestampFormat)
	data["message"] = entry.Message
	data["level"] = entry.Level.String()

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
