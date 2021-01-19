package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Do stuff BEFORE the tests!")
	exitVal := m.Run()
	log.Println("Do stuff AFTER the tests!")

	os.Exit(exitVal)
}

func TestHandler_Person(t *testing.T) {
	testCases := []struct {
		name           string
		payload        []string
		expectedCode   int
		expectedMethod string
	}{
		{
			name:           "valid",
			expectedCode:   http.StatusOK,
			expectedMethod: http.MethodPost,
			payload: []string{
				"https://google.com",
				"https://google.com",
				"https://google.com",
				"https://google.com",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/", nil)
			Router().ServeHTTP(rec, req)

			// Check the status code is what we expect.
			if rec.Code != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", rec.Code, tc.expectedCode)
			}
			// if rec.Result().Request.Method != tc.expectedMethod {
			// 	t.Errorf("handler returned wrong mthod: got %v want %v", rec.Result().Request.Method, tc.expectedMethod)
			// }
		})
	}
}
