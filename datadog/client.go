package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	API_HOST          = "https://app.datadoghq.com"
	METRIC_UPDATE_URL = API_HOST + "/api/v1/series"
	JSON_MIME_TYPE    = "application/json"
)

type Client struct {
	APIKey         string
	ApplicationKey string
	authSuffix     string
}

func NewClient(apiKey, applicationKey string) *Client {
	c := &Client{
		APIKey:         apiKey,
		ApplicationKey: applicationKey,
	}
	v := make(url.Values)
	v.Add("api_key", apiKey)
	if applicationKey != "" {
		v.Add("application_key", applicationKey)
	}
	c.authSuffix = "?" + v.Encode()
	return c
}

func (c *Client) SendMetricUpdate(name string, value float64, timestamp int64, isCounter bool, host string, device string, tags []string) error {
	var pa PointArray
	var typ string
	if isCounter {
		pa = PointArray{
			IsCounter:    isCounter,
			CountContent: []uint64{uint64(value)},
			Timestamps:   []int64{timestamp},
		}
		typ = "counter"
	} else {
		pa = PointArray{
			IsCounter:    isCounter,
			GaugeContent: []float64{value},
			Timestamps:   []int64{timestamp},
		}
		typ = "gauge"
	}

	series := Series{
		Metric: name,
		Points: pa,
		Type:   typ,
		Host:   host,
		Device: device,
		Tags:   tags,
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	io.WriteString(buf, `{"series": [`)
	enc := json.NewEncoder(buf)
	enc.Encode(&series)
	io.WriteString(buf, "]}")
	println("updating", buf.String())

	url := METRIC_UPDATE_URL + c.authSuffix
	r, err := http.Post(url, JSON_MIME_TYPE, buf)
	if err != nil {
		return err
	}

	r.Write(os.Stderr)
	if r.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%s: unexpected response code", r.Status)
	}
	return nil
}
