package main

import (
	"fmt"
	"log"

	"github.com/vd09/csvutils"
)

// Person defines the structure for the CSV records.
type Person struct {
	Name  string `csv:"name"`
	Age   int    `csv:"age"`
	Email string `csv:"email"`
}

func main() {
	// Create some sample data.
	people := []interface{}{
		Person{Name: "Alice", Age: 30, Email: "alice@example.com"},
		Person{Name: "Bob", Age: 25, Email: "bob@example.com"},
		Person{Name: "David", Email: "David@example.com"},
		Person{Name: "John", Age: 20},
	}

	fileName := "./example/people.csv"
	// Write the sample data to a CSV file.
	err := csvutils.WriteCSV(fileName, people)
	if err != nil {
		log.Fatalf("failed to write CSV: %v", err)
	}
	fmt.Println("CSV file written successfully")

	// Handler function to process each record read from the CSV.
	handler := func(record interface{}) error {
		person := record.(*Person)
		fmt.Printf("Read record: Name=%s, Age=%d, Email=%s\n", person.Name, person.Age, person.Email)
		return nil
	}

	// Read the data back from the CSV file.
	err = csvutils.ReadCSV(fileName, &Person{}, csvutils.WithHandler(handler))
	if err != nil {
		log.Fatalf("failed to read CSV: %v", err)
	}
	fmt.Println("CSV file read successfully")
}
