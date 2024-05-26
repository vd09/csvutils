package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vd09/csvutils"
)

type UserPtrEx struct {
	Name    string        `csv:"name"`
	Age     int           `csv:"age"`
	Email   string        `csv:"email"`
	Address *AddressPtrEx `csv:"address"`
}

type AddressPtrEx struct {
	PinCode  string `csv:"pin_code"`
	Address1 string `csv:"address1"`
	Address2 string `csv:"address2"`
}

func main() {
	fileName := "./example/User.csv"

	os.Remove(fileName)
	// Example of writing CSV
	user := []interface{}{
		UserPtrEx{
			Name:  "John Doe",
			Age:   30,
			Email: "john.doe@example.com",
			//Address: nil,
			Address: &AddressPtrEx{
				PinCode:  "123456",
				Address1: "123 Main St",
				Address2: "Apt 4B",
			},
		},
		// Add more records as needed
	}

	err := csvutils.WriteCSV(fileName, user)
	if err != nil {
		log.Fatalf("Failed to write CSV: %v", err)
	}

	// Example of reading CSV
	err = csvutils.ReadCSV(fileName, &UserPtrEx{}, func(record interface{}) error {
		user := record.(*UserPtrEx)
		fmt.Printf("Read record: %+v %+v\n", user, *user.Address)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}
}
