package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	errOnlyMethodsPOSTAndOPTIONS = errors.New("Sorry, only POST and OPTIONS methods are supported")
	errLimitedURLs               = errors.New("limited urls")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	httpServer := &http.Server{
		Addr:        ":8080",
		Handler:     Router(),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
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

// Router ...
func Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerURL)
	return mux
}

func handlerURL(w http.ResponseWriter, r *http.Request) {
	var urls []string
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
	default:
		http.Error(w, errOnlyMethodsPOSTAndOPTIONS.Error(), http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(urls) > 20 {
		http.Error(w, errLimitedURLs.Error(), http.StatusBadRequest)
		return
	}
	for _, uri := range urls {
		u, err := url.Parse(uri)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := urlRequest(*u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func urlRequest(u url.URL) error {
	transport := &http.Transport{
		MaxConnsPerHost: 100,
		MaxIdleConns:    4,
	}
	client := http.Client{
		Timeout:   1 * time.Second,
		Transport: transport,
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	log.Printf("resposne %v", res)
	if res.StatusCode != http.StatusOK {
		return errors.New("is not status code 200")
	}
	return nil
}
