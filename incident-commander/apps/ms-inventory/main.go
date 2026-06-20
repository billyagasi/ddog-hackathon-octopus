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

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/config"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/handler"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/infra"
	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/model"
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

	reader := infra.NewKafkaReaderOrderCreated(cfg)
	defer reader.Close()

	writer := infra.NewKafkaWriterProducer(cfg)
	defer writer.Close()

	h := handler.NewInventoryHandler(cfg, db, rdb, writer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[main] inventory-service started, consuming order-created...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, rdrErr := reader.FetchMessage(ctx)
			if rdrErr != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[main] fetch error: %v", rdrErr)
				time.Sleep(time.Second)
				continue
			}

			var event model.OrderCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[main] unmarshal error: %v", err)
				reader.CommitMessages(ctx, msg)
				continue
			}

			if err := h.ProcessOrder(ctx, event); err != nil {
				log.Printf("[main] inventory processing error: %v", err)
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
