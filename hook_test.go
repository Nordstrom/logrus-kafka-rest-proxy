package logrest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func Test_async_flush_on_record_length(t *testing.T) {
	batchCh := make(chan TestBatch, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch TestBatch
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &batch)
		if err != nil {
			t.Error(err)
		}

		batchCh <- batch
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	hook, err := NewHook(server.URL, "hello-logs", &Options{
		Async: &AsyncOptions{RecordLength: 2, QueueTimeout: 10 * time.Second},
	})
	assert.NoError(t, err)

	hook.Fire(&logrus.Entry{
		Message: "message 1",
	})
	hook.Fire(&logrus.Entry{
		Message: "message 2",
	})
	hook.Fire(&logrus.Entry{
		Message: "message 3",
	})

	select {
	case batch := <-batchCh:
		if len(batch.Records) != 2 {
			t.Fatalf("Expected two records but got %d", len(batch.Records))
		}
		assert.Equal(t, "message 1", batch.Records[0].Value["msg"])
		assert.Equal(t, "message 2", batch.Records[1].Value["msg"])

	case <-time.After(500 * time.Millisecond):
		t.Error("Timed-out without waiting for flushed batches")
	}
}

func Test_async_flush_on_queue_timeout(t *testing.T) {
	batchCh := make(chan TestBatch, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch TestBatch
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &batch)
		if err != nil {
			t.Error(err)
		}

		batchCh <- batch
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	hook, err := NewHook(server.URL, "hello-logs", &Options{
		Async: &AsyncOptions{RecordLength: 200, QueueTimeout: 10 * time.Millisecond},
	})
	assert.NoError(t, err)

	hook.Fire(&logrus.Entry{
		Message: "message 1",
	})
	hook.Fire(&logrus.Entry{
		Message: "message 2",
	})
	time.Sleep(20 * time.Millisecond)
	hook.Fire(&logrus.Entry{
		Message: "message 3",
	})

	select {
	case batch := <-batchCh:
		if len(batch.Records) != 2 {
			t.Fatalf("Expected two records but got %d", len(batch.Records))
		}
		assert.Equal(t, "message 1", batch.Records[0].Value["msg"])
		assert.Equal(t, "message 2", batch.Records[1].Value["msg"])

	case <-time.After(500 * time.Millisecond):
		t.Error("Timed-out without waiting for flushed batches")
	}
}
