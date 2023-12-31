package repos

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/almerlucke/go-iban/iban"
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
		BalanceBefore   *Decimal      `json:"balance_before,omitempty" db:"balance_before"`
		BalanceAfter    *Decimal      `json:"balance_after,omitempty" db:"balance_after"`
		Comment         *string       `json:"comment,omitempty" db:"comment"`
		RejectionCause  *string       `json:"rejection_cause,omitempty" db:"rejection_cause"`
	}

	// InputPayload represents the accepted user input for a message.
	InputPayload struct {
		CorrespondantBank string        `json:"correspondant_bank"`
		OperationType     OperationType `json:"operation_type"`
		CreditorAccount   string        `json:"creditor_account"`
		DebtorAccount     string        `json:"debtor_account"`
		Amount            Decimal       `json:"amount"`
		Currency          string        `json:"currency"`
		Comment           *string       `json:"comment,omitempty"`
	}

	// MessagePredicate is used to specify filters when querying Messages
	MessagePredicate struct {
		UpdatedSince         *time.Time
		NotUpdatedSince      *time.Time
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
			Comment:         p.Comment,
		},
	}
}

func (p Message) Validate() error {
	// TODO: validation - check ConsumerID is legit
	return p.Payload.Validate()
}

var zero = apd.New(0, 2)

// Validate the presence of required field and legit IBAN account identifiers.
func (p Payload) Validate() error {
	if !p.OperationType.IsValid() {
		return fmt.Errorf("invalid operation type: %d", p.OperationType)
	}

	if len(p.CreditorAccount) == 0 {
		return fmt.Errorf("required creditor account: %q", p.CreditorAccount)
	}

	if len(p.DebtorAccount) == 0 {
		return fmt.Errorf("required debtor account: %q", p.DebtorAccount)
	}

	if len(p.Currency) != 3 {
		return fmt.Errorf("invalid currency: %q", p.Currency)
	}

	if !p.Amount.Valid || p.Amount.Decimal.Cmp(zero) <= 0 {
		return fmt.Errorf("invalid amount: %v", p.Amount)
	}

	if p.Comment != nil && len(*p.Comment) > 255 {
		return fmt.Errorf("comment is too long: %d chars", len(*p.Comment))
	}

	if _, err := iban.NewIBAN(p.CreditorAccount); err != nil {
		return fmt.Errorf("creditor account is an invalid IBAN: %q: %w", p.CreditorAccount, err)
	}

	if _, err := iban.NewIBAN(p.DebtorAccount); err != nil {
		return fmt.Errorf("debtor account is an invalid IBAN: %q: %w", p.DebtorAccount, err)
	}

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

func WithForceUpdate(enabled bool) UpdateOption {
	return func(o *UpdateOptions) {
		o.Force = enabled
	}
}
