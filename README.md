# logmatic-go
Go helpers to send logs to Logmatic.io.

## How to use it

```go
package main

import (
  log "github.com/Sirupsen/logrus"
  "github.com/gpolaert/logmatic-go"
)

func main() {
  
	// use JSONFormatter
  	log.SetFormatter(&logmatic.JSONFormatter{})
	
	// instantiate a new Logger with your Logmatic APIKey
  	log.AddHook(logmatic.NewLogmaticHook("<YOUR_API_KEY>"))
  
	// log an event as usual with logrus
 	log.WithFields(log.Fields{"string": "foo", "int": 1, "float": 1.1 }).Info("My first event from golang")
  
}
```
