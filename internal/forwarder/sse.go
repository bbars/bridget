package forwarder

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"sort"
	"strings"

	"github.com/bbars/bridget/internal/event"
)

const SseMimeType = "text/event-stream"

func ForwardSse(subr event.Subscriber, ch <-chan event.Event, w http.ResponseWriter, r *http.Request) error {
	var flusher http.Flusher
	if v, ok := w.(http.Flusher); !ok {
		err := fmt.Errorf("streaming is not supported")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	} else {
		flusher = v
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set(httpHeaderSubscriberId, subr.Id.String())
	w.Header().Set(httpHeaderSubscriberName, subr.Name)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	defer flusher.Flush()

	bw := bufio.NewWriter(w)
	for {
		select {
		case <-r.Context().Done():
			return nil
		case evt, ok := <-ch:
			if !ok {
				return nil
			}

			err := writeSse(bw, evt.Data)
			go evt.SendForwardResult(r.Context(), subr.Id, err)
			flusher.Flush()
			if err != nil {
				return fmt.Errorf("write event: %w", err)
			}
		}
	}
}

func writeSse(bw *bufio.Writer, evtData event.EventData) error {
	if _, err := bw.WriteString(fmt.Sprintf("event: %s\ndata: ", evtData.Path)); err != nil {
		return fmt.Errorf("write event: %w", err)
	}

	jsonData, err := json.Marshal(sseEventData(evtData))
	if err != nil {
		log.Printf("failed to marshal event data: %v", err)
		jsonData, _ = json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("marshal event data: %v", err),
		})
	}
	if _, err = bw.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write event data: %w", err)
	}

	if _, err = bw.Write([]byte{'\n', '\n'}); err != nil {
		return fmt.Errorf("failed to write event data delimiter: %w", err)
	}

	if err = bw.Flush(); err != nil {
		return fmt.Errorf("flush buffer: %w", err)
	}

	return nil
}

type sseEventData event.EventData

func (e sseEventData) MarshalJSON() ([]byte, error) {
	var kind string
	switch mediaType, _, err := mime.ParseMediaType(e.Header.Get("Content-Type")); {
	case err != nil:
		kind = "binary"
	case mediaType == "application/json":
		kind = "value"
	case strings.HasPrefix(mediaType, "text/"):
		kind = "text"
	default:
		kind = "binary"
	}

	var body any
	switch kind {
	case "value":
		body = json.RawMessage(e.Body)
	case "text":
		body = string(e.Body)
	default:
		body = base64.StdEncoding.EncodeToString(e.Body)
	}

	return json.Marshal(struct {
		Kind       string     `json:"kind,omitempty"`
		ClientAddr string     `json:"client_addr,omitempty"`
		Method     string     `json:"method,omitempty"`
		Path       string     `json:"path,omitempty"`
		Query      string     `json:"query,omitempty"`
		Header     httpHeader `json:"header,omitempty"`
		Body       any        `json:"body,omitempty"`
	}{
		Kind:       kind,
		ClientAddr: e.ClientAddr,
		Method:     e.Method,
		Path:       e.Path,
		Query:      e.Query,
		Header:     httpHeader(e.Header),
		Body:       body,
	})
}

type httpHeader http.Header

func (h httpHeader) MarshalJSON() ([]byte, error) {
	sets := make([][]string, 0, len(h))
	for k, vv := range h {
		sets = append(sets,
			append([]string{k}, vv...),
		)
	}

	sort.Slice(sets, func(i, j int) bool {
		return sets[i][0] < sets[j][0]
	})

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)

	buf.WriteByte('{')
	for i, set := range sets {
		if i > 0 {
			// Write property separator.
			buf.WriteByte(',')
		}

		// Write property name.
		_ = enc.Encode(set[0])
		buf.WriteByte(':')

		// Write value.
		if len(set) == 2 {
			_ = enc.Encode(set[1]) // single value as a JSON string
		} else {
			_ = enc.Encode(set[1:]) // multiple values as a JSON array
		}
	}
	buf.WriteByte('}')

	return buf.Bytes(), nil
}
