package repos

import (
	"context"

	"github.com/fredbi/go-patterns/iterators"
)

type (
	MessageIterator = iterators.StructIterator[Message]

	// Iface serves interfaces to persistent repositories.
	Iface interface {
		Messages() MessageRepo
	}

	// MessageRepo exposes the persistent repository for messages.
	MessageRepo interface {
		// Create a new Message in the DB
		Create(context.Context, Message) error

		// Update a Message in the DB.
		Update(context.Context, Message, ...UpdateOption) error

		// Get retrieves a Message by its unique ID
		Get(context.Context, string) (Message, error)

		// List Messages using a MessagePredicate filter. The result is an iterator to the
		// fetched rows.
		List(context.Context, MessagePredicate) (MessageIterator, error)

		// UpdateConfirmed is an entry point for consumers to store their own view of the message status
		UpdateConfirmed(context.Context, string, MessageStatus) error

		// UpdateReplay is an entry point for producer keep track of how many times messages have been replayed
		UpdateReplay(context.Context, Message, ...UpdateOption) error
	}
)
