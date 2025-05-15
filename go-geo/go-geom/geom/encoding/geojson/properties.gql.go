package geojson

import (
	"errors"
	"io"
)

func (p Properties) UnmarshalGQL(v interface{}) error {
	props, ok := v.(string)
	if !ok {
		return errors.New("geojson properties must be strings")
	}
	if props == "" {
		return nil
	}

	return json.UnmarshalFromString(props, &p)
}

func (p Properties) MarshalGQL(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(p)
}
