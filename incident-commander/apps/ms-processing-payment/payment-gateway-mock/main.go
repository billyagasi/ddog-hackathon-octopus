package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	tracer.Start(
		tracer.WithService(getEnv("DD_SERVICE", "payment-gateway")),
		tracer.WithEnv(getEnv("DD_ENV", "hackathon")),
		tracer.WithServiceVersion(getEnv("DD_VERSION", "1.0.0")),
	)
	defer tracer.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/payment", handlePayment)

	addr := ":8081"
	log.Printf("Payment Gateway Mock running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	status := "SUCCESS"
	if rand.Intn(10) == 0 {
		status = "FAILED"
	}

	resp := map[string]interface{}{
		"transaction_id": fmt.Sprintf("TXN-%d", rand.Intn(999999)),
		"status":         status,
		"message":        "payment processed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
