package repos

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"

	"github.com/cockroachdb/apd/v3"
)

// Errors returned by this repo
var (
	ErrAlreadyProcessed = errors.New("message already processed")
)

type (
	// UpdateOption provides a bit of flexibility to Updates.
	UpdateOption func(*UpdateOptions)

	UpdateOptions struct {
		Force bool
	}

	// Decimal is an exact decimal type that knows
	// how to marshal/unmarshal from JSON and DB.
	//
	// It supports arbitrary decimal precision. Implemented on top of cockroachdb's Decimal,
	// with additions to support gob encode/decode.
	Decimal struct {
		*apd.NullDecimal
	}

	// Message contains a Payload for processing as well a various headers used
	// ensure a no loss transmission.
	Message struct {
		ID         string `json:"id" db:"id"`
		ProducerID string `json:"producer_id" db:"producer_id"`
		ConsumerID string `json:"consumer_id" db:"consumer_id"`

		MessageStatus    MessageStatus    `json:"message_status" db:"message_status"`
		ProcessingStatus ProcessingStatus `json:"processing_status" db:"processing_status"`

		InceptionTime time.Time `json:"inception_time" db:"inception_time"`
		LastTime      time.Time `json:"last_time" db:"last_time"`

		ProducerReplays uint `json:"producer_replays" db:"producer_replays"`
		ConsumerReplays uint `json:"consumer_replays" db:"consumer_replays"`

		Payload `json:"payload"`
	}

	// Payload represents the functional payload of a message.
	//
	// In this example, the message payload is a typical bank transfer.
	Payload struct {
		OperationType   OperationType `json:"operation_type" db:"operation_type"`
		CreditorAccount string        `json:"creditor_account" db:"creditor_account"`
		DebtorAccount   string        `json:"debtor_account" db:"debtor_account"`
		Amount          Decimal       `json:"amount" db:"amount"`
		Currency        string        `json:"currency" db:"currency"`
		BalanceBefore   *Decimal      `json:"balance_before" db:"balance_before"`
		BalanceAfter    *Decimal      `json:"balance_after" db:"balance_after"`
		RejectionCause  *string       `json:"rejection_cause" db:"rejection_cause"`
	}

	// InputPayload represents the accepted user input for a message.
	InputPayload struct {
		CorrespondantBank string        `json:"correspondant_bank"`
		OperationType     OperationType `json:"operation_type" db:"operation_type"`
		CreditorAccount   string        `json:"creditor_account" db:"creditor_account"`
		DebtorAccount     string        `json:"debtor_account" db:"debtor_account"`
		Amount            Decimal       `json:"amount" db:"amount"`
		Currency          string        `json:"currency" db:"currency"`
	}

	// MessagePredicate is used to specify filters when querying Messages
	MessagePredicate struct {
		UpdatedSince         *time.Time
		WithMessageStatus    *MessageStatus
		WithProcessingStatus *ProcessingStatus
		MaxMessageStatus     *MessageStatus
		MaxProcessingStatus  *ProcessingStatus
		FromProducer         *string
		FromConsumer         *string
		Limit                uint64
		Unconfirmed          bool

		_ struct{}
	}
)

func (p InputPayload) AsMessage() Message {
	return Message{
		ConsumerID: p.CorrespondantBank,
		Payload: Payload{
			OperationType:   p.OperationType,
			CreditorAccount: p.CreditorAccount,
			DebtorAccount:   p.DebtorAccount,
			Amount:          p.Amount,
			Currency:        p.Currency,
		},
	}
}

func (p Message) Validate() error {
	// TODO
	// TODO: validation - check ConsumerID is legit
	return p.Payload.Validate()
}

func (p Payload) Validate() error {
	// TODO
	return nil
}

func (p Message) Bytes() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func NewDecimal(coef int64, exponent int32) *Decimal {
	return &Decimal{NullDecimal: &apd.NullDecimal{Decimal: *apd.New(coef, exponent), Valid: true}}
}

func (d *Decimal) UnmarshalText(data []byte) error {
	if d == nil || data == nil {
		d.NullDecimal = &apd.NullDecimal{}

		return nil
	}

	if d.NullDecimal == nil {
		b := apd.New(0, 2)
		if err := b.UnmarshalText(data); err != nil {
			return err
		}

		d.NullDecimal = &apd.NullDecimal{Decimal: *b, Valid: true}

		return nil
	}

	if err := d.Decimal.UnmarshalText(data); err != nil {
		return err
	}

	d.Valid = true

	return nil
}

func (d *Decimal) GobDecode(data []byte) error {
	return d.UnmarshalText(data)
}

func (d Decimal) GobEncode() ([]byte, error) {
	if d.NullDecimal == nil || !d.NullDecimal.Valid {
		return []byte{}, nil
	}

	return d.Decimal.MarshalText()
}
func (d *Decimal) Scan(value interface{}) error {
	if d.NullDecimal == nil {
		b := &apd.NullDecimal{}

		if err := b.Scan(value); err != nil {
			return err
		}

		d.NullDecimal = b

		return nil
	}

	return d.NullDecimal.Scan(value)
}

func WithForceUpdate(enabled bool) UpdateOption {
	return func(o *UpdateOptions) {
		o.Force = enabled
	}
}
