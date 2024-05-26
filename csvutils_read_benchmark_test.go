package csvutils

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"
)

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
