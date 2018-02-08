package logruskrp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Hook struct {
	ProxyURL   string
	KafkaTopic string
	client     *http.Client
	formatter  *logrus.JSONFormatter
	options    *Options
}

type Options struct {
	HTTPHeaders map[string]string
	CertPath    string
	CertKeyPath string
	CertCaPath  string
}

func NewHook(proxyURL, kafkaTopic string, options *Options) (*Hook, error) {
	_, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	if kafkaTopic == "" {
		return nil, errors.New("A Kafka topic must be specified")
	}

	client, err := newHttpClient(options)
	if err != nil {
		return nil, err
	}

	if options == nil {
		options = &Options{
			HTTPHeaders: map[string]string{},
		}
	}

	return &Hook{
		ProxyURL:   proxyURL,
		KafkaTopic: kafkaTopic,
		client:     client,
		formatter:  &logrus.JSONFormatter{},
		options:    options,
	}, nil
}

func (h *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	payload, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	records := fmt.Sprintf(`{"records":[{"value":%s}]}`, string(payload))
	request, err := http.NewRequest("POST", h.ProxyURL+"/topics/"+h.KafkaTopic, bytes.NewReader([]byte(records)))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/vnd.kafka.json.v1+json")
	for name, value := range h.options.HTTPHeaders {
		request.Header.Set(name, value)
	}

	response, err := h.client.Do(request)
	if err != nil {
		return nil
	}

	if response.StatusCode != http.StatusOK {
		//respBuff := new(bytes.Buffer)
		//respBuff.ReadFrom(response.Body)

		return errors.New(fmt.Sprintf("Unable to deliver payload to Kafka REST Proxy on topic '%s' HTTP:%d", h.KafkaTopic, response.StatusCode))
	}

	return nil
}

func newHttpClient(options *Options) (*http.Client, error) {
	tlsConfig := tls.Config{}
	if options.CertCaPath != "" {
		caCert, err := ioutil.ReadFile(options.CertCaPath)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	if options.CertPath != "" && options.CertKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(options.CertPath, options.CertKeyPath)
		if err != nil {
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tlsConfig},
	}, nil
}
