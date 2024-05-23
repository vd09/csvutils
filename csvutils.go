// Package csvutils provides utilities for reading and writing CSV files with struct support.
package csvutils

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

// WriteCSV writes a slice of structs to a CSV file at the specified filePath.
func WriteCSV(filePath string, records []interface{}) error {
	if len(records) == 0 {
		return errors.New("no records to write")
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	elemType := reflect.TypeOf(records[0])
	if elemType.Kind() != reflect.Struct {
		return errors.New("records elements must be struct")
	}

	header := make([]string, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		header[i] = field.Tag.Get("csv")
		if header[i] == "" {
			header[i] = field.Name
		}
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, record := range records {
		v := reflect.ValueOf(record)
		recordData := make([]string, v.NumField())
		for j := 0; j < v.NumField(); j++ {
			recordData[j] = fmt.Sprintf("%v", v.Field(j).Interface())
		}
		if err := writer.Write(recordData); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// RecordHandler is a type for functions that handle a single CSV record.
type RecordHandler func(record interface{}) error

// ReadCSV reads a CSV file and converts each record to the specified struct type.
func ReadCSV(filePath string, recordType interface{}, handler RecordHandler) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	columnIndex := make(map[string]int, len(headers))
	for i, header := range headers {
		columnIndex[header] = i
	}

	elemType := reflect.TypeOf(recordType)
	if elemType.Kind() != reflect.Ptr || elemType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("recordType must be a pointer to a struct")
	}
	elemType = elemType.Elem()

	fieldInfo := make([]struct {
		index  int
		setter func(reflect.Value, string) error
	}, elemType.NumField())

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			csvTag = field.Name
		}
		index, ok := columnIndex[csvTag]
		if !ok {
			return fmt.Errorf("missing CSV column: %s", csvTag)
		}
		setter, err := getFieldSetter(field.Type)
		if err != nil {
			return fmt.Errorf("unsupported field type for field %s: %w", field.Name, err)
		}
		fieldInfo[i] = struct {
			index  int
			setter func(reflect.Value, string) error
		}{index: index, setter: setter}
	}

	for recordNum := 1; ; recordNum++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read record at line %d: %w", recordNum, err)
		}

		recordValue := reflect.New(elemType).Elem()

		for i := 0; i < elemType.NumField(); i++ {
			fieldValue := recordValue.Field(i)
			value := record[fieldInfo[i].index]
			if err := fieldInfo[i].setter(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field value for field %s at line %d: %w", elemType.Field(i).Name, recordNum, err)
			}
		}

		if err := handler(recordValue.Addr().Interface()); err != nil {
			return fmt.Errorf("handler error at line %d: %w", recordNum, err)
		}
	}

	return nil
}

// getFieldSetter returns a function to set a field value based on its type.
func getFieldSetter(fieldType reflect.Type) (func(reflect.Value, string) error, error) {
	switch fieldType.Kind() {
	case reflect.String:
		return func(v reflect.Value, s string) error {
			v.SetString(s)
			return nil
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value, s string) error {
			intValue, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing int value %s: %w", s, err)
			}
			v.SetInt(intValue)
			return nil
		}, nil
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value, s string) error {
			floatValue, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("error parsing float value %s: %w", s, err)
			}
			v.SetFloat(floatValue)
			return nil
		}, nil
	case reflect.Bool:
		return func(v reflect.Value, s string) error {
			boolValue, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("error parsing bool value %s: %w", s, err)
			}
			v.SetBool(boolValue)
			return nil
		}, nil
	}
	return nil, fmt.Errorf("unsupported field type: %v", fieldType)
}
