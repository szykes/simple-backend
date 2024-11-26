package errors

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	As     = errors.As
	Is     = errors.Is
	Join   = errors.Join
	Unwrap = errors.Unwrap
)

const (
	errorDetailDelimiter         = "'"
	errorDetailKeyValueSeparator = "="
	errorDetailSeparator         = ", "
)

func formatDetails(detailKeyValuePairs ...any) string {
	builder := strings.Builder{}
	for i := 0; i < len(detailKeyValuePairs); i += 2 {
		if i != 0 {
			builder.WriteString(errorDetailSeparator)
		}

		key := detailKeyValuePairs[i]
		builder.WriteString(
			fmt.Sprintf(errorDetailDelimiter+"%+v"+errorDetailDelimiter+errorDetailKeyValueSeparator, key),
		)

		if i+1 >= len(detailKeyValuePairs) {
			builder.WriteString(errorDetailDelimiter + errorDetailDelimiter)
		} else {
			value := detailKeyValuePairs[i+1]
			formatString := "%#v"
			reflectValue := reflect.ValueOf(value)

			if reflectValue.Kind() == reflect.Pointer {
				formatString = "(" + reflectValue.Type().String() + ")%#v" // Note: encoding pointer to (fulltype)value.
			}

			for reflectValue.Kind() == reflect.Pointer { // Note: dereferencing pointers recursively for their value.
				reflectValue = reflectValue.Elem()
				value = reflectValue.Interface()
			}

			builder.WriteString(fmt.Sprintf(errorDetailDelimiter+formatString+errorDetailDelimiter, value))
		}
	}

	return builder.String()
}

func New(message string, detailKeyValuePairs ...any) error {
	if len(detailKeyValuePairs) == 0 {
		return errors.New(message)
	}

	return fmt.Errorf("%s, %s", message, formatDetails(detailKeyValuePairs...))
}

func Wrap(err error, message string, detailKeyValuePairs ...any) error {
	if err == nil {
		return nil
	} else if message == "" &&
		len(detailKeyValuePairs) == 0 {
		return err
	}

	if len(detailKeyValuePairs) == 0 {
		return fmt.Errorf(
			"%s: %w",
			message, err,
		)
	}

	return fmt.Errorf(
		"%s: %w, %s",
		message, err, formatDetails(detailKeyValuePairs...),
	)
}
