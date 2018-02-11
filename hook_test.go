package logrest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type TestBatch struct {
	Records []TestRecord
}

type TestRecord struct {
	Value map[string]string
}

func Test_fire_hook(t *testing.T) {
	var batch TestBatch
	var field, message string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &batch)
		if err != nil {
			t.Error(err)
		}

		if len(batch.Records) != 1 {
			t.Fatalf("Expected a single record but got %d", len(batch.Records))
		}

		field = batch.Records[0].Value["field"]
		message = batch.Records[0].Value["msg"]

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
