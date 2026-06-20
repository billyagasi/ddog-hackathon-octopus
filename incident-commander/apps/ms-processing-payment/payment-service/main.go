package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/config"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/handler"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/infra"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-processing-payment/payment-service/internal/model"
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

	es, err := infra.NewElasticsearchClient(cfg)
	if err != nil {
		log.Fatalf("elasticsearch init failed: %v", err)
	}

	infra.EnsureTopics(cfg)

	reader := infra.NewKafkaReader(cfg)
	defer reader.Close()

	writer := infra.NewKafkaWriter(cfg)
	defer writer.Close()

	h := handler.New(cfg, db, es, writer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[main] payment-service started, consuming inventory-reserved...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[main] fetch error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var event model.InventoryReservedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[main] unmarshal error: %v", err)
				reader.CommitMessages(ctx, msg)
				continue
			}

			if err := h.ProcessPayment(ctx, event); err != nil {
				log.Printf("[main] payment processing error: %v", err)
			}

			reader.CommitMessages(ctx, msg)
		}
	}()

	<-sigCh
	log.Println("[main] shutting down...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Println("[main] stopped")
}
