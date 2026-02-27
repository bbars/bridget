package forwarder

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bbars/bridget/internal/event"
)

const MultipartRawMimeType = "multipart/raw"

func ForwardMultipartRaw(subr event.Subscriber, ch <-chan event.Event, w http.ResponseWriter, r *http.Request) interface{} {
	var flusher http.Flusher
	if v, ok := w.(http.Flusher); !ok {
		err := fmt.Errorf("streaming is not supported")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	} else {
		flusher = v
	}

	delimiter := "################################################################"
	w.Header().Set("Content-Type", MultipartRawMimeType+"; delimiter="+delimiter)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set(httpHeaderSubscriberId, subr.Id.String())
	w.Header().Set(httpHeaderSubscriberName, subr.Name)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	defer flusher.Flush()

	bw := bufio.NewWriter(w)
	if _, err := bw.WriteString("--" + delimiter + "\n"); err != nil {
		return fmt.Errorf("write multipart delimiter: %w", err)
	}

	for {
		select {
		case <-r.Context().Done():
			return nil
		case evt, ok := <-ch:
			if !ok {
				return nil
			}

			err := writeMultipartRaw(bw, evt.Data, delimiter)
			go evt.SendForwardResult(r.Context(), subr.Id, err)
			flusher.Flush()
			if err != nil {
				return fmt.Errorf("write event: %w", err)
			}
		}
	}
}

func writeMultipartRaw(bw *bufio.Writer, evtData event.EventData, delimiter string) error {
	h := evtData.Header.Clone()
	h.Set("X-Event-Method", evtData.Method)
	h.Set("X-Event-Path", evtData.Path)
	h.Set("X-Event-Query", evtData.Query)
	h.Set("X-Event-Emitter-Address", evtData.ClientAddr)
	h.Set("X-Event-Date", time.Now().UTC().Format(http.TimeFormat))
	h.Set("Content-Length", strconv.Itoa(len(evtData.Body)))
	if err := h.Write(bw); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if len(evtData.Body) > 0 {
		if _, err := bw.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("write header delimiter: %w", err)
		}
		if _, err := bw.Write(evtData.Body); err != nil {
			return fmt.Errorf("write body: %w", err)
		}
	}

	if _, err := bw.WriteString("\n--" + delimiter + "\n"); err != nil {
		return fmt.Errorf("write multipart delimiter: %w", err)
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush buffer: %w", err)
	}

	return nil
}
