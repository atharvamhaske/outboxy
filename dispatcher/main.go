package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type OutboxMsg struct {
	ID      uuid.UUID
	Topic   string
	Message []byte
}

func processOutboxMessages(ctx context.Context, pool *pgxpool.Pool, redisClient *redis.Client) error {
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

	// No pending messages
	if !rows.Next() {
		return nil
	}

	if err := rows.Scan(&msg.ID, &msg.Topic, &msg.Message); err != nil {
		return err
	}

	// We must fully close the rows before issuing another query on the same
	// transaction connection, otherwise pgx reports "conn busy".
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	log.Printf("Publishing messages %s to channel %s", msg.ID, msg.Topic)

	if err := redisClient.Publish(ctx, msg.Topic, msg.Message).Err(); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE outbox SET state ='processed', processed_at = now() WHERE id = $1 ", msg.ID)
	if err != nil {
		return err
	}

	log.Printf("Marked message %s as processed", msg.ID)

	return tx.Commit(ctx)
}

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	var redisClient *redis.Client
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Invalid REDIS_URL: %v", err)
		}
		redisClient = redis.NewClient(opt)
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
	}
	defer redisClient.Close()

	// feel free to use another interval
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := processOutboxMessages(ctx, pool, redisClient); err != nil {
			log.Printf("Error processing outbox: %v", err)
		}
	}
}
