/*
Copyright © 2025 Francesco Giudici <dev@foggy.day>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fgday/quotaday/pkg/quote"
)

func TestGetQuote_Random_JSON(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest("GET", "/quote", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.GetQuote(w, req, GetQuoteParams{})

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 200 or 400, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Expected application/json content-type, got %q", ct)
	}
	var q quote.Quotation
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
			t.Errorf("Failed to decode JSON: %v", err)
		}
		if q.Quote == "" || q.Author == "" {
			t.Errorf("Expected non-empty quotation, got %+v", q)
		}
	}
}

func TestGetQuote_Random_HTML(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest("GET", "/quote", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	s.GetQuote(w, req, GetQuoteParams{})

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 200 or 400, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Expected text/html content-type, got %q", ct)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	if resp.StatusCode == http.StatusOK && !strings.Contains(body, "<html>") {
		t.Errorf("Expected HTML output, got %q", body)
	}
}

func TestGetQuote_ByID(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest("GET", "/quote?id=0", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	id := 0
	s.GetQuote(w, req, GetQuoteParams{Id: &id})

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	var q quote.Quotation
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		t.Errorf("Failed to decode JSON: %v", err)
	}
	if q.Quote == "" || q.Author == "" {
		t.Errorf("Expected non-empty quotation, got %+v", q)
	}
}

func TestGetQuote_BadID(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest("GET", "/quote?id=999", nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	id := 999
	s.GetQuote(w, req, GetQuoteParams{Id: &id})

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "out of bounds") {
		t.Errorf("Expected error message about out of bounds, got %q", string(body))
	}
}

func TestPostQuote_Success(t *testing.T) {
	s := NewServer()
	q := quote.Quotation{Quote: "Hello", Author: "Tester"}
	body, _ := json.Marshal(q)
	req := httptest.NewRequest("POST", "/quote", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.PostQuote(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}
	var got quote.Quotation
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Errorf("Failed to decode JSON: %v", err)
	}
	if got != q {
		t.Errorf("Expected %+v, got %+v", q, got)
	}
}

func TestPostQuote_BadBody(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest("POST", "/quote", strings.NewReader("not-json"))
	w := httptest.NewRecorder()
	s.PostQuote(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "could not read request body") {
		t.Errorf("Expected error message, got %q", string(body))
	}
}

func TestPostQuote_StorageFull(t *testing.T) {
	s := NewServer()
	// Fill up the quote book
	for i := 0; i <= 21; i++ {
		q := quote.Quotation{Quote: "Q", Author: "A"}
		body, _ := json.Marshal(q)
		req := httptest.NewRequest("POST", "/quote", bytes.NewReader(body))
		w := httptest.NewRecorder()
		s.PostQuote(w, req)
	}
	// The last request should fail with 507
	q := quote.Quotation{Quote: "Q", Author: "A"}
	body, _ := json.Marshal(q)
	req := httptest.NewRequest("POST", "/quote", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.PostQuote(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInsufficientStorage {
		t.Errorf("Expected 507 Insufficient Storage, got %d", resp.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(bodyBytes), "full") {
		t.Errorf("Expected error message about full, got %q", string(bodyBytes))
	}
}
