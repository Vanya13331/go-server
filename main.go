package main

import (
	"context"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"time"
)

const (
	address        = "localhost:8000"
	tooLongTimeout = 5000
)

type SlowRequest struct {
	Timeout uint64
}

func middlewareCheckRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "" {
			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			if mt != "application/json" {
				http.NotFound(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func middlewareCheckTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		background := r.Context()
		cCtx, cCancel := context.WithCancel(background)
		r.WithContext(cCtx)

		go func() {
			next.ServeHTTP(w, r)
			cCancel()
		}()

		select {
		case <-time.After(time.Millisecond * time.Duration(int64(tooLongTimeout))):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "timeout too long"}`))
		case <-cCtx.Done():
			return
		}
	})
}

func timeoutResposne(ch chan<- uint8, timeout uint64) {
	select {
	case <-time.After(time.Millisecond * time.Duration(int64(timeout))):
		ch <- 1
	}
}

func apiSlow(w http.ResponseWriter, r *http.Request) {
	var timeout SlowRequest
	err := json.NewDecoder(r.Body).Decode(&timeout)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ch := make(chan uint8)

	go timeoutResposne(ch, timeout.Timeout)

	select {
	case <-ch:
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}
}

func handlers() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/slow", middlewareCheckRequest(middlewareCheckTimeout(http.HandlerFunc(apiSlow))))
	return mux
}

func main() {
	log.Printf("Startin service on %s", address)
	log.Fatalf("Couldn't start service, error: %v", http.ListenAndServe(address, handlers()))
}
