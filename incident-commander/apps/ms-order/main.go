package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/config"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/handler"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/infra"
)

func main() {
	cfg := config.Load()

	tracer.Start(
		tracer.WithService(cfg.DDService),
		tracer.WithEnv(cfg.DDEnv),
		tracer.WithServiceVersion(cfg.DDVersion),
	)
	defer tracer.Stop()

	db, err := infra.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer db.Close()

	rdb, err := infra.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	defer rdb.Close()

	infra.EnsureTopics(cfg)

	orderWriter := infra.NewKafkaWriterProducer(cfg)
	defer orderWriter.Close()

	paymentReader := infra.NewKafkaReaderPaymentCompleted(cfg)
	defer paymentReader.Close()

	orderHandler := handler.NewOrderHandler(cfg, db, rdb, orderWriter)
	paymentHandler := handler.NewPaymentHandler(orderHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", orderHandler.CreateOrder)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		log.Printf("[main] order-service HTTP server on :%s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Kafka consumer for payment-completed
	go func() {
		log.Println("[main] consuming payment-completed...")
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, rdrErr := paymentReader.FetchMessage(ctx)
			if rdrErr != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[main] fetch error: %v", rdrErr)
				time.Sleep(time.Second)
				continue
			}

			if err := paymentHandler.HandlePaymentCompleted(ctx, msg.Value); err != nil {
				log.Printf("[main] payment processing error: %v", err)
			}

			paymentReader.CommitMessages(ctx, msg)
		}
	}()

	<-sigCh
	log.Println("[main] shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)

	time.Sleep(2 * time.Second)
	log.Println("[main] stopped")
}
