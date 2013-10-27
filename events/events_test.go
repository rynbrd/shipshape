package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestEventValidate(t *testing.T) {
	successes := []Event {
		Event{"test", Running},
		Event{"test", Stopped},
	}

	failures := []Event {
		Event{},
		Event{"test", ""},
		Event{"", Running},
	}

	for _, event := range successes {
		if err := event.Validate(); err != nil {
			t.Errorf(`%v.Validate() => error{"%v"}, want nil`, event, err)
		}
	}

	for _, event := range failures {
		if err := event.Validate(); err == nil {
			t.Errorf(`%v.Validate() => nil, want error`, event)
		}
	}
}

func serve(ch chan *Event, port int) (url string) {
	host := "127.0.0.1"
	path := fmt.Sprintf("/event%v", port)
	addr := fmt.Sprintf("%v:%v", host, port)
	url = fmt.Sprintf("http://%v%v", addr, path)
	handler := MakeHandler(ch)

	go func() {
		http.HandleFunc(path, handler)
		http.ListenAndServe(addr, nil)
	}()

	return
}

func TestServer(t *testing.T) {
	ch := make(chan *Event, 10)
	url := serve(ch, 4090)

	decodeResult := func(res *http.Response) (result *EventResult, err error) {
		defer func() {
			res.Body.Close()
		}()

		decoder := json.NewDecoder(res.Body)
		result = new(EventResult)
		err = decoder.Decode(result)
		return
	}

	testSend := func(jsonSuccess bool, resultSuccess bool, post[]byte, event *Event) {
		var response *http.Response
		var err error

		if post == nil {
			response, err = http.Get(url)
			if err != nil {
				t.Errorf(`Get() -> error{"%v"}, want nil`, err)
			}
			return
		} else {
			buf := bytes.NewReader(post)
			response, err = http.Post(url, "text/json", buf)
			if err != nil {
				t.Errorf(`Post(%s) -> error{"%v"}, want nil`, post, err)
			}
		}

		result, err := decodeResult(response)
		if err != nil {
			t.Errorf(`Post(%s) -> error{"%v"}, want JSON decode success`, post, err)
		}

		if jsonSuccess {
			switch {
			case resultSuccess && result.Status != Success:
				t.Errorf(`Post(%s) -> %s, want Success`, post, result.Status)
			case !resultSuccess && result.Status != Failure:
				t.Errorf(`Post(%s) -> %s, want Failure`, post, result.Status)
			}
		} else {
			if result.Status != Failure {
				t.Errorf(`Post(%s) -> result Success, want Failure`)
			}
		}

		if event != nil {
			recvEvent := <- ch
			if *recvEvent != *event {
				t.Errorf(`Handler(%s) -> %s, want %s`, recvEvent, event)
			}
		}
	}

	event := Event{"test", Running}
	data, _ := json.Marshal(event)
	testSend(true, true, data, &event)

	event = Event{"test", "invalid"}
	data, _ = json.Marshal(event)
	testSend(true, false, data, nil)

	data = []byte("this is not json")
	testSend(false, false, data, nil)

	testSend(false, false, nil, nil)
}

func TestClient(t *testing.T) {
	ch := make(chan *Event, 10)
	url := serve(ch, 4091)

	successes := []Event {
		Event{"test", Running},
		Event{"test", Stopped},
	}

	failures := []Event {
		Event{"test", "invalid"},
	}

	client := NewClient(url)
	for _, event := range successes {
		err := client.Send(&event)
		if err != nil {
			t.Errorf(`Send(%s) -> error{"%s"}, want nil`, event, err)
		}
	}

	for _, event := range failures {
		err := client.Send(&event)
		if err == nil {
			t.Errorf(`Send(%s) -> nil, want error`, event)
		}
	}

	client = NewClient(url + "/invalid")
	err := client.Send(&successes[0])
	if err == nil {
		t.Errorf(`Send(%s) -> nil, want error`, successes[0])
	}
}
