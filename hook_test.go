package logruskrp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type Payload struct {
	Records []Record
}

type Record struct {
	Value map[string]string
}

func Test_fire_hook(t *testing.T) {
	var payload Payload
	var field, message string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &payload)
		if err != nil {
			t.Error(err)
		}

		if len(payload.Records) != 1 {
			t.Fatalf("Expected a single record but got %d", len(payload.Records))
		}

		field = payload.Records[0].Value["field"]
		message = payload.Records[0].Value["msg"]

		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	hook, err := NewHook(server.URL, "hello-logs", &Options{})
	assert.NoError(t, err)

	err = hook.Fire(&logrus.Entry{
		Message: "hello world",
		Data:    logrus.Fields{"field": "foobar"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "foobar", field)
	assert.Equal(t, "hello world", message)
}
