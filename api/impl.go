package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fgday/quotaday/pkg/quote"
)

type Server struct {
	qb *quote.QuoteBook
}

var _ ServerInterface = (*Server)(nil)

func NewServer() *Server {
	server := Server{}
	server.qb = quote.New()
	server.qb.FillExample()
	return &server
}

// GET quote serves a quotation from the available ones
func (s *Server) GetQuote(w http.ResponseWriter, r *http.Request, params GetQuoteParams) {
	log.Println(getRemoteHostInfo(r))

	var q *quote.Quotation
	var err error
	if params.Id == nil {
		q, err = s.qb.RandomQuotation()
	} else {
		q, err = s.qb.GetQuote(*params.Id)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
	}

	//mimeTypes := r.Header.Values("Accept")
	mimeTypes := r.Header.Values("Accept")

	for _, mt := range mimeTypes {
		for _, val := range strings.Split(mt, ",") {
			var err error
			switch val {
			case "text/html":
				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				err = q.WriteHTML(w)
			case "application/json", "*/*":
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				err = q.WriteJSON(w)
			default:
				log.Printf("Skipping MIME type %q", val)
				continue
			}

			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Serving MIME type %q", val)
			return
		}
	}

	// Default
	log.Print("No \"Accept\" header found")
	if err := q.WriteJSON(w); err != nil {
		log.Fatal(err)
	}
}

// POST quote adds a quote to the available ones
func (s *Server) PostQuote(w http.ResponseWriter, r *http.Request) {
	log.Print(getRemoteHostInfo(r))

	var newQuote quote.Quotation
	if err := json.NewDecoder(r.Body).Decode(&newQuote); err != nil {
		log.Printf("json decode failed: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("could not read request body"))
		return
	}

	if err := s.qb.AddQuote(newQuote); err != nil {
		log.Printf("QuoteBook Add failed: %s", err)
		w.WriteHeader(http.StatusInsufficientStorage)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Quote added:\n%q\n%q", newQuote.Quote, newQuote.Author)
	_ = newQuote.WriteJSON(w)
}

func getRemoteHostInfo(r *http.Request) string {
	// Sample Headers:
	// Accept:[*/*]
	// Accept-Encoding:[gzip, br]
	// Cdn-Loop:[cloudflare; loops=1]
	// Cf-Connecting-Ip:[1.2.3.4]
	// Cf-Ipcountry:[IT]
	// Cf-Ray:[93804dd5edd859f5-MXP]
	// Cf-Visitor:[{"scheme":"https"}]
	// User-Agent:[curl/7.88.1]
	// X-Forwarded-For:[2.3.4.5]
	// X-Forwarded-Host:[quote.example.com]
	// X-Forwarded-Port:[80]
	// X-Forwarded-Proto:[http]
	// X-Forwarded-Server:[traefik-32bfd46sce-74c3h]
	// X-Real-Ip:[2.3.4.5]]

	remoteAddr := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")

	// Proxied through Cloudflare?
	if remote := r.Header.Get("Cf-Connecting-Ip"); remote != "" {
		remoteAddr = fmt.Sprintf("%s (%s)", remote, r.Header.Get("Cf-Ipcountry"))
	} else if remote := r.Header.Get("X-Real-Ip"); remote != "" {
		remoteAddr = remote
	} else if remote := r.Header.Get("X-Forwarded-For"); remote != "" {
		remoteAddr = remote
	}

	return fmt.Sprintf("%s %q - %s %s %q", remoteAddr, userAgent, r.Method, r.Proto, r.URL.String())
}
