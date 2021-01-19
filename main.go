package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", Router()))
}

// Router ...
func Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerURL)
	return mux
}

func handlerURL(w http.ResponseWriter, r *http.Request) {
	var urls []string
	// receptionMethods := []string{http.MethodOptions, http.MethodPost}
	// for _, m := range receptionMethods {
	// 	if r.Method != m {
	// 		http.Error(w, "", http.StatusMethodNotAllowed)
	// 		return
	// 	}
	// }
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(urls) > 20 {
		http.Error(w, "the number of url is more than 20", http.StatusBadRequest)
		return
	}
	for _, uri := range urls {
		u, err := url.Parse(uri)
		if err != nil {
			http.Error(w, fmt.Sprintf("parse url err %s", u), http.StatusBadRequest)
			return
		}
		urlRequest(*u)
	}
	fmt.Fprintf(w, "URLs: %+v", urls)
}

func urlRequest(u url.URL) error {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxConnsPerHost:     200,
	}
	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	log.Printf("inc %v", req)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New("is not status code 200")
	}
	return nil
}
