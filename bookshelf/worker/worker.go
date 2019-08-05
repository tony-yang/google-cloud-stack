// Worker demonstrates the use of the Cloud Pub/Sub API
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"

	"github.com/tony-yang/google-cloud-stack/bookshelf"
)

var (
	countMu      sync.Mutex
	count        int
	subscription *pubsub.Subscription
)

const subName = "book-worker-sub"

// update retrieves book info and updates the database with details.
// This is a mocked function to simulate actual service call for demo only.
func update(bookID int64) error {
	book, err := bookshelf.DB.GetBook(bookID)
	if err != nil {
		return err
	}
	book.Title = "Updated " + book.Title
	return bookshelf.DB.UpdateBook(book)
}

func subscribe() {
	ctx := context.Background()
	err := subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var id int64
		if err := json.Unmarshal(msg.Data, &id); err != nil {
			log.Printf("could not decode message data: %#v", msg)
			msg.Ack()
			return
		}
		log.Printf("[ID %d] Processing.", id)
		if err := update(id); err != nil {
			log.Printf("[ID %d] could not update: %v", id, err)
			msg.Nack()
			return
		}
		countMu.Lock()
		count++
		countMu.Unlock()
		msg.Ack()
		log.Printf("[ID %d] ACK", id)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.Background()
	if bookshelf.PubsubClient == nil {
		log.Fatal("Configure the Pub/Sub client")
	}

	// Create pubsub topic if it does not yet exist.
	topic := bookshelf.PubsubClient.Topic(bookshelf.PubsubTopicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Error checking for topic: %v", err)
	}
	if !exists {
		if _, err := bookshelf.PubsubClient.CreateTopic(ctx, bookshelf.PubsubTopicID); err != nil {
			log.Fatalf("Failed to create topic: %v", err)
		}
	}

	// Create topic subscription if it doesn't yet exist
	subscription = bookshelf.PubsubClient.Subscription(subName)
	exists, err = subscription.Exists(ctx)
	if err != nil {
		log.Fatalf("Error checking for subscription: %v", err)
	}
	if !exists {
		if _, err = bookshelf.PubsubClient.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{Topic: topic}); err != nil {
			log.Fatalf("Failed to create subscription: %v", err)
		}
	}

	// Start worker goroutine
	go subscribe()

	// Publish a count of processed request to the server homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		countMu.Lock()
		defer countMu.Unlock()
		fmt.Fprintf(w, "This worker has processed %d books.", count)
	})

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
