logrus-kafka-rest-proxy
-----------------------

A [Kafka REST Proxy](https://www.confluent.io/blog/a-comprehensive-open-source-rest-proxy-for-kafka) hook for [Logrus](https://github.com/sirupsen/logrus)

## Example

A simple, synchronous example

```go
import (
	"github.com/Nordstrom/logrus-kafka-rest-proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	hook, _ := logrest.NewHook("http://localhost:8082", "hello-logs", &logrest.Options{})

	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.Info("Hello world")
}
```

An asynchronous example with all the fixings

```go
import (
	"github.com/Nordstrom/logrus-kafka-rest-proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	hook, _ := logrest.NewHook("https://kafka.example.com", "hello-logs", &logrest.Options{
		Async: &logrest.AsyncOptions{
			QueueTimeout: 30 * time.Second,
			RecordLength: 10,
		},
		Client: &logrest.ClientOptions{
			Headers: map[string]string{
				"X-API-Key": "my-blue-suede-shoes",
			},
			Timeout: 10 * time.Second,
			TLSConfig: &tls.Config{
				...
			},
		},
		Fields: &logrest.FieldOptions{
			FieldRenameMap: map[string]string{
				"msg":  "message",
				"time": "timestamp",
			},
			AdditionalFields: map[string]string{
				"env":  "test",
				"host": os.Getenv("HOSTNAME"),
			},
		},
	})

	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.Info("Hello world")
	hook.Flush()
}
```
