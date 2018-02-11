package logrest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

type Hook struct {
	ProxyURL   string
	KafkaTopic string
	client     *http.Client
	options    *Options
	formatter  logrus.Formatter

	entryCh chan *logrus.Entry
	flushCh chan chan bool
}

type Options struct {
	Async  *AsyncOptions
	Client *ClientOptions
	Fields *FieldOptions
}

type FieldOptions struct {
	AdditionalFields map[string]string
	FieldRenameMap   map[string]string
}

type ClientOptions struct {
	Headers   map[string]string
	Timeout   time.Duration
	TLSConfig *tls.Config
}

type AsyncOptions struct {
	// QueueTimeout specifies the length of time to wait
	// before flushing the records. The timeout counter
	// begins when there is at least one record. A value
	// of Duration(0) means that the RecordLength must
	// be reached before any records are flushed.
	//
	QueueTimeout time.Duration

	// RecordLength specifies the number of records to
	// hold onto before flushing the records. A value of
	// uint(0) means that the QueueTimeout must be reached
	// before any records are flushed.
	//
	RecordLength uint
}

const defaultQueueTimeoutSeconds = 30
const defaultRecordLength = 10

func NewHook(proxyURL, kafkaTopic string, options *Options) (*Hook, error) {
	_, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	if kafkaTopic == "" {
		return nil, errors.New("A Kafka topic must be specified")
	}

	if options == nil {
		options = &Options{}
	}

	if options.Fields == nil {
		options.Fields = &FieldOptions{
			FieldRenameMap:   map[string]string{},
			AdditionalFields: map[string]string{},
		}
	}

	if options.Client == nil {
		options.Client = &ClientOptions{
			Headers: map[string]string{},
		}
	}

	if options.Async != nil {
		if options.Async.QueueTimeout == 0 && options.Async.RecordLength == 0 {
			options.Async.QueueTimeout = defaultQueueTimeoutSeconds * time.Second
			options.Async.RecordLength = defaultRecordLength
		}
	}

	client, err := newClient(options.Client)
	if err != nil {
		return nil, err
	}

	hook := &Hook{
		ProxyURL:   proxyURL,
		KafkaTopic: kafkaTopic,
		client:     client,
		formatter:  newFormatter(options.Fields.AdditionalFields, options.Fields.FieldRenameMap),
		options:    options,
	}

	if options.Async != nil {
		hook.entryCh = make(chan *logrus.Entry)
		hook.flushCh = make(chan chan bool)
		go hook.collectEntries()
	}

	return hook, nil
}

func (h *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	if h.entryCh == nil {
		return h.flushEntry(entry)
	}

	h.entryCh <- entry
	return nil
}

func (h *Hook) Flush() {
	if h.flushCh == nil {
		return
	}

	done := make(chan bool)
	h.flushCh <- done
	<-done
}

func (h *Hook) collectEntries() {
	entries := make([]*logrus.Entry, 0)
	queueTimeout := h.options.Async.QueueTimeout
	recordLength := int(h.options.Async.RecordLength)

	timer := time.NewTimer(queueTimeout)
	var timerCh <-chan time.Time

	for {
		select {
		case entry := <-h.entryCh:
			entries = append(entries, entry)

			if len(entries) >= recordLength && recordLength > 0 {
				go h.flushEntries(entries)

				entries = make([]*logrus.Entry, 0)
				timerCh = nil
			} else if timerCh == nil && queueTimeout > 0 {
				timer.Reset(queueTimeout)
				timerCh = timer.C
			}

		case <-timerCh:
			if len(entries) > 0 {
				go h.flushEntries(entries)
				entries = make([]*logrus.Entry, 0)
			}

			timerCh = nil

		case done := <-h.flushCh:
			if len(entries) > 0 {
				h.flushEntries(entries)
				entries = make([]*logrus.Entry, 0)
			}

			timerCh = nil
			done <- true
		}
	}
}

type Batch struct {
	Records []Record `json:"records"`
}

type Record struct {
	Value json.RawMessage `json:"value"`
}

func (h *Hook) flushEntry(entry *logrus.Entry) error {
	return h.flushEntries([]*logrus.Entry{entry})
}

func (h *Hook) flushEntries(entries []*logrus.Entry) error {
	batch := Batch{
		Records: []Record{},
	}

	for _, entry := range entries {
		message, err := h.formatter.Format(entry)
		if err != nil {
			continue
		}

		batch.Records = append(batch.Records, Record{Value: json.RawMessage(message)})
	}

	buffer, err := json.Marshal(&batch)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", h.ProxyURL+"/topics/"+h.KafkaTopic, bytes.NewReader(buffer))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/vnd.kafka.json.v1+json")
	for name, value := range h.options.Client.Headers {
		request.Header.Set(name, value)
	}

	response, err := h.client.Do(request)
	if err != nil {
		return nil
	}

	if response.StatusCode != http.StatusOK {
		//respBuff := new(bytes.Buffer)
		//respBuff.ReadFrom(response.Body)

		return errors.New(fmt.Sprintf("Unable to deliver payload to Kafka REST Proxy on topic '%s' HTTP:%d",
			h.KafkaTopic, response.StatusCode))
	}

	return nil
}

func newClient(options *ClientOptions) (*http.Client, error) {
	return &http.Client{
		Timeout:   options.Timeout,
		Transport: &http.Transport{TLSClientConfig: options.TLSConfig},
	}, nil
}
