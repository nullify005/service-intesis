package mock

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nullify005/service-intesis/pkg/intesishome/cloud"
)

const (
	DefaultHTTPListen string = "127.0.0.1:5001"
	healthPath        string = "/health"
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
	router := gin.Default()
	router.POST(cloud.Endpoint, handleEndpoint)
	router.GET(healthPath, handleHealth)
	log.Fatal(router.Run(h.Listen))
}

func handleEndpoint(c *gin.Context) {
	c.String(http.StatusOK, string(_responsePayload))
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
