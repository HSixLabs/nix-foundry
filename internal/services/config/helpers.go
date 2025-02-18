package config

import (
	"fmt"
	"reflect"
	"strconv"
)

// getNestedValue retrieves a value from a nested structure using dot notation
func getNestedValue(obj interface{}, path []string) (interface{}, error) {
	current := reflect.ValueOf(obj)

	for _, field := range path {
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return nil, fmt.Errorf("invalid path: %s is not a struct", field)
		}

		current = current.FieldByName(field)
		if !current.IsValid() {
			return nil, fmt.Errorf("invalid field: %s", field)
		}
	}

	return current.Interface(), nil
}

// Add lint ignore comment for intentionally kept unused function
//lint:ignore U1000 This is used via reflection in other packages
func setNestedValue(obj interface{}, path []string, value string) error {
	current := reflect.ValueOf(obj)
	if current.Kind() == reflect.Ptr {
		current = current.Elem()
	}

	// Navigate to the parent of the field we want to set
	for i := 0; i < len(path)-1; i++ {
		if current.Kind() != reflect.Struct {
			return fmt.Errorf("invalid path: %s is not a struct", path[i])
		}

		current = current.FieldByName(path[i])
		if !current.IsValid() {
			return fmt.Errorf("invalid field: %s", path[i])
		}
	}

	// Get the field we want to set
	field := current.FieldByName(path[len(path)-1])
	if !field.IsValid() {
		return fmt.Errorf("invalid field: %s", path[len(path)-1])
	}

	// Convert and set the value based on the field type
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		field.SetInt(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.SetBool(v)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}
