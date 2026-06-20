package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type OrderRequest struct {
	CustomerID string `json:"customer_id"`
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
}

var (
	successCount atomic.Int64
	errorCount   atomic.Int64
	client       *http.Client
	targetURL    string
)

func main() {
	targetURL = getEnv("ORDER_SERVICE_URL", "http://order-service:8080") + "/api/orders"
	workers := getEnvInt("WORKERS", 5)
	rps := getEnvInt("RPS", 10)
	durationSec := getEnvInt("DURATION_SECONDS", 0)

	tracer.Start(
		tracer.WithService(getEnv("DD_SERVICE", "load-generator")),
		tracer.WithEnv(getEnv("DD_ENV", "hackathon")),
		tracer.WithServiceVersion(getEnv("DD_VERSION", "1.0.0")),
	)
	defer tracer.Stop()

	client = httptrace.WrapClient(http.DefaultClient)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	stopCh := make(chan struct{})

	if durationSec > 0 {
		go func() {
			time.Sleep(time.Duration(durationSec) * time.Second)
			log.Printf("[loadgen] duration reached (%ds), stopping...", durationSec)
			close(stopCh)
		}()
	}

	go func() {
		<-sigCh
		log.Println("[loadgen] signal received, stopping...")
		close(stopCh)
	}()

	interval := time.Second / time.Duration(rps)
	log.Printf("[loadgen] starting: workers=%d rps=%d target=%s", workers, rps, targetURL)

	sem := make(chan struct{}, workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			ticker := time.NewTicker(interval * time.Duration(workers))
			defer ticker.Stop()
			for {
				select {
				case <-stopCh:
					return
				case <-ticker.C:
					sem <- struct{}{}
					sendOrder(id)
					<-sem
				}
			}
		}(i)
	}

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(10 * time.Second):
				log.Printf("[loadgen] stats: success=%d errors=%d",
					successCount.Load(), errorCount.Load())
			}
		}
	}()

	<-stopCh
	time.Sleep(2 * time.Second)
	log.Printf("[loadgen] final stats: success=%d errors=%d",
		successCount.Load(), errorCount.Load())
}

func sendOrder(workerID int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	req := OrderRequest{
		CustomerID: fmt.Sprintf("cust-%d", r.Intn(999)+1),
		ProductID:  fmt.Sprintf("SKU-%03d", r.Intn(5)+1),
		Quantity:   r.Intn(5) + 1,
	}

	body, _ := json.Marshal(req)
	resp, err := client.Post(targetURL, "application/json", bytes.NewReader(body))
	if err != nil {
		errorCount.Add(1)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		successCount.Add(1)
	} else {
		errorCount.Add(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
