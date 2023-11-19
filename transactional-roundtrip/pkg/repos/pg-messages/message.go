package messages

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/fredbi/go-patterns/iterators"
	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/tracer"
	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// type safeguard at build time
	_ repos.MessageRepo = &Repo{}

	messageColumns = []string{
		"id",
		"producer_id",
		"consumer_id",
		"message_status",
		"processing_status",
		"inception_time",
		"last_time",
		"producer_replays",
		"consumer_replays",
		"operation_type",
		"creditor_account",
		"debtor_account",
		"amount",
		"currency",
		"balance_before",
		"balance_after",
		"rejection_cause",
		"comment",
	}

	_ repos.MessageRepo = &Repo{}
)

// New instance of the sample repository
func New(db *sqlx.DB, log log.Factory, cfg *viper.Viper) *Repo {
	return &Repo{
		log: log,
		db:  db,
		cfg: cfg,
	}
}

// Repo implements the repos.MessageRepo interface against a postgres DB.
type Repo struct {
	log log.Factory
	db  *sqlx.DB
	cfg *viper.Viper

	_ struct{} // prevents unkeyed initialization
}

func (r *Repo) DB() *sqlx.DB {
	return r.db
}

// Logger used by tracer
func (r *Repo) Logger() log.Factory {
	return r.log
}

func (r *Repo) Get(parentCtx context.Context, id string) (repos.Message, error) {
	ctx, span, lg := tracer.StartSpan(parentCtx, r)
	defer span.End()

	query := psql.Select(messageColumns...).From("message").Where(sq.Eq{"id": id})
	q, args := query.MustSql()
	lg.Debug("get message query", zap.String("sql", q), zap.Any("args", args))

	var message repos.Message
	err := r.DB().QueryRowxContext(ctx, q, args...).StructScan(&message)

	return message, err
}

func (r *Repo) Create(parentCtx context.Context, message repos.Message) error {
	ctx, span, lg := tracer.StartSpan(parentCtx, r)
	defer span.End()

	query := psql.Insert("message").Columns(messageColumns...).Values(
		message.ID,
		message.ProducerID,
		message.ConsumerID,
		message.MessageStatus,
		message.ProcessingStatus,
		message.InceptionTime,
		message.LastTime,
		message.ProducerReplays,
		message.ConsumerReplays,
		message.OperationType,
		message.CreditorAccount,
		message.DebtorAccount,
		message.Amount,
		message.Currency,
		message.BalanceBefore,
		message.BalanceAfter,
		message.RejectionCause,
		message.Comment,
	).Suffix(`ON CONFLICT (id) DO NOTHING`) // mute duplicate errors: this is captured by the row count = 0

	q, args := query.MustSql()
	lg.Debug("insert message statement", zap.String("sql", q), zap.Any("args", args))

	// start an explicit transaction, we know exactly what's happening: no autocommit magic
	cancellable, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := r.DB().BeginTxx(cancellable, nil)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(cancellable, q, args...)
	if err != nil {
		return err
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return repos.ErrAlreadyProcessed
	}

	return tx.Commit()
}

