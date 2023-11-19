package consumer

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
)

type (
	// MessageProcessor knows how to process a message.
	MessageProcessor interface {
		Process(context.Context, *repos.Message) error
	}

	// DummyProcessor processes messages with some random behavior for testing purpose.
	//
	// TODO: add some more realistic processing (e.g. build output files...)
	DummyProcessor struct {
		rt injected.Runtime //nolint:unused

		dummyProcessorSettings
	}

	dummyProcessorSettings struct {
		BalanceMax      int64   // balances in response are random integers, but we want to avoid overflow in DB
		FailureRate     float64 // probability of Process returning an error
		RejectionRate   float64 // probabilty of Process rejecting the message
		HardFailureRate float64 // probability of Process exiting right away. By default, this is disabled
	}
)

var defaultDummyProcessorSettings = dummyProcessorSettings{
	BalanceMax:      1_000_000_000, // max precision in DB column is numeric(15,2)
	FailureRate:     0.1,
	RejectionRate:   0.1,
	HardFailureRate: 0,
}

func NewDummyProcessor() *DummyProcessor {
	return &DummyProcessor{
		dummyProcessorSettings: defaultDummyProcessorSettings,
	}
}

// Process a message with some random behavior.
//
// Notice that this does not interact with the database, primarily owned by the producer node.
func (p DummyProcessor) Process(_ context.Context, msg *repos.Message) error {
	toss := rand.Float64() //#nosec
	if toss < p.HardFailureRate {
		// that's rude, but let the k8s controller restart this container
		fmt.Fprintln(os.Stderr, "OMG, they killed Kenny")

		os.Exit(1)
	}

	toss = rand.Float64() //#nosec
	if toss < p.FailureRate {
		return errors.New("processing failed")
	}

	// for demo, just put random numbers to fill-in the balances
	msg.BalanceBefore = repos.NewDecimal(rand.Int63n(p.BalanceMax), 2) //#nosec
	msg.BalanceAfter = repos.NewDecimal(rand.Int63n(p.BalanceMax), 2)  //#nosec

	toss = rand.Float64() // #nosec
	if toss < p.RejectionRate {
		msg.ProcessingStatus = repos.ProcessingStatusRejected
		cause := "because of bad luck"
		msg.RejectionCause = &cause
	} else {
		msg.ProcessingStatus = repos.ProcessingStatusOK
	}

	return nil
}
