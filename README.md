logrus-kafka-rest-proxy
-----------------------

A [Kafka REST Proxy](https://www.confluent.io/blog/a-comprehensive-open-source-rest-proxy-for-kafka) hook for [Logrus](https://github.com/sirupsen/logrus)

## Example

A simple, synchronous example

```
import (
	"github.com/Nordstrom/logrus-kafka-rest-proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	hook, _ := logruskrp.NewHook("http://localhost:8082", "hello-logs", &logruskrp.Options{})

	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.Info("Hello world")
}
```

An asynchronous example with all the fixings

```
import (
	"github.com/Nordstrom/logrus-kafka-rest-proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	hook, _ := logruskrp.NewHook("https://kafka.example.com", "hello-logs", &logruskrp.Options{
		Async: &logruskrp.AsyncOptions{
			QueueTimeout: 30 * time.Second,
			RecordLength: 10,
		},
		Client: &logruskrp.ClientOptions{
			Headers: map[string]string{
				"X-API-Key": "my-blue-suede-shoes",
			},
			Timeout: 10 * time.Second,
			TLSConfig: &tls.Config{
				...
			},
		},
	})

	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.Info("Hello world")
	hook.Flush()
}
```
