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

package quote

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
)

type errorWriter struct{}

func (errorWriter) Write(p []byte) (int, error) { return 0, io.ErrClosedPipe }

func TestNewQuoteBook(t *testing.T) {
	qb := New()
	if qb == nil {
		t.Fatal("New() returned nil")
	}
	if len(qb.quoteList) != 0 {
		t.Errorf("Expected empty quoteList, got %d", len(qb.quoteList))
	}
}

func TestAddQuoteAndGetQuote(t *testing.T) {
	qb := New()
	q := Quotation{Quote: "Hello", Author: "World"}
	if err := qb.AddQuote(q); err != nil {
		t.Fatalf("AddQuote failed: %v", err)
	}
	got, err := qb.GetQuote(0)
	if err != nil {
		t.Fatalf("GetQuote failed: %v", err)
	}
	if *got != q {
		t.Errorf("Expected %+v, got %+v", q, *got)
	}
}

func TestAddQuote_FullBook(t *testing.T) {
	qb := New()
	for i := 0; i <= maxQuotes+1; i++ {
		err := qb.AddQuote(Quotation{Quote: "Q", Author: "A"})
		if i <= maxQuotes {
			if err != nil {
				t.Fatalf("Unexpected error before full: %v", err)
			}
		} else {
			if err == nil {
				t.Error("Expected error when adding beyond maxQuotes, got nil")
			}
		}
	}
}

func TestGetQuote_Errors(t *testing.T) {
	qb := New()
	// Empty book
	_, err := qb.GetQuote(0)
	if err == nil {
		t.Error("Expected error for empty QuoteBook, got nil")
	}
	// Out-of-bounds
	_ = qb.AddQuote(Quotation{Quote: "A", Author: "B"})
	_, err = qb.GetQuote(2)
	if err == nil {
		t.Error("Expected error for out-of-bounds index, got nil")
	}
}

func TestRandomQuotation(t *testing.T) {
	qb := New()
	// Empty
	_, err := qb.RandomQuotation()
	if err == nil {
		t.Error("Expected error on empty QuoteBook, got nil")
	}
	// Normal
	qb.FillExample()
	got, err := qb.RandomQuotation()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	found := false
	for _, q := range qb.quoteList {
		if *got == q {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("RandomQuotation returned quote not in list: %+v", got)
	}
}

func TestWriteHTML(t *testing.T) {
	q := Quotation{Quote: "HTML", Author: "Tester"}
	var buf bytes.Buffer
	err := q.WriteHTML(&buf)
	if err != nil {
		t.Fatalf("WriteHTML error: %v", err)
	}
	output := buf.String()
	if output == "" || !bytes.Contains(buf.Bytes(), []byte("HTML")) {
		t.Errorf("HTML output incorrect: %s", output)
	}
}

func TestWriteHTML_ErrorWriter(t *testing.T) {
	q := Quotation{Quote: "X", Author: "Y"}
	err := q.WriteHTML(errorWriter{})
	if err == nil {
		t.Error("Expected error from WriteHTML with errorWriter, got nil")
	}
}

func TestWriteJSON(t *testing.T) {
	q := Quotation{Quote: "JSON", Author: "Tester"}
	var buf bytes.Buffer
	if err := q.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	var got Quotation
	if err := json.NewDecoder(&buf).Decode(&got); err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if got != q {
		t.Errorf("Expected %+v, got %+v", q, got)
	}
}

func TestWriteJSON_ErrorWriter(t *testing.T) {
	q := Quotation{}
	err := q.WriteJSON(errorWriter{})
	if err == nil {
		t.Error("Expected error from WriteJSON with errorWriter, got nil")
	}
}

// --- Concurrency tests ---

func TestConcurrentAddQuote(t *testing.T) {
	qb := New()
	var wg sync.WaitGroup
	errCh := make(chan error, maxQuotes+2)
	for i := 0; i < maxQuotes+2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errCh <- qb.AddQuote(Quotation{Quote: "Q", Author: "A"})
		}(i)
	}
	wg.Wait()
	close(errCh)
	countErrors := 0
	countOK := 0
	for err := range errCh {
		if err != nil {
			countErrors++
		} else {
			countOK++
		}
	}
	if countOK != maxQuotes+1 { // The last allowed add is index maxQuotes, then error on next
		t.Errorf("Expected %d successful adds, got %d", maxQuotes+1, countOK)
	}
	if countErrors == 0 {
		t.Error("Expected errors adding quotes concurrently beyond maxQuotes")
	}
}

func TestConcurrentRandomAndAdd(t *testing.T) {
	qb := New()
	qb.FillExample()
	var wg sync.WaitGroup
	var randomErrs, addErrs int32
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := qb.RandomQuotation()
			if err != nil && !errors.Is(err, nil) {
				randomErrs++
			}
		}()
		go func() {
			defer wg.Done()
			err := qb.AddQuote(Quotation{Quote: "C", Author: "D"})
			if err != nil && !errors.Is(err, nil) {
				addErrs++
			}
		}()
	}
	wg.Wait()
}
