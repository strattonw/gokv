package mongodb_test

import (
	"log"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/philippgille/gokv/mongodb"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
//
// Note: This test is only executed if the initial connection to MongoDB works.
func TestClient(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MongoDB could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, mongodb.JSON)
		test.TestStore(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, mongodb.Gob)
		test.TestStore(client, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
//
// Note: This test is only executed if the initial connection to MongoDB works.
func TestTypes(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MongoDB could be established. Probably not running in a proper test environment.")
	}

	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		client := createClient(t, mongodb.JSON)
		test.TestTypes(client, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		client := createClient(t, mongodb.Gob)
		test.TestTypes(client, t)
	})
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the MongoDB client.
//
// Note: This test is only executed if the initial connection to MongoDB works.
func TestClientConcurrent(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MongoDB could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, mongodb.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, client)
}

// TestErrors tests some error cases.
//
// Note: This test is only executed if the initial connection to MongoDB works.
func TestErrors(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MongoDB could be established. Probably not running in a proper test environment.")
	}

	// Test with a bad MarshalFormat enum value

	client := createClient(t, mongodb.MarshalFormat(19))
	err := client.Set("foo", "bar")
	if err == nil {
		t.Error("An error should have occurred, but didn't")
	}
	// TODO: store some value for "foo", so retrieving the value works.
	// Just the unmarshalling should fail.
	// _, err = client.Get("foo", new(string))
	// if err == nil {
	// 	t.Error("An error should have occurred, but didn't")
	// }

	// Test empty key
	err = client.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = client.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = client.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}

	// Test bad connection string
	options := mongodb.Options{
		ConnectionString: "forceError!",
	}
	_, err = mongodb.NewClient(options)
	if err.Error() != "no reachable servers" {
		t.Errorf(`Expected a "no reachable servers" error, but instead got: %v`, err)
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
//
// Note: This test is only executed if the initial connection to MongoDB works.
func TestNil(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MongoDB could be established. Probably not running in a proper test environment.")
	}

	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		client := createClient(t, mongodb.JSON)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		client := createClient(t, mongodb.Gob)
		err := client.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(mf mongodb.MarshalFormat) func(t *testing.T) {
		return func(t *testing.T) {
			client := createClient(t, mf)

			// Prep
			err := client.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = client.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
			_, err = client.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = client.Get("foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(mongodb.JSON))
	t.Run("get with nil / nil value parameter", createTest(mongodb.Gob))
}

// TestClose tests if the close method returns any errors.
//
// Note: This test is only executed if the initial connection to MongoDB works.
func TestClose(t *testing.T) {
	if !checkConnection() {
		t.Skip("No connection to MongoDB could be established. Probably not running in a proper test environment.")
	}

	client := createClient(t, mongodb.JSON)
	err := client.Close()
	if err != nil {
		t.Error(err)
	}
}

// checkConnection returns true if a connection could be made, false otherwise.
func checkConnection() bool {
	session, err := mgo.DialWithTimeout("localhost", 2*time.Second)
	if err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	if err = session.Ping(); err != nil {
		log.Printf("An error occurred during testing the connection to the server: %v\n", err)
		return false
	}
	return true
}

func createClient(t *testing.T, mf mongodb.MarshalFormat) mongodb.Client {
	options := mongodb.Options{}
	options.MarshalFormat = mf
	client, err := mongodb.NewClient(options)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
