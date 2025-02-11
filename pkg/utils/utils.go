package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ovh/utask"

	"github.com/juju/errors"
)

// ValidString asserts that a string is within the minimum and maximum length configured for µtask
func ValidString(field, value string) error {
	if len(value) < utask.MinTextSize {
		return errors.NotValidf("%s can't be shorter than %d characters", field, utask.MinTextSize)
	}
	if len(value) > utask.MaxTextSize {
		return errors.NotValidf("%s can't be longer than %d characters", field, utask.MaxTextSize)
	}
	return nil
}

// ValidText asserts that a long text string is within the minimum and maximum length configured for µtask
func ValidText(field, value string) error {
	if len(value) < utask.MinTextSize {
		return errors.NotValidf("%s can't be shorter than %d characters", field, utask.MinTextSize)
	}
	if len(value) > utask.MaxTextSizeLong {
		return errors.NotValidf("%s can't be longer than %d characters", field, utask.MaxTextSizeLong)
	}
	return nil
}

// NormalizeName trims leading and trailing spaces on a string, and converts its characters to lowercase
func NormalizeName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ListContainsString asserts that a string slice contains a given string
func ListContainsString(list []string, item string) bool {
	if list != nil {
		for _, i := range list {
			if i == item {
				return true
			}
		}
	}
	return false
}

// PrintJSON prints out an interface{} as an indented json string
func PrintJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "   ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}

// StrPtr returns the pointer to a string's value
func StrPtr(s string) *string {
	return &s
}

// JSONnumberUnmarshal unmarshals a json string with numbers cast as json.Number, not float64 (to avoid scientific notation on large IDs)
func JSONnumberUnmarshal(r io.Reader, i interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(i)
}

// ConvertJSONRowToSlice takes a json-formatted array and returns a string slice
func ConvertJSONRowToSlice(in string) ([]string, error) {
	var tmpslice []string
	err := json.Unmarshal([]byte(in), &tmpslice)
	return tmpslice, err
}
