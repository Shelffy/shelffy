package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type EventsProcessor interface {
	Run(ctx context.Context) error
}

const (
	booksStreamName       = "BOOKS"
	deleteBookDurableName = "books-deleter"
	deleteBookBatch       = 100
	deleteBookMaxWait     = time.Minute
)

type natsEventProcessor struct {
	js      jetstream.JetStream
	storage FileStorage
	logger  *slog.Logger
}

func NewNATSEventProcessor(js jetstream.JetStream, storage FileStorage, logger *slog.Logger) EventsProcessor {
	return &natsEventProcessor{
		js:      js,
		storage: storage,
		logger:  logger,
	}
}

func (ep *natsEventProcessor) handleDeleteBookEvents(ctx context.Context, cons jetstream.Consumer) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msgBatch, err := cons.Fetch(
				deleteBookBatch,
				jetstream.FetchMaxWait(deleteBookMaxWait),
			)
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue
				}
				ep.logger.Error("fetch error: %w", err)
			}
			msgs := msgBatch.Messages()
			paths := make([]string, 0)
			for msg := range msgs {
				e, err := FromJSON[DeleteBookEvent](msg.Data())
				if err != nil {
					return err
				}
				paths = append(paths, e.Path)
			}
			notDeleted, err := ep.storage.BatchDelete(ctx, paths...)
			if err != nil {
				ep.logger.Error("error while batch deleting books", "error", err)
			}
			if notDeleted != nil {
				for _, nd := range notDeleted {
					// TODO: need to handle it better
					ep.logger.Error("error deleting book from storage", "path", nd.Path, "cause", nd.Cause)
				}
			}
		}
	}
}

func (ep *natsEventProcessor) Run(ctx context.Context) error {
	_, err := ep.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     booksStreamName,
		Subjects: []string{SubjBooksBase + ".*"},
		Storage:  jetstream.FileStorage,
	})
	if err != nil {
		return err
	}
	deleteBookCons, err := ep.js.CreateOrUpdateConsumer(
		ctx,
		booksStreamName,
		jetstream.ConsumerConfig{
			Durable:       deleteBookDurableName,
			DeliverPolicy: jetstream.DeliverAllPolicy,
			FilterSubject: SubjDeleteBook,
		},
	)
	if err != nil {
		return fmt.Errorf("pull subscribe error: %w", err)
	}
	go func() {
		if err := ep.handleDeleteBookEvents(ctx, deleteBookCons); err != nil {
			log.Printf("handler error: %v", err)
		}
	}()
	<-ctx.Done()
	return nil
}
