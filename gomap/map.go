package gomap

import (
	"errors"
	"github.com/philippgille/gokv/marshal"
	"sync"

	"github.com/philippgille/gokv/util"
)

// Store is a gokv.Store implementation for a Go map with a sync.RWMutex for concurrent access.
type Store struct {
	m             map[string][]byte
	lock          *sync.RWMutex
	marshalFormat marshal.Format
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (m Store) Set(k string, v interface{}) error {
	data, err := marshal.Marshal(k, v, m.marshalFormat)

	if err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[k] = data
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (m Store) Get(k string, v interface{}) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	m.lock.RLock()
	data, found := m.m[k]
	// Unlock right after reading instead of with defer(),
	// because following unmarshalling will take some time
	// and we don't want to block writing threads until that's done.
	m.lock.RUnlock()
	if !found {
		return false, nil
	}

	switch m.marshalFormat {
	case marshal.JSON:
		return true, util.FromJSON(data, v)
	case marshal.Gob:
		return true, util.FromGob(data, v)
	default:
		return true, errors.New("The store seems to be configured with a marshal format that's not implemented yet")
	}
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (m Store) Delete(k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	delete(m.m, k)
	return nil
}

// Close closes the store.
// When called, the store's pointer to the internal Go map is set to nil,
// leading to the map being free for garbage collection.
func (m Store) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m = nil
	return nil
}


// Options are the options for the Go map store.
type Options struct {
	// (Un-)marshal format.
	// Optional (JSON by default).
	MarshalFormat marshal.Format
}

// DefaultOptions is an Options object with default values.
// MarshalFormat: JSON
var DefaultOptions = Options{
	// No need to set MarshalFormat to JSON
	// because its zero value is fine.
}

// NewStore creates a new Go map store.
func NewStore(options Options) Store {
	return Store{
		m:             make(map[string][]byte),
		lock:          new(sync.RWMutex),
		marshalFormat: options.MarshalFormat,
	}
}
