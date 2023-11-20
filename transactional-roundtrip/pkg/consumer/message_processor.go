package consumer

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	"go.uber.org/zap"
)

type (
	// MessageProcessor knows how to process a message.
	MessageProcessor interface {
		Process(context.Context, *repos.Message) error
	}

	// DummyProcessor processes messages with some random behavior for testing purpose.
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

func NewDummyProcessor(rt injected.Runtime) *DummyProcessor {
	return &DummyProcessor{
		rt:                     rt,
		dummyProcessorSettings: defaultDummyProcessorSettings,
	}
}

// Process a message with some random behavior.
//
// Notice that this does not interact with the database, primarily owned by the producer node.
func (p DummyProcessor) Process(ctx context.Context, msg *repos.Message) error {
	l := p.rt.Logger().For(ctx)

	// may fail and need a restart
	toss := rand.Float64() //#nosec
	if toss < p.HardFailureRate {
		// that's rude, but let the k8s controller restart this container
		fmt.Fprintln(os.Stderr, "OMG, they killed Kenny")

		if err := p.audit(ctx, "hardfailure", msg); err != nil {
			l.Error("audit failed", zap.Error(err))

			return err
		}
		_ = l.Zap().Sync()

		os.Exit(1)
	}

	// may fail temporarily
	toss = rand.Float64() //#nosec
	if toss < p.FailureRate {
		if err := p.audit(ctx, "softfailure", msg); err != nil {
			l.Error("audit failed", zap.Error(err))

			return err
		}

		return errors.New("processing failed")
	}

	// for demo, just put random numbers to fill-in the balances
	msg.BalanceBefore = repos.NewDecimal(rand.Int63n(p.BalanceMax), 2) //#nosec
	msg.BalanceAfter = repos.NewDecimal(rand.Int63n(p.BalanceMax), 2)  //#nosec

	// may reject definitively
	toss = rand.Float64() // #nosec
	if toss < p.RejectionRate {
		msg.ProcessingStatus = repos.ProcessingStatusRejected
		cause := "because of bad luck"
		msg.RejectionCause = &cause
		l.Warn("rejected message", zap.String("id", msg.ID), zap.Stringp("cause", msg.RejectionCause))

		if err := p.audit(ctx, "rejected", msg); err != nil {
			l.Error("audit failed", zap.Error(err))

			return err
		}
	} else {
		msg.ProcessingStatus = repos.ProcessingStatusOK
		msg.RejectionCause = nil

		if err := p.audit(ctx, "ok", msg); err != nil {
			l.Error("audit failed", zap.Error(err))

			return err
		}
	}

	return nil
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// audit process actions (so we can follow-up with a mass injection)
func (p DummyProcessor) audit(ctx context.Context, action string, msg *repos.Message) error {
	db := p.rt.DB()

	query := psql.Insert("process_audit").Columns(
		"id", "processing_status", "action").Values(msg.ID, msg.ProcessingStatus, action)

	q, args := query.MustSql()
	_, err := db.ExecContext(ctx, q, args...)

	return err
}
