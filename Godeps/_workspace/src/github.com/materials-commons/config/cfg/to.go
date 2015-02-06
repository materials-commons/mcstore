package cfg

import (
	"fmt"
	"github.com/spf13/cast"
	"strconv"
	"time"
)

// ToTime converts in to a time.Time value.
func ToTime(in interface{}) (time.Time, error) {
	switch val := in.(type) {
	case int64:
		return time.Unix(val, 0), nil
	default:
		t, err := cast.ToTimeE(val)
		if err != nil {
			return t, ErrBadType
		}
		return t, nil
	}
}

// ToBool converts in to a bool value.
func ToBool(in interface{}) (bool, error) {
	switch val := in.(type) {
	case string:
		sval, err := strconv.ParseBool(val)
		if err != nil {
			return false, ErrBadType
		}
		return sval, nil
	default:
		value, err := cast.ToBoolE(val)
		if err != nil {
			return false, ErrBadType
		}
		return value, nil
	}
}

// ToInt converts in to a int value.
func ToInt(in interface{}) (int, error) {
	if val, err := cast.ToIntE(in); err == nil {
		return val, nil
	}
	return 0, ErrBadType
}

// ToInt64 converts in to an int value
func ToInt64(in interface{}) (int64, error) {
	if val, err := ToInt64E(in); err == nil {
		return val, nil
	}
	return 0, ErrBadType
}

// ToInt64E is the same as cast.ToIntE, except it returns a 64 bit int.
func ToInt64E(in interface{}) (int64, error) {
	switch s := in.(type) {
	case int:
		return int64(s), nil
	case int64:
		return s, nil
	case int32:
		return int64(s), nil
	case int16:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case string:
		v, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to Cast %#v to int", in)
		}
		return v, nil
	case float64:
		return int64(s), nil
	case bool:
		if bool(s) {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to Cast %#v to int", in)
	}
}

// ToString converts in to a string value.
func ToString(in interface{}) (string, error) {
	if val, err := cast.ToStringE(in); err == nil {
		return val, nil
	}
	return "", ErrBadType
}
