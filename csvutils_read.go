package csvutils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	"github.com/vd09/gr_worker/worker_pool"
)

type RecordHandler func(interface{}) error

type csvOptions struct {
	handler     RecordHandler
	concurrency int32
}

func WithHandler(handler RecordHandler) func(*csvOptions) {
	return func(opts *csvOptions) {
		opts.handler = handler
	}
}

func WithConcurrency(concurrency int32) func(*csvOptions) {
	return func(opts *csvOptions) {
		opts.concurrency = concurrency
	}
}

func newCsvOptions(options []func(*csvOptions)) *csvOptions {
	opts := &csvOptions{
		handler:     nil,
		concurrency: 1,
	}

	for _, option := range options {
		option(opts)
	}
	return opts
}

func ReadCSV(filePath string, recordType interface{}, options ...func(*csvOptions)) error {
	csvOptions := newCsvOptions(options)

	pool, err := worker_pool.NewWorkerPoolAdapter(
		worker_pool.WithMaxWorkers(csvOptions.concurrency),
		worker_pool.WithMinWorkers(csvOptions.concurrency),
	)
	if err != nil {
		return fmt.Errorf("failed to create worker pool: %w", err)
	}
	defer pool.WaitAndStop()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	columnIndex := make(map[string]int, len(headers))
	for i, header := range headers {
		columnIndex[header] = i
	}

	elemType := reflect.TypeOf(recordType).Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("recordType must be a pointer to a struct")
	}
	fieldInfo, err := buildFieldInfo(elemType, columnIndex, "", []int{})
	if err != nil {
		return fmt.Errorf("failed to build field info: %w", err)
	}

	recordNum := 1
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read record at line %d: %w", recordNum, err)
		}
		pool.AddTask(processRecord, record, elemType, fieldInfo, csvOptions.handler)
		recordNum++
	}
	return nil
}

func processRecord(record []string, elemType reflect.Type, fieldInfo []fieldInfo, handler RecordHandler) error {
	recordValue := reflect.New(elemType).Elem()
	initNestedPointers(recordValue)

	for _, info := range fieldInfo {
		fieldValue := recordValue.FieldByIndex(info.index)
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem()
		}
		var value string
		if info.columnIndex >= 0 && info.columnIndex < len(record) {
			value = record[info.columnIndex]
		}
		if value == "" {
			value = info.defaultValue
		}
		if err := info.setter(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field value for field %s: %w", info.fieldName, err)
		}
	}
	if handler != nil {
		if err := handler(recordValue.Addr().Interface()); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}
	return nil
}

func initNestedPointers(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			initNestedPointers(v.Field(i))
		}
	}
}

func buildFieldInfo(elemType reflect.Type, columnIndex map[string]int, parentTag string, parentFieldIndex []int) ([]fieldInfo, error) {
	var fieldInfos []fieldInfo
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			csvTag = field.Name
		}
		if parentTag != "" {
			csvTag = parentTag + "_" + csvTag
		}
		newFieldIndex := append(parentFieldIndex, field.Index...)

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			nestedFieldInfos, err := buildFieldInfo(fieldType, columnIndex, csvTag, newFieldIndex)
			if err != nil {
				return nil, err
			}
			fieldInfos = append(fieldInfos, nestedFieldInfos...)
		} else {
			index, ok := columnIndex[csvTag]
			defaultValue := field.Tag.Get("default")
			if !ok {
				//if defaultValue == "" {
				//	return nil, fmt.Errorf("missing CSV column: %s", csvTag)
				//}
				index = -1 // Indicate that the column is missing and should use the default value
			}
			setter, err := getFieldSetter(field.Type)
			if err != nil {
				return nil, fmt.Errorf("unsupported field type for field %s: %w", field.Name, err)
			}
			fieldInfos = append(fieldInfos, fieldInfo{
				fieldName:    field.Name,
				index:        newFieldIndex,
				columnIndex:  index,
				setter:       setter,
				defaultValue: defaultValue,
			})
		}
	}
	return fieldInfos, nil
}

type fieldInfo struct {
	fieldName    string
	index        []int
	columnIndex  int
	setter       func(reflect.Value, string) error
	defaultValue string
}

func getFieldSetter(fieldType reflect.Type) (func(reflect.Value, string) error, error) {
	switch fieldType.Kind() {
	case reflect.String:
		return func(v reflect.Value, s string) error { v.SetString(s); return nil }, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value, s string) error {
			if s == "" {
				s = "0"
			}
			intValue, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing int value %s: %w", s, err)
			}
			v.SetInt(intValue)
			return nil
		}, nil
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value, s string) error {
			if s == "" {
				s = "0"
			}
			floatValue, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("error parsing float value %s: %w", s, err)
			}
			v.SetFloat(floatValue)
			return nil
		}, nil
	case reflect.Bool:
		return func(v reflect.Value, s string) error {
			if s == "" {
				s = "false"
			}
			boolValue, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("error parsing bool value %s: %w", s, err)
			}
			v.SetBool(boolValue)
			return nil
		}, nil
	default:
		return nil, fmt.Errorf("unsupported field type: %v", fieldType)
	}
}
