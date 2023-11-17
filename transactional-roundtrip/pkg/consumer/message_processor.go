package consumer

import (
	"context"
	"errors"
	"math/rand"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
)

type MessageProcessor interface {
	Process(context.Context, *repos.Message) error
}

// DummyProcessor processes messages with some random behavior for testing purpose.
//
// TODO: add some more realistic processing (e.g. build output files...)
type DummyProcessor struct {
	rt injected.Runtime

	settings
}

func NewDummyProcessor() *DummyProcessor {
	return &DummyProcessor{}
}

func (p DummyProcessor) Process(_ context.Context, msg *repos.Message) error {
	toss := rand.Float64()
	if toss < 0.1 {
		return errors.New("processing failed")
	}

	// for demo, just put random numbers to fill-in the balances
	msg.BalanceBefore = repos.NewDecimal(rand.Int63n(1_000_000_000), 2) //#nosec
	msg.BalanceAfter = repos.NewDecimal(rand.Int63n(1_000_000_000), 2)  //#nosec

	toss = rand.Float64()
	if toss < 0.1 {
		msg.ProcessingStatus = repos.ProcessingStatusRejected
		cause := "because of bad luck"
		msg.RejectionCause = &cause
	} else {
		msg.ProcessingStatus = repos.ProcessingStatusOK
	}

	return nil
}
