package repos

import "fmt"

type (
	MessageStatus    uint8
	ProcessingStatus uint8
	OperationType    uint8
)

// Message acknoledgement statuses.
//
// NOTE: statuses are ordered
const (
	MessageStatusNacked    MessageStatus = iota // initial posting by producer
	MessageStatusPosted                         // message ACK-ed by consumer
	MessageStatusReceived                       // response ACK-ed by producer
	MessageStatusConfirmed                      // confirmation ACK-ed by consumer
)

func NewMessageStatus(s MessageStatus) *MessageStatus {
	v := s

	return &v
}

func (s MessageStatus) String() string {
	switch s {
	case MessageStatusNacked:
		return "nacked"
	case MessageStatusPosted:
		return "posted"
	case MessageStatusReceived:
		return "received"
	case MessageStatusConfirmed:
		return "confirmed"
	default:
		panic(fmt.Sprintf("invalid message status: %d", s))
	}
}
func (s MessageStatus) IsValid() bool {
	switch s {
	case MessageStatusNacked, MessageStatusPosted, MessageStatusReceived, MessageStatusConfirmed:
		return true
	default:
		return false
	}
}

func (s MessageStatus) Less(m MessageStatus) bool {
	return uint8(s) < uint8(m)
}

func (s MessageStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil // TODO: perf, could be preallocated slice
}

func (s *MessageStatus) UnmarshalText(data []byte) error {
	in := string(data)
	switch in {
	case "nacked":
		*s = MessageStatusNacked
	case "posted":
		*s = MessageStatusPosted
	case "received":
		*s = MessageStatusReceived
	case "confirmed":
		*s = MessageStatusConfirmed
	default:
		return fmt.Errorf("invalid message status: %q", in)
	}

	return nil
}

// Message result status.
//
// NOTE: statuses are ordered
const (
	ProcessingStatusPending  ProcessingStatus = iota // message being processed
	ProcessingStatusRejected                         // processing outcome decided: rejected
	ProcessingStatusOK                               // processing outcome decided: OK
)

func NewProcessingStatus(s ProcessingStatus) *ProcessingStatus {
	v := s

	return &v
}

func (s ProcessingStatus) String() string {
	switch s {
	case ProcessingStatusPending:
		return "pending"
	case ProcessingStatusRejected:
		return "rejected"
	case ProcessingStatusOK:
		return "ok"
	default:
		panic(fmt.Sprintf("invalid processing status: %d", s))
	}
}

func (s ProcessingStatus) IsValid() bool {
	switch s {
	case ProcessingStatusPending, ProcessingStatusRejected, ProcessingStatusOK:
		return true
	default:
		return false
	}
}

func (s ProcessingStatus) Less(m ProcessingStatus) bool {
	return uint8(s) < uint8(m)
}

func (s ProcessingStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil // TODO: perf, could be preallocated slice
}

func (s *ProcessingStatus) UnmarshalText(data []byte) error {
	in := string(data)
	switch in {
	case "pending":
		*s = ProcessingStatusPending
	case "rejected":
		*s = ProcessingStatusRejected
	case "ok":
		*s = ProcessingStatusOK
	default:
		return fmt.Errorf("invalid processing status: %q", in)
	}

	return nil
}

// Operation types
const (
	OperationTypeDebit OperationType = iota + 1
	OperationTypeCredit
	OperationTypeBalance
	OperationTypeCancel
)

func (s OperationType) String() string {
	switch s {
	case OperationTypeDebit:
		return "debit"
	case OperationTypeCredit:
		return "credit"
	case OperationTypeBalance:
		return "balance"
	case OperationTypeCancel:
		return "cancel"
	default:
		panic(fmt.Sprintf("invalid operation type: %d", s))
	}
}

func (s OperationType) IsValid() bool {
	switch s {
	case OperationTypeDebit, OperationTypeCredit, OperationTypeBalance:
		return true
	default:
		return false
	}
}

func (s OperationType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil // TODO: perf, could be preallocated slice
}

func (s *OperationType) UnmarshalText(data []byte) error {
	in := string(data)
	switch in {
	case "debit":
		*s = OperationTypeDebit
	case "credit":
		*s = OperationTypeCredit
	case "balance":
		*s = OperationTypeBalance
	case "cancel":
		*s = OperationTypeCancel
	default:
		return fmt.Errorf("invalid operation type: %q", in)
	}

	return nil
}
