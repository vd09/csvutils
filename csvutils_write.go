package csvutils

import (
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
)

// WriteCSV writes a slice of structs to a CSV file at the specified filePath.
func WriteCSV[T any](filePath string, records []T) error {
	if len(records) == 0 {
		return errors.New("no records to write")
	}

	file, err := openOrCreateFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	elemType := reflect.TypeOf(records[0])
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return errors.New("records elements must be struct")
	}

	headers, err := extractHeaders(elemType, "")
	if err != nil {
		return fmt.Errorf("failed to extract headers: %w", err)
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, record := range records {
		recordValue := reflect.ValueOf(record)
		if recordValue.Kind() == reflect.Ptr {
			recordValue = recordValue.Elem()
		}
		recordValues, err := extractValues(recordValue)
		if err != nil {
			return fmt.Errorf("failed to extract values: %w", err)
		}
		if err := writer.Write(recordValues); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// extractHeaders extracts CSV headers from a struct type, including nested structs.
func extractHeaders(t reflect.Type, prefix string) ([]string, error) {
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			csvTag = field.Name
		}
		headerName := prefix + csvTag
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			nestedHeaders, err := extractHeaders(fieldType, headerName+"_")
			if err != nil {
				return nil, err
			}
			headers = append(headers, nestedHeaders...)
		} else {
			headers = append(headers, headerName)
		}
	}
	return headers, nil
}

// extractValues extracts field values from a struct, including nested structs.
func extractValues(v reflect.Value) ([]string, error) {
	var values []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				// Handle nil pointer by adding empty values for each field in the struct
				fieldType := field.Type().Elem()
				for j := 0; j < fieldType.NumField(); j++ {
					values = append(values, "")
				}
				continue
			}
			field = field.Elem()
		}
		if field.Kind() == reflect.Struct {
			nestedValues, err := extractValues(field)
			if err != nil {
				return nil, err
			}
			values = append(values, nestedValues...)
		} else {
			values = append(values, fmt.Sprintf("%v", field.Interface()))
		}
	}
	return values, nil
}
