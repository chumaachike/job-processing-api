package queue

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/chumaachike/job-processing-api/internal/job"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRabbitMQIntegration(t *testing.T) {
	url := os.Getenv("TEST_RABBITMQ_URL")
	if url == "" {
		t.Skip("set TEST_RABBITMQ_URL to run RabbitMQ integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	q, err := NewRabbitMQ("test-"+uuid.NewString(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	}()
	defer func() {
		if _, err := q.channel.QueueDelete(q.queueName, false, false, false); err != nil {
			t.Errorf("delete test queue: %v", err)
		}
	}()
	deliveries, err := q.Consume(ctx)
	if err != nil {
		t.Fatal(err)
	}
	message := job.JobMessage{JobId: uuid.New(), Type: "email"}
	if err = q.Publish(ctx, message); err != nil {
		t.Fatal(err)
	}
	select {
	case d := <-deliveries:
		var got job.JobMessage
		if err := json.Unmarshal(d.Body, &got); err != nil {
			t.Fatal(err)
		}
		if got != message || d.ContentType != "application/json" || d.DeliveryMode != amqp.Persistent {
			t.Fatalf("delivery = %+v", d)
		}
		if err := d.Ack(false); err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("delivery timed out")
	}
}
