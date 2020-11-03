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
		var timeout SlowRequest

		err := json.NewDecoder(r.Body).Decode(&timeout)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if timeout.Timeout > tooLongTimeout {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "timeout too long"}`))
			return
		}
		ctx := context.WithValue(r.Context(), "timeout", timeout.Timeout)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func timeoutResposne(ch chan<- string, timeout uint64) {
	time.Sleep(time.Millisecond * time.Duration(int64(timeout)))
	ch <- "Successful result"
}

func apiSlow(w http.ResponseWriter, r *http.Request) {
	timeout := r.Context().Value("timeout")
	ch := make(chan string)

	go timeoutResposne(ch, timeout.(uint64))

	select {
	case <-ch:
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}

	close(ch)
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
