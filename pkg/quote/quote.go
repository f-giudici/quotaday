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
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"sync"
)

// Quotation contains the data of a single quote
type Quotation struct {
	Quote  string
	Author string
}

const maxQuotes = 20

// QuoteBook is a collection of Quotations
type QuoteBook struct {
	quoteList []Quotation
	sync.Mutex
}

func New() *QuoteBook {
	return new(QuoteBook)
}

func (q *QuoteBook) RandomQuotation() (*Quotation, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.quoteList) == 0 {
		return nil, fmt.Errorf("empty QuoteBook")
	}

	idx := rand.Intn(len(q.quoteList))
	quote := q.quoteList[idx]
	return &quote, nil
}

func (q *QuoteBook) FillExample() {
	q.quoteList = []Quotation{
		{"Start before you are ready. Don't prepare, begin.", "Mel Robbins"},
		{"Eat the frog first.", "Brian Tracy"},
		{"Imperfect action beats perfect inaction.", "Harry S. Truman"},
		{"Succeed or survive (but try).", "Mel Robbins"},
		{"Be responsible for telling people the truth, not managing people's reactions to it.", "Mel Robbins"},
		{"Today's favor is tomorrow's expectation.", "Mel Robbins"},
	}
}

func (q *QuoteBook) AddQuote(quote Quotation) error {
	q.Lock()
	defer q.Unlock()
	if len(q.quoteList) > maxQuotes {
		return fmt.Errorf("QuoteBook is full")
	}
	q.quoteList = append(q.quoteList, quote)
	return nil
}

func (q *QuoteBook) GetQuote(num int) (*Quotation, error) {
	q.Lock()
	defer q.Unlock()
	if len(q.quoteList) == 0 {
		return nil, fmt.Errorf("empty QuoteBook")
	}
	if num >= len(q.quoteList) {
		return nil, fmt.Errorf("id %d out of bounds", num)
	}
	quote := q.quoteList[num]
	return &quote, nil
}

func (q *Quotation) WriteHTML(w io.Writer) error {
	const tpl = `
<!DOCTYPE html>
<html>
<body>

<q style=font-size:200%;font-family:cursive>{{.Quote}}</q>
<p><i>{{.Author}}</i></p>

</body>
</html>`

	tmpl, err := template.New("html").Parse(tpl)
	if err != nil {
		return err
	}
	tmpl.Execute(w, q)
	return nil
}

func (q *Quotation) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(q)
}
