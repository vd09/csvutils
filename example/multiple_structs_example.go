package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vd09/csvutils"
)

type User struct {
	Name    string  `csv:"name"`
	Age     int     `csv:"age"`
	Email   string  `csv:"email"`
	Address Address `csv:"address"`
}

type Address struct {
	PinCode  string `csv:"pin_code"`
	Address1 string `csv:"address1"`
	Address2 string `csv:"address2"`
}

func main() {
	// Example of writing CSV
	user := []interface{}{
		User{
			Name:  "John Doe",
			Age:   30,
			Email: "john.doe@example.com",
			Address: Address{
				PinCode:  "123456",
				Address1: "123 Main St",
				Address2: "Apt 4B",
			},
		},
		// Add more records as needed
	}
	fileName := "./example/User.csv"
	os.Remove(fileName)

	err := csvutils.WriteCSV(fileName, user)
	if err != nil {
		log.Fatalf("Failed to write CSV: %v", err)
	}

	// Example of reading CSV
	err = csvutils.ReadCSV(fileName, &User{}, func(record interface{}) error {
		user := record.(*User)
		fmt.Printf("Read record: %+v\n", user)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}
}
