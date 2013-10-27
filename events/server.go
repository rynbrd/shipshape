package events

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func MakeHandler(ch chan *Event) (func(http.ResponseWriter, *http.Request)) {
	handler := func(res http.ResponseWriter, req *http.Request) {
		encoder := json.NewEncoder(res)
		decoder := json.NewDecoder(req.Body)

		if req.Method != "POST" {
			msg := fmt.Sprintf("invalid method %s", req.Method)
			encoder.Encode(EventResult{Failure, msg})
			return
		}

		event := new(Event)
		if err := decoder.Decode(event); err != nil {
			encoder.Encode(EventResult{Failure, err.Error()})
			return
		}

		if err := event.Validate(); err != nil {
			encoder.Encode(EventResult{Failure, err.Error()})
			return
		}

		ch <- event
		encoder.Encode(EventResult{Success, "event processed"})
	}
	return handler
}
