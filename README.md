logrus-kafka-rest-proxy
-----------------------

A [Kafka REST Proxy](https://www.confluent.io/blog/a-comprehensive-open-source-rest-proxy-for-kafka) hook for [Logrus](https://github.com/sirupsen/logrus)

## Example

```
import (
	"fmt"

	"github.com/Nordstrom/logrus-kafka-rest-proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	hook, _ := logkrp.NewHook("http://localhost:8082", "hello-logs", &logkrp.Options{})

	logger := logrus.New()
	logger.Hooks.Add(hook)

	logger.Info("Hello world")
}
```
