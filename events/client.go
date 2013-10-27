package events

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Client struct {
	Url string
}

// NewClient creates a new event client communicating with the server at the
// provided URL.
func NewClient(url string) *Client {
	client := new(Client)
	client.Url = url
	return client
}

// Send transmits an event to the event server.
func (client Client) Send(event *Event) (err error) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	buf := bytes.NewReader(data)
	res, err := http.Post(client.Url, "text/json", buf)
	if err != nil {
		return
	}

	defer func() {
		res.Body.Close()
	}()

	if res.StatusCode != 200 {
		msg := fmt.Sprintf("http %v received", res.StatusCode)
		err = errors.New(msg)
		return
	}

	decoder := json.NewDecoder(res.Body)
	result := EventResult{}
	err = decoder.Decode(&result)
	if err != nil {
		return
	}

	if result.Status != Success {
		err = errors.New(result.Message)
	}
	return
}
