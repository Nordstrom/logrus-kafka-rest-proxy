package logruskrp

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

const defaultTimeFormat string = "2006-01-02T15:04:05.999Z07:00"

type formatter struct {
	additionalFields map[string]string
	fieldNameMap     map[string]string
}

func newFormatter(additionalFields, fieldNameMap map[string]string) *formatter {
	return &formatter{
		additionalFields: additionalFields,
		fieldNameMap:     fieldNameMap,
	}
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	record := map[string]interface{}{}
	for key, value := range f.additionalFields {
		name := f.fieldName(key)
		record[name] = value
	}

	for key, value := range entry.Data {
		name := f.fieldName(key)

		switch value := value.(type) {
		case error:
			record[name] = value.Error()
		default:
			record[name] = value
		}
	}

	record[f.fieldName("msg")] = entry.Message
	record[f.fieldName("time")] = entry.Time.Format(defaultTimeFormat)
	record[f.fieldName("level")] = entry.Level.String()

	return json.Marshal(record)
}

func (f *formatter) fieldName(name string) string {
	if desiredName, ok := f.fieldNameMap[name]; ok {
		return desiredName
	}

	return name
}
