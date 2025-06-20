package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/f-giudici/quotaday/pkg/quote"
)

type Server struct {
	*quote.QuoteBook
}

var _ ServerInterface = (*Server)(nil)

func NewServer() *Server {
	server := Server{}
	server.QuoteBook = quote.New()
	server.FillExample()
	return &server
}

// GET quotes
func (s *Server) GetQuotes(w http.ResponseWriter, r *http.Request, params GetQuotesParams) {
	log.Println(getRemoteHostInfo(r))

	q := s.RandomQuotation()
	if err := q.WriteJSON(w); err != nil {
		log.Fatal(err)
	}
}

// POST quotes
func (s *Server) PostQuotes(w http.ResponseWriter, r *http.Request) {
	log.Println(getRemoteHostInfo(r))

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("could not read request body"))
		return
	}

	q := quote.Quotation{}
	err = json.Unmarshal(body, &q)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid quote format"))
		return
	}

	s.AddQuote(&q)
	w.WriteHeader(http.StatusCreated)
	if err := q.WriteJSON(w); err != nil {
		log.Print("error writing quotation")
		return
	}
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
