package mock

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/nullify005/service-intesis/pkg/intesishome"
)

const (
	DefaultHTTPListen string = "127.0.0.1:5001"
)

//go:embed assets/validControlResponse.json
var _responsePayload []byte

type HTTPOption func(h *HTTPServer)

type HTTPServer struct {
	Listen string
}

// set the host:port to listen on
func WithHTTPListen(l string) HTTPOption {
	return func(h *HTTPServer) {
		h.Listen = DefaultHTTPListen
		if l != "" {
			h.Listen = l
		}
	}
}

func NewHTTPServer(opts ...HTTPOption) HTTPServer {
	h := HTTPServer{
		Listen: DefaultHTTPListen,
	}
	for _, o := range opts {
		o(&h)
	}
	return h
}

func (h *HTTPServer) Run() {
	log.Printf("HTTPServer listening on: %s", h.Listen)
	http.HandleFunc(intesishome.ControlEndpoint, handleEndpoint)
	log.Fatal(http.ListenAndServe(h.Listen, nil))
}

func handleEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Printf("received connection from: %s", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(_responsePayload); err != nil {
		log.Printf("unable to write: %s", err.Error())
	}
}
