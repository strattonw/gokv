package marshal

import (
	"errors"
	"fmt"
	"github.com/philippgille/gokv/util"
	"sync"
)

// MarshalFormat is an enum for the available (un-)marshal formats of this gokv.Store implementation.
type Format int

const (
	// JSON is the MarshalFormat for (un-)marshalling to/from JSON
	JSON Format = iota
	// Gob is the MarshalFormat for (un-)marshalling to/from gob
	Gob
)

type Marshaller interface {
	Marshal(string, interface{}) ([]byte, error)
}

type MarshallerFunc func(string, interface{}) ([]byte, error)

func (mf MarshallerFunc) Marshal(k string, v interface{}) ([]byte, error) {
	return mf(k, v)
}

type Register interface {
	Register(Format, Marshaller)
	Marshal(string, interface{}, Format) ([]byte, error)
}

type marshal struct {
	Marshaller map[Format]Marshaller
	mu         *sync.RWMutex
}

func (m *marshal) Register(f Format, mar Marshaller) {
	m.mu.Lock()
	m.Marshaller[f] = mar
	m.mu.Unlock()
}

func (m *marshal) Marshal(k string, v interface{}, f Format) ([]byte, error) {
	m.mu.RLock()
	mar := m.Marshaller[f]
	m.mu.RUnlock()

	if mar == nil {
		return make([]byte, 0), errors.New(fmt.Sprintf("Could not find marshaller for format %d", f))
	}

	return mar.Marshal(k, v)
}

var DefaultMarshalRegister Register = &marshal{
	Marshaller: map[Format]Marshaller{
		JSON: MarshallerFunc(func(k string, v interface{}) ([]byte, error) {
			if err := util.CheckKeyAndValue(k, v); err != nil {
				return make([]byte, 0), err
			}

			return util.ToJSON(v)
		}),
		Gob: MarshallerFunc(func(k string, v interface{}) ([]byte, error) {
			if err := util.CheckKeyAndValue(k, v); err != nil {
				return make([]byte, 0), err
			}

			return util.ToGob(v)
		}),
	},
}
