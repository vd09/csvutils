package csvutils_test

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/vd09/csvutils" // import your package path here
)

// Define a sample struct for testing
type TestStruct struct {
	Name  string `csv:"Name"`
	Age   int    `csv:"Age"`
	Email string `csv:"Email"`
}

func TestWriteCSV(t *testing.T) {
	// Create a temporary file for writing CSV data
	filePath := "test.csv"
	defer os.Remove(filePath)

	// Create some sample data
	records := []interface{}{
		TestStruct{Name: "Alice", Age: 30, Email: "alice@example.com"},
		TestStruct{Name: "Bob", Age: 35, Email: "bob@example.com"},
	}

	// Call WriteCSV function with the file path
	err := csvutils.WriteCSV(filePath, records)
	if err != nil {
		t.Fatalf("WriteCSV returned error: %v", err)
	}

	// Open the CSV file for reading
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// Parse the CSV data from the file
	reader := csv.NewReader(file)
	parsedRecords, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Error reading CSV data: %v", err)
	}

	// Validate the parsed records
	expected := [][]string{
		{"Name", "Age", "Email"},
		{"Alice", "30", "alice@example.com"},
		{"Bob", "35", "bob@example.com"},
	}

	if !reflect.DeepEqual(parsedRecords, expected) {
		t.Errorf("Parsed records do not match expected. Got: %v, Expected: %v", parsedRecords, expected)
	}
}

func TestReadCSV(t *testing.T) {
	// Sample CSV data
	data := []byte(`Name,Age,Email
Alice,30,alice@example.com
Bob,35,bob@example.com`)

	// Write data to a temporary CSV file
	filePath := "test.csv"
	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		t.Fatalf("Error writing test data to file: %v", err)
	}
	defer os.Remove(filePath)

	// Create variables to store records
	var records []TestStruct

	// Define a handler function to process each record
	handler := func(record interface{}) error {
		// Convert interface{} to TestStruct
		rec, ok := record.(*TestStruct)
		if !ok {
			t.Fatalf("Invalid type for record: %T", record)
		}
		// Append the record to the slice
		records = append(records, *rec)
		return nil
	}

	// Call ReadCSV function with the file path
	err = csvutils.ReadCSV(filePath, &TestStruct{}, csvutils.WithHandler(handler))
	if err != nil {
		t.Fatalf("ReadCSV returned error: %v", err)
	}

	// Define expected records
	expected := []TestStruct{
		{Name: "Alice", Age: 30, Email: "alice@example.com"},
		{Name: "Bob", Age: 35, Email: "bob@example.com"},
	}

	// Validate the parsed records
	if !reflect.DeepEqual(records, expected) {
		t.Errorf("Parsed records do not match expected. Got: %v, Expected: %v", records, expected)
	}
}

func TestReadCSV_DefaultForNoData(t *testing.T) {
	// Sample CSV data
	data := []byte(`Name,Age,Email
Alice,,alice@example.com
Bob,35,`)

	// Write data to a temporary CSV file
	filePath := "test.csv"
	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		t.Fatalf("Error writing test data to file: %v", err)
	}
	defer os.Remove(filePath)

	// Create variables to store records
	var records []TestStruct

	// Define a handler function to process each record
	handler := func(record interface{}) error {
		// Convert interface{} to TestStruct
		rec, ok := record.(*TestStruct)
		if !ok {
			t.Fatalf("Invalid type for record: %T", record)
		}
		// Append the record to the slice
		records = append(records, *rec)
		return nil
	}

	// Call ReadCSV function with the file path
	err = csvutils.ReadCSV(filePath, &TestStruct{}, csvutils.WithHandler(handler))
	if err != nil {
		t.Fatalf("ReadCSV returned error: %v", err)
	}

	// Define expected records
	expected := []TestStruct{
		{Name: "Alice", Age: 0, Email: "alice@example.com"},
		{Name: "Bob", Age: 35, Email: ""},
	}

	// Validate the parsed records
	if !reflect.DeepEqual(records, expected) {
		t.Errorf("Parsed records do not match expected. Got: %v, Expected: %v", records, expected)
	}
}
