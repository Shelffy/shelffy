package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	SubjBooksBase  = "books"
	SubjDeleteBook = SubjBooksBase + ".delete"
)

type EventType int

const (
	EventTypeDeleteBook EventType = iota
)

type DeleteBookEvent struct {
	Path string `json:"path"`
}

func (e *DeleteBookEvent) ToJSON() []byte {
	d, _ := json.Marshal(*e)
	return d
}

type BooksEventsPublisher interface {
	PublishDeleteBookEvent(ctx context.Context, storagePath string) error
}

type natsBooksEventPublisher struct {
	js jetstream.JetStream
}

func NewNATSBooksEventPublisher(js jetstream.JetStream) BooksEventsPublisher {
	return &natsBooksEventPublisher{js: js}
}

func (ep *natsBooksEventPublisher) PublishDeleteBookEvent(ctx context.Context, storagePath string) error {
	event := DeleteBookEvent{Path: storagePath}
	ack, err := ep.js.Publish(ctx, SubjDeleteBook, event.ToJSON())
	if err != nil {
		return fmt.Errorf("%s, ack=%v", err.Error(), ack)
	}
	return nil
}
