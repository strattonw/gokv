package syncmap_test

import (
	"testing"

	"github.com/philippgille/gokv/syncmap"
	"github.com/philippgille/gokv/test"
)

// TestStore tests if reading and writing to the store works properly.
func TestStore(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store := createStore(t, syncmap.JSON)
		test.TestStore(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store := createStore(t, syncmap.Gob)
		test.TestStore(store, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store := createStore(t, syncmap.JSON)
		test.TestTypes(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store := createStore(t, syncmap.Gob)
		test.TestTypes(store, t)
	})
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
// The store is a sync.Map, so the concurrency should be supported by the used package.
func TestStoreConcurrent(t *testing.T) {
	store := createStore(t, syncmap.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, store)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test with a bad MarshalFormat enum value

	store := createStore(t, syncmap.MarshalFormat(19))
	err := store.Set("foo", "bar")
	if err == nil {
		t.Error("An error should have occurred, but didn't")
	}
	// TODO: store some value for "foo", so retrieving the value works.
	// Just the unmarshalling should fail.
	// _, err = store.Get("foo", new(string))
	// if err == nil {
	// 	t.Error("An error should have occurred, but didn't")
	// }

	// Test empty key
	err = store.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = store.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = store.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		store := createStore(t, syncmap.JSON)
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		store := createStore(t, syncmap.Gob)
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf syncmap.MarshalFormat) func(t *testing.T) {
		return func(t *testing.T) {
			store := createStore(t, mf)

			// Prep
			err := store.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = store.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
			_, err = store.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = store.Get("foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(syncmap.JSON))
	t.Run("get with nil / nil value parameter", createTest(syncmap.Gob))
}

// TestClose tests if the close method returns any errors.
func TestClose(t *testing.T) {
	store := createStore(t, syncmap.JSON)
	err := store.Close()
	if err != nil {
		t.Error(err)
	}
}

func createStore(t *testing.T, mf syncmap.MarshalFormat) syncmap.Store {
	options := syncmap.Options{
		MarshalFormat: mf,
	}
	store := syncmap.NewStore(options)
	return store
}
