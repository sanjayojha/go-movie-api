package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int

// Implement a MarshalJSON() method on the Runtime type so that it satisfies the json.Marshaler interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it in double quotes. It needs to be surrounded by double quotes in order to be a valid JSON string.
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// Implement a UnmarshalJSON() method on the Runtime type so that it satisfies the json.Unmarshaler interface. IMPORTANT: Because UnmarshalJSON() needs to modify the receiver (our Runtime type), we must use a pointer receiver for this to work correctly. Otherwise, we will only be modifying a copy (which is then discarded when this method returns).

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	errInvalidRuntimeFormat := errors.New("invalid runtime format")

	// We expect that the incoming JSON value will be a string in the format  "<runtime> mins", and the first thing we need to do is remove the surrounding  double quotes from this string. If we can't unquote it, then we return the  errInvalidRuntimeFormat error.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return errInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number.
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format. If it isn't, we return the errInvalidRuntimeFormat error again.
	if len(parts) != 2 || parts[1] != "mins" {
		return errInvalidRuntimeFormat
	}

	// Otherwise, parse the string containing the number into an int. Again, if this fails return the ErrInvalidRuntimeFormat error.
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return errInvalidRuntimeFormat
	}

	// Convert the int to a Runtime type and assign this to the receiver. Note that we use the * operator to dereference the receiver (which is a pointer to a Runtime type) in order to set the underlying value of the pointer.
	*r = Runtime(i)

	return nil
}
