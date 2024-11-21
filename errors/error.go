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
	// errorDetailDelimiter is the string delimiting the error detail keys and
	// values.
	errorDetailDelimiter = "'"

	// errorDetailKeyValueSeparator is the string separating error detail keys
	// from their values.
	errorDetailKeyValueSeparator = "="

	// errorDetailSeparator is the string separating messages from details and
	// details from each other.
	errorDetailSeparator = ", "
)

// CompoundUnwrapper is an interface describing a compound error combining
// multiple errors and implementing a multiple error unwrapping interface.
type CompoundUnwrapper interface {
	Unwrap() []error
}

// formatDetails returns a string encoding the provided details key value pair
// slice for an error in the commonly used format.
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

// New returns a created error object encoding the specified message and the
// provided details.
//
// When no details are provided New is just a proxy function to the standard
// errors package's similarly named function, otherwise it will use the standard
// fmt package's Errorf function to create the error.
func New(message string, detailKeyValuePairs ...any) error {
	if len(detailKeyValuePairs) == 0 {
		return errors.New(message)
	}

	return fmt.Errorf("%s, %s", message, formatDetails(detailKeyValuePairs...))
}

// UnwrapCompound returns a slice of errors unwrapped from the specified
// compound error, returns a slice of a single error in case of a non-compound
// error being specified and returns nil if the specified error or all the
// unwrapped errors are nil.
func UnwrapCompound(err error) []error {
	if err == nil {
		return nil
	}

	compoundErr, ok := err.(CompoundUnwrapper)
	if !ok {
		return []error{err} // Note: the inspiration for this comes from https://pkg.go.dev/emperror.dev/errors#GetErrors.
	}

	return compoundErr.Unwrap()
}

// Wrap returns a new error after wrapping the specified one with the provided
// message and given details.
//
// If the original error is nil then wrap returns nil as well.
//
// If the message is empty and no details are provided then the original error
// is returned as is.
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
