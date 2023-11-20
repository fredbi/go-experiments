//nolint:unused,revive
package nats

import zap "github.com/jackc/pgx-zap"

// natsLogAdapter implements the interface of the NATS Server, with an underlying zap.Logger
//
// TODO(fred): leverage logging of the embedded server, while keeping structured logging.
type natsLogAdapter struct {
	zlg *zap.Logger
}

func newNatsLogAdapter(zlg *zap.Logger) *natsLogAdapter {
	return &natsLogAdapter{zlg: zlg}
}

// Log a notice statement
func (n natsLogAdapter) Noticef(format string, v ...interface{}) {
}

// Log a warning statement
func (n natsLogAdapter) Warnf(format string, v ...interface{}) {
}

// Log a fatal error
func (n natsLogAdapter) Fatalf(format string, v ...interface{}) {
}

// Log an error
func (n natsLogAdapter) Errorf(format string, v ...interface{}) {
}

// Log a debug statement
func (n natsLogAdapter) Debugf(format string, v ...interface{}) {
}

// Log a trace statement
func (n natsLogAdapter) Tracef(format string, v ...interface{}) {
}