// Update a message, specifically to maintain a producer's view.
//
// Reminder: in this demo, consumers do not update the database.
//
// However, it is required for consumers to track confirmed messages:
// confirmed messages won't be replayed any longer.
func (r *Repo) UpdateConfirmed(parentCtx context.Context, id string, messageStatus repos.MessageStatus) error {
	ctx, span, lg := tracer.StartSpan(parentCtx, r)
	defer span.End()

	query := psql.Update("message").
		Where(sq.Eq{"id": id}).
		Set("consumer_message_status", messageStatus).
		Where(sq.Expr(`consumer_message_status < ?`, messageStatus))

	q, args := query.MustSql()
	lg.Debug("update message statement", zap.String("sql", q), zap.Any("args", args))

	cancellable, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := r.DB().BeginTxx(cancellable, nil)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(cancellable, q, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if n, _ := result.RowsAffected(); n == 0 {
		return repos.ErrAlreadyProcessed
	}

	return tx.Commit()
}

// Update a message.
//
// Only statuses, last_time, replays, balances and rejection cause are mutable.
//
// An update is performed only if the message status is strictly increasing or if
// the message status remains unchanged and the processing status increases.
// Otherwise, the operation is ignored.
func (r *Repo) Update(parentCtx context.Context, message repos.Message, opts ...repos.UpdateOption) error {
	ctx, span, lg := tracer.StartSpan(parentCtx, r)
	defer span.End()

	o := repos.UpdateOptions{}
	for _, apply := range opts {
		apply(&o)
	}

	query := psql.Update("message").
		SetMap(map[string]interface{}{
			"message_status":    message.MessageStatus,
			"processing_status": message.ProcessingStatus,
			"last_time":         message.LastTime,
			"producer_replays":  message.ProducerReplays,
			"consumer_replays":  message.ConsumerReplays,
			"balance_before":    message.BalanceBefore,
			"balance_after":     message.BalanceAfter,
			"rejection_cause":   message.RejectionCause,
		}).
		Where(sq.Eq{"id": message.ID})

	if !o.Force {
		// unless we want specifically to apply some update (in the case of the consumer, as we use the same table
		// as a convenient implementation), we forbid updates which do not bring some progress to the state of
		// the message.
		query = query.Where(sq.Expr(
			`(message_status < ?) OR (message_status = ? AND processing_status < ?)`,
			message.MessageStatus, message.MessageStatus, message.ProcessingStatus,
		))
	}
	q, args := query.MustSql()
	lg.Debug("update message statement", zap.String("sql", q), zap.Any("args", args))

	cancellable, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := r.DB().BeginTxx(cancellable, nil)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(cancellable, q, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if o.Force && err != nil {
		return err
	}

	if n, _ := result.RowsAffected(); n == 0 {
		// well this is a bit optimistic about the behavior of the app: could be that the ID is not found...
		// A well-behaving app does not fall into this trap however, so we're good with ErrAlreadyProcessed.
		return repos.ErrAlreadyProcessed
	}

	return tx.Commit()
}

func (r *Repo) List(ctx context.Context, p repos.MessagePredicate) (repos.MessageIterator, error) {
	ctx, span, lg := tracer.StartSpan(ctx, r)
	defer span.End()

	query := psql.Select(messageColumns...).From("message").OrderBy("last_time")

	// add predicate filters
	if p.UpdatedSince != nil {
		query = query.Where(sq.Expr("last_time > ?", *p.UpdatedSince))
	}
	if p.NotUpdatedSince != nil {
		query = query.Where(sq.Expr("last_time < ?", *p.NotUpdatedSince))
	}
	if p.WithMessageStatus != nil {
		query = query.Where(sq.Eq{"message_status": *p.WithMessageStatus})
	}
	if p.WithProcessingStatus != nil {
		query = query.Where(sq.Eq{"processing_status": *p.WithProcessingStatus})
	}
	if p.MaxMessageStatus != nil {
		query = query.Where(sq.Expr("message_status < ?", *p.MaxMessageStatus))
	}
	if p.MaxProcessingStatus != nil {
		query = query.Where(sq.Expr("processing_status < ?", *p.MaxProcessingStatus))
	}
	if p.FromProducer != nil {
		query = query.Where(sq.Eq{"producer_id": *p.FromProducer})
	}
	if p.FromConsumer != nil {
		query = query.Where(sq.Eq{"consumer_id": *p.FromConsumer})
	}
	if p.Limit > 0 {
		query = query.Limit(p.Limit)
	}
	if p.Unconfirmed {
		query = query.Where(sq.Expr("consumer_message_status < ?", repos.MessageStatusConfirmed))
	}

	q, args := query.MustSql()
	lg.Debug("list message query", zap.String("sql", q), zap.Any("args", args))

	rows, err := r.DB().QueryxContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	return iterators.NewSqlxIterator[repos.Message](rows), nil
}
