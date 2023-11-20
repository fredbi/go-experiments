package repos

import (
	"github.com/cockroachdb/apd/v3"
)

// Decimal is an exact decimal type that knows
// how to marshal/unmarshal from JSON and DB.
//
// It supports arbitrary decimal precision. Implemented on top of cockroachdb's Decimal,
// with additions to support gob encode/decode.
type Decimal struct {
	*apd.NullDecimal
}

func NewDecimal(coef int64, exponent int32) *Decimal {
	return &Decimal{NullDecimal: &apd.NullDecimal{Decimal: *apd.New(coef, exponent), Valid: true}}
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return d.Decimal.MarshalText()
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	return d.UnmarshalText(data)
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
