package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(t *testing.T) {
	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
		method       string
	}{
		{
			name:         "valid",
			expectedCode: http.StatusOK,
			method:       http.MethodPost,
			payload: []string{
				"http://google.com",
				"https://youtube.com",
			},
		},
		{
			name:         "method invalid",
			expectedCode: http.StatusMethodNotAllowed,
			method:       http.MethodGet,
			payload:      nil,
		},
		{
			name:         "body request invalid",
			expectedCode: http.StatusBadRequest,
			method:       http.MethodPost,
			payload:      "",
		},
		{
			name:         "limit url to 20",
			expectedCode: http.StatusBadRequest,
			method:       http.MethodPost,
			payload: []string{
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
				"http://google.com",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			buf := &bytes.Buffer{}
			json.NewEncoder(buf).Encode(tc.payload)
			req, _ := http.NewRequest(tc.method, "/", buf)

			Mux().ServeHTTP(rec, req)
			if rec.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rec.Code, tc.expectedCode)
			}
		})
	}
}
func TestLimit(t *testing.T) {
	testCases := []struct {
		name         string
		payload      interface{}
		expectedCode int
		method       string
	}{
		{
			name:         "valid",
			expectedCode: http.StatusOK,
			method:       http.MethodPost,
			payload: []string{
				"http://google.com",
				"https://youtube.com",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			buf := &bytes.Buffer{}
			json.NewEncoder(buf).Encode(tc.payload)
			req, _ := http.NewRequest(tc.method, "/", buf)
			for i := 0; 100 < maxConnectionIn; i++ {
				go func(rec *httptest.ResponseRecorder, req *http.Request) {
					Mux().ServeHTTP(rec, req)
					log.Println(rec.Code)
					if rec.Code == tc.expectedCode {
						t.Errorf("handler returned wrong status code: got %v want %v", rec.Code, tc.expectedCode)
					}
				}(rec, req)
			}
		})
	}
}

// func TestLimit() {
// 	runtime.GOMAXPROCS(runtime.NumCPU())
// 	reqChan := make(chan *http.Request)
// 	respChan := make(chan Response)
// 	start := time.Now()
// 	go dispatcher(reqChan)
// 	go workerPool(reqChan, respChan)
// 	conns, size := consumer(respChan)
// 	took := time.Since(start)
// 	ns := took.Nanoseconds()
// 	av := ns / conns
// 	average, err := time.ParseDuration(fmt.Sprintf("%d", av) + "ns")
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	fmt.Printf("Connections:\t%d\nConcurrent:\t%d\nTotal size:\t%d bytes\nTotal time:\t%s\nAverage time:\t%s\n", conns, max, size, took, average)
// }
