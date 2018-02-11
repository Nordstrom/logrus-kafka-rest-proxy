package logrest

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_formatter_additional_fields(t *testing.T) {
	formatter := newFormatter(map[string]string{
		"hello": "world",
	}, map[string]string{})

	entry := newEntry()
	serialized, err := formatter.Format(entry)
	assert.NoError(t, err)

	var deserialized map[string]string
	err = json.Unmarshal(serialized, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, "world", deserialized["hello"])
}

func Test_formatter_field_name_map(t *testing.T) {
	formatter := newFormatter(map[string]string{},
		map[string]string{
			"hello": "buongiorno",
		})

	entry := newEntry()
	serialized, err := formatter.Format(
		entry.WithFields(logrus.Fields{"hello": "world"}))
	assert.NoError(t, err)

	var deserialized map[string]string
	err = json.Unmarshal(serialized, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, "world", deserialized["buongiorno"])
}

func newEntry() *logrus.Entry {
	logger := logrus.New()
	logger.Out = &bytes.Buffer{}
	return logrus.NewEntry(logger)
}
