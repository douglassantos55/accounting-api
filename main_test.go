package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func Post(url string, data string) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	return req, err
}

func TestHelloWorld(t *testing.T) {
	router := NewRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %v", w.Code)
	}

	if w.Body.String() != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %v", w.Body.String())
	}
}

func TestCreateAccount(t *testing.T) {
	router := NewRouter()

	w := httptest.NewRecorder()
	req, _ := Post("/accounts", "name=Foo")

	router.ServeHTTP(w, req)
	contentType := w.Result().Header.Get("content-type")

	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON response, got %v", contentType)
	}

    var responseJSON gin.H
    json.Unmarshal(w.Body.Bytes(), &responseJSON)

	if responseJSON["name"] != "Foo" {
		t.Errorf("Expected Bar, got %v", responseJSON)
	}
}
