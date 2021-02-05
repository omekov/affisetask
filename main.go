package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	timeout          = time.Second
	maxConnectionIn  = 10
	maxConnectionOut = 4
)

var atmvar int32

type result struct {
	URL            string `json:"url"`
	ResponseStatus string `json:"responseStatus"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	httpServer := &http.Server{
		Addr:        ":8080",
		Handler:     Mux(),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	<-signalChan
	log.Print("os.Interrupt - shutting down...\n")

	go func() {
		<-signalChan
		log.Fatal("os.Kill - terminating...\n")
	}()

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := httpServer.Shutdown(gracefullCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
		defer os.Exit(1)
		return
	} else {
		log.Printf("gracefully stopped\n")
	}

	cancel()

	defer os.Exit(0)
	return
}

// Mux ...
func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", limitConnections(urlHandler()))
	return mux
}

func limitConnections(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if atmvar > maxConnectionIn {
			http.Error(rw, errors.New(http.StatusText(http.StatusTooManyRequests)).Error(), http.StatusTooManyRequests)
			return
		}
		atomic.AddInt32(&atmvar, 1)
		next.ServeHTTP(rw, r)
		atomic.AddInt32(&atmvar, -1)
	})
}

func urlHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		if r.Method != http.MethodPost {
			http.Error(rw, errors.New(http.StatusText(http.StatusMethodNotAllowed)).Error(), http.StatusMethodNotAllowed)
			return
		}
		var urls []string
		if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		if len(urls) > 20 {
			http.Error(rw, errors.New("only 20 urls").Error(), http.StatusBadRequest)
			return
		}
		results, err := request(ctx, urls)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		byteResult, err := json.Marshal(results)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(byteResult)
		return
	})
}

func request(ctx context.Context, urls []string) ([]result, error) {
	client := http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        maxConnectionIn,
			MaxIdleConnsPerHost: maxConnectionIn,
		},
	}
	results := make([]result, 0)
	for _, u := range urls {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		results = append(results, result{
			URL:            u,
			ResponseStatus: resp.Status,
		})
	}
	ctx.Done()
	return results, nil
}
