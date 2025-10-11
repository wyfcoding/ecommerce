package cmshandler

import (
	"fmt"
	"log"
	"net/http"

	"ecommerce/internal/cms/service"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(cmsService *service.CmsService, port string) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fw, _ := fmt.Fprint(w, "OK")
	})

	portStr := fmt.Sprintf(":%s", port)
	log.Printf("Starting HTTP server on port %s", portStr)

	if err := http.ListenAndServe(portStr, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
