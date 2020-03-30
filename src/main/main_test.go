package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
	"strings"
	"github.com/go-chi/chi"
)


func TestGetAllDescriptions(t *testing.T) {
	//preparations to test case
	var requestBodies [][]byte
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/15","interval":20}`))
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/10","interval":20}`))
	storage := newStorage()

	r := chi.NewRouter()
	r.Post("/api/fetcher", func(w http.ResponseWriter, r *http.Request) { postWebPage(storage, w, r) })

	for _, requestBody := range(requestBodies) {
		req, _ := http.NewRequest("POST", "/api/fetcher", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	//test case
	r.Get("/api/fetcher", func(w http.ResponseWriter, r *http.Request) { getAllDescriptions(storage, w, r) })

	req, err := http.NewRequest("GET", "/api/fetcher", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	status := w.Code
	expectedStatus := http.StatusOK
	if  status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
	}

	var pages []PageDescription
	pages = append(pages, PageDescription{0, "https://httpbin.org/range/15", 20})
	pages = append(pages, PageDescription{1, "https://httpbin.org/range/10", 20})
	expected, err := json.Marshal(&pages)
	if err != nil {
		t.Fatal(err)
	}
	var expectedBody string = string(expected)
	var wBody string = w.Body.String()
	if strings.Compare(wBody, expectedBody) != 0 {	
		t.Errorf("handler returned unexpected body: got %v want %v", wBody, expectedBody)
	}
}


func TestGetPageHistory(t *testing.T) {
	//preparations to test case
	storage := newStorage()

	r := chi.NewRouter()
	r.Post("/api/fetcher", func(w http.ResponseWriter, r *http.Request) { postWebPage(storage, w, r) })
	
	
	req, err := http.NewRequest("POST", "/api/fetcher", 
	  bytes.NewBuffer([]byte(`{"url":"https://httpbin.org/range/15","interval":20}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//test case 1
	r.Get("/api/fetcher/{id}/history", func(w http.ResponseWriter, r *http.Request) { getPageHistory(storage, w, r) })

	req, _ = http.NewRequest("GET", "/api/fetcher/0/history", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	status := w.Code
	expectedStatus := http.StatusOK

	if  status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
	}

	//test case 2
	req, _ = http.NewRequest("GET", "/api/fetcher/1/history", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	status = w.Code
	expectedStatus = http.StatusNotFound

	if  status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
	}

	//test case 3
	req, _ = http.NewRequest("GET", "/api/fetcher/abc/history", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	status = w.Code
	expectedStatus = http.StatusBadRequest

	if  status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
	}
}


func TestPostWebPage(t *testing.T) {
	var requestBodies [][]byte
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/15","interval":20}`))
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/15","interval":20}`))
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/10","interval":20}`))
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/15","interval":10}`))
	requestBodies = append(requestBodies, []byte(`{"url"}`))                                                //incorrect JSON
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/15","interval":'a'}`))  //incorrect JSON

	storage := newStorage()
	r := chi.NewRouter()
	r.Post("/api/fetcher", func(w http.ResponseWriter, r *http.Request) { postWebPage(storage, w, r) })

	for i, requestBody := range(requestBodies) {
		req, _ := http.NewRequest("POST", "/api/fetcher", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		status := w.Code
		expectedStatus := http.StatusOK
		var id uint32
		switch i {
			case 0: { 
				expectedStatus = http.StatusOK
				id = 0 
			}
			case 1: {
				expectedStatus = http.StatusOK
				id = 0 
			}
			case 2: {
				expectedStatus = http.StatusOK
				id = 1 		
			}
			case 3: {
				expectedStatus = http.StatusOK
				id = 0		
			}
			case 4: {
				expectedStatus = http.StatusBadRequest
			}
			case 5: {
				expectedStatus = http.StatusBadRequest
			}
		}
		if  status != expectedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
		}

		if status == http.StatusBadRequest {
			continue
		}

		expectedBody, err := json.Marshal(&Id{id})
		if err != nil {
			t.Fatal(err)
		}
		var body string = string(expectedBody)
		var wBody string = w.Body.String()
		if strings.Compare(wBody, body) != 0 {	
			t.Errorf("handler returned unexpected body: got %v want %v", wBody, body)
		}
	}
}


func TestDeleteWebPage(t *testing.T) {
	//preparations to test case
	var requestBodies [][]byte
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/15","interval":20}`))
	requestBodies = append(requestBodies, []byte(`{"url":"https://httpbin.org/range/10","interval":20}`))
	storage := newStorage()

	r := chi.NewRouter()
	r.Post("/api/fetcher", func(w http.ResponseWriter, r *http.Request) { postWebPage(storage, w, r) })
	
	for _, requestBody := range(requestBodies) {
		req, _ := http.NewRequest("POST", "/api/fetcher", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	//test case -> try deletion the same web page twice
	for i := 0; i < 2; i++ {
		r.Delete("/api/fetcher/{id}", func(w http.ResponseWriter, r *http.Request) { deleteWebPage(storage, w, r) })

		req, _ := http.NewRequest("DELETE", "/api/fetcher/1", nil)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	
		status := w.Code
		expectedStatus := http.StatusOK
		
		switch (i) {
			case 0:  expectedStatus = http.StatusOK
			case 1:  expectedStatus = http.StatusNotFound
		}

		if  status != expectedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
		}
	}
}