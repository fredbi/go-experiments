package repo

import (
	"time"
)

const (
	MessageStatusNacked MessageStatus = iota
	MessageStatusPosted
	MessageStatusReceived
)

const (
	ProcessingStatusPending ProcessingStatus = iota
	ProcessingStatusRejected
	ProcessingStatusOK
)

type (
	MessageStatus    uint8
	ProcessingStatus uint8

	Message struct {
		ID               string           `json:"id" db:"id"`
		ProducerID       string           `json:"producer_id" db:"producer_id"`
		ConsumerID       string           `json:"consumer_id" db:"consumer_id"`
		MessageStatus    MessageStatus    `json:"message_status" db:"message_status"`
		ProcessingStatus ProcessingStatus `json:"processing_status" db:"processing_status"`

		InceptionTime time.Time `json:"inception_time" db:"inception_time"`
		LastTime      time.Time `json:"last_time" db:"last_time"`

		ProducerReplays uint `json:"producer_replays" db:"producer_replays"`
		ConsumerReplays uint `json:"consumer_replays" db:"consumer_replays"`

		Payload `json:"payload"`
	}

	Payload struct {
		OperationName string `json:"operation_name" db:"operation_name"`
		Result        string `json:"result" db:"result"`
	}
)
