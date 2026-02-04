package main

import (
	"context"
	"log"

	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxMsg struct {
	ID      uuid.UUID
	Topic   string
	Message []byte
}

func processOutboxMessages(ctx context.Context, pool *pgxpool.Pool, pubsubClient *pubsub.Client) error {
	tx, err := pool.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return err
	}

	// lock the next pending message so other dispatcher instances don't grab it.
	rows, err := tx.Query(ctx,
	`SELECT id, topic, message
	FROM outbox
	WHERE state = 'pending'
	ORDER BY created_at
	LIMIT 1
	FOR UPDATE SKIP LOCKED`)

	if err != nil {
		return err
	}
	defer rows.Close()

	var msg OutboxMsg
	if rows.Next() {
		if err := rows.Scan(&msg.ID, &msg.Topic, &msg.Message); err != nil {
			return err
		}

		log.Printf("Publishing messages %s to topic %s", msg.ID, msg.Topic)

		result := pubsubClient.Topic(msg.Topic).Publish(ctx, &pubsub.Message{
			Data: msg.Message,
		})

		_, err := result.Get(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "UPDATE outbox SET state ='processed', processed_at = now() WHERE id = $1 ", msg.ID)
		if err != nil {
			return err
		}

		log.Printf("Marked message %s as processed", msg.ID)

		return tx.Commit(ctx)
	}
	return nil
}

func main() {
	// TODO: initialize actual Postgres and Pubsub connections
	var (
		pool         *pgxpool.Pool
		pubsubClient *pubsub.Client
	)

	// feel free to use another interval
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := processOutboxMessages(context.Background(), pool, pubsubClient); err != nil {
			log.Printf("Error processing outbox: %v", err)
		}
	}
}
