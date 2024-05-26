package csvutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// Define structs for testing nested objects
type Address struct {
	Street string `csv:"street"`
	City   string `csv:"city"`
}

type Person struct {
	Name    string  `csv:"name"`
	Age     int     `csv:"age"`
	Address Address `csv:"address"`
}

// Define a struct with a pointer to an object
type PersonPtr struct {
	Name    string   `csv:"name"`
	Age     int      `csv:"age"`
	Address *Address `csv:"address"`
}

// Create a temporary file with the given content and return its path
func createTempFile(t *testing.T, content string) string {
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed to close temporary file: %v", err)
	}

	return tmpfile.Name()
}

func TestReadCSV_NestedObject(t *testing.T) {
	csvData := `name,age,address_street,address_city
John,30,Main St,New York
Jane,25,Elm St,Boston
`

	csvFilePath := createTempFile(t, csvData)
	defer os.Remove(csvFilePath) // Clean up

	// Set up a handler function for testing nested objects
	var testRecords []*Person
	handler := func(record interface{}) error {
		testRecords = append(testRecords, record.(*Person))
		return nil
	}

	// Read the CSV data
	err := ReadCSV(csvFilePath, &Person{}, WithHandler(handler))
	if err != nil {
		t.Fatalf("error reading CSV: %v", err)
	}

	// Verify the records
	expected := []*Person{
		{Name: "John", Age: 30, Address: Address{Street: "Main St", City: "New York"}},
		{Name: "Jane", Age: 25, Address: Address{Street: "Elm St", City: "Boston"}},
	}

	if !reflect.DeepEqual(testRecords, expected) {
		t.Errorf("nested object records mismatch\nExpected: %v\nGot: %v", expected, testRecords)
	}
}

func TestReadCSV_PointerToObject(t *testing.T) {
	csvData := `name,age,address_street,address_city
John,30,Main St,New York
Jane,25,Elm St,Boston
`

	csvFilePath := createTempFile(t, csvData)
	defer os.Remove(csvFilePath) // Clean up

	// Set up a handler function for testing pointers to objects
	var testRecords []*PersonPtr
	handler := func(record interface{}) error {
		testRecords = append(testRecords, record.(*PersonPtr))
		return nil
	}

	// Read the CSV data
	err := ReadCSV(csvFilePath, &PersonPtr{}, WithHandler(handler))
	if err != nil {
		t.Fatalf("error reading CSV: %v", err)
	}

	// Verify the records
	expected := []*PersonPtr{
		{Name: "John", Age: 30, Address: &Address{Street: "Main St", City: "New York"}},
		{Name: "Jane", Age: 25, Address: &Address{Street: "Elm St", City: "Boston"}},
	}

	if !reflect.DeepEqual(testRecords, expected) {
		t.Errorf("pointer to object records mismatch\nExpected: %v\nGot: %v", expected, testRecords)
	}
}

func TestReadCSV_Concurrency(t *testing.T) {
	// Create a large CSV data
	csvData := "name,age,address_street,address_city\n"
	for i := 0; i < 10000; i++ {
		csvData += "John," + strconv.Itoa(i) + ",Main St,New York\n"
	}

	// Create a temporary CSV file
	csvFilePath := createTempFile(t, csvData)
	defer os.Remove(csvFilePath) // Clean up

	// Set up a handler function
	mx := sync.Mutex{}
	actualMap := make(map[string]*Person)
	handler := func(record interface{}) error {
		rc := record.(*Person)
		key := fmt.Sprintf("%s_%d", rc.Name, rc.Age)
		mx.Lock()
		actualMap[key] = rc
		mx.Unlock()
		return nil
	}

	// Read the CSV data concurrently
	err := ReadCSV(csvFilePath, &Person{}, WithHandler(handler), WithConcurrency(10))
	if err != nil {
		t.Fatalf("error reading CSV: %v", err)
	}

	// Define the expected records
	expectedMap := make(map[string]*Person)
	for i := 0; i < 10000; i++ {
		record := &Person{Name: "John", Age: i, Address: Address{Street: "Main St", City: "New York"}}
		key := fmt.Sprintf("%s_%d", record.Name, record.Age)
		expectedMap[key] = record
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("pointer to object records mismatch\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}
func createExpectedRecords() map[string]*Person {
	expectedMap := make(map[string]*Person)
	for i := 0; i < 10000; i++ {
		record := &Person{Name: "John", Age: i, Address: Address{Street: "Main St", City: "New York"}}
		key := fmt.Sprintf("%s_%d", record.Name, record.Age)
		expectedMap[key] = record
	}
	return expectedMap
}

func BenchmarkReadCSV_Concurrency(b *testing.B) {
	concurrencyValues := []int32{1, 5, 10, 20, 80}
	for _, concurrency := range concurrencyValues {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			runBenchmarkWithConcurrency(b, concurrency)
		})
	}
}

func runBenchmarkWithConcurrency(b *testing.B, concurrency int32) {
	t := (*testing.T)(unsafe.Pointer(b))
	for i := 0; i < b.N; i++ {
		// Create a large CSV data
		csvData := "name,age,address_street,address_city\n"
		for i := 0; i < 10000; i++ {
			csvData += "John," + strconv.Itoa(i) + ",Main St,New York\n"
		}

		// Create a temporary CSV file
		csvFilePath := createTempFile(t, csvData)
		defer os.Remove(csvFilePath) // Clean up

		// Set up a handler function
		mx := sync.Mutex{}
		actualMap := make(map[string]*Person)
		handler := func(record interface{}) error {
			rc := record.(*Person)
			key := fmt.Sprintf("%s_%d", rc.Name, rc.Age)
			mx.Lock()
			actualMap[key] = rc
			mx.Unlock()
			return nil
		}

		// Start measuring time
		startTime := time.Now()

		// Read the CSV data concurrently
		err := ReadCSV(csvFilePath, &Person{}, WithHandler(handler), WithConcurrency(concurrency))
		if err != nil {
			b.Fatalf("error reading CSV: %v", err)
		}

		// Stop measuring time
		elapsedTime := time.Since(startTime)
		b.Logf("Elapsed time with concurrency %d: %s", concurrency, elapsedTime)

		// Compare expected and actual records
		expectedMap := createExpectedRecords()
		if !reflect.DeepEqual(expectedMap, actualMap) {
			b.Errorf("pointer to object records mismatch\nExpected: %v\nGot: %v", expectedMap, actualMap)
		}
	}
}
