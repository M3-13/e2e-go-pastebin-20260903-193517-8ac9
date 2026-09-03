package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: loggingMiddleware(os.Stdout, newRouter()),
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newRouter wires the health endpoint and the paste handler together. Every
// route is registered as an explicit method pattern so the mux carries a real
// route entry for each of them; the "/" catch-all remains as fallback so that
// unsupported methods still reach Handler.ServeHTTP for the JSON 405 (with
// Allow header) and unknown paths still get the JSON 404.
func newRouter() *http.ServeMux {
	store := NewStore()
	handler := NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /pastes", handler.createPaste)
	mux.HandleFunc("GET /pastes", handler.listPastes)
	mux.HandleFunc("GET /pastes/{id}", handler.getPaste)
	mux.HandleFunc("DELETE /pastes/{id}", handler.deletePaste)
	mux.Handle("/", handler)
	return mux
}

// healthHandler answers GET /health with 200 (skeleton liveness).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// statusRecorder captures the status code written to a ResponseWriter.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// loggingMiddleware writes one line per request to out: method, path, status and
// response duration. It never logs request bodies or paste contents.
func loggingMiddleware(out interface{ Write([]byte) (int, error) }, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		fmt.Fprintf(out, "%s %s %d %s\n", r.Method, r.URL.Path, status, time.Since(start))
	})
}
