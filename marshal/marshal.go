package marshal

import (
	"errors"
	"github.com/philippgille/gokv/util"
)

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type Format int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON Format = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

func Marshal(k string, v interface{}, f Format) ([]byte, error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return make([]byte, 0), err
	}

	// First turn the passed object into something that Redis can handle
	// (the Set method takes an interface{}, but the Get method only returns a string,
	// so it can be assumed that the interface{} parameter type is only for convenience
	// for a couple of builtin types like int etc.).
	switch f {
	case JSON:
		return util.ToJSON(v)
	case Gob:
		return util.ToGob(v)
	default:
		return make([]byte, 0), errors.New("the store seems to be configured with a marshal format that's not implemented yet")
	}
}