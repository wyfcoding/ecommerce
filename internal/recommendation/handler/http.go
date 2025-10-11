package recommendationhandler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"ecommerce/internal/recommendation/service"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(recommendationService *service.RecommendationService, port string) {
	http.HandleFunc("/v1/recommendation", recommendationService.HandleGetRecommendation)
	http.HandleFunc("/healthz", healthCheckHandler)   // Health check endpoint
	http.HandleFunc("/readyz", readinessCheckHandler) // Readiness check endpoint

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Recommendation Service listening on %s\n", addr)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", addr, err)
	}
}

// healthCheckHandler simple health check
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// readinessCheckHandler simple readiness check
func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Add more complex logic here, e.g., check database connection, dependent services, etc.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ready")
}
