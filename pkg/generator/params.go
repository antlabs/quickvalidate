package generator

import (
	"fmt"
	"strconv"
	"time"
)

// The helpers below mirror go-playground/validator's asInt/asUint/asFloat/asBool
// family, but run at generation time so that the emitted code holds literals.

func paramInt(param string) (int64, error) {
	i, err := strconv.ParseInt(param, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid parameter %q: expected an integer", param)
	}
	return i, nil
}

func paramUint(param string) (uint64, error) {
	i, err := strconv.ParseUint(param, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid parameter %q: expected an unsigned integer", param)
	}
	return i, nil
}

func paramFloat(param string, bits int) (float64, error) {
	f, err := strconv.ParseFloat(param, bits)
	if err != nil {
		return 0, fmt.Errorf("invalid parameter %q: expected a number", param)
	}
	return f, nil
}

func paramBool(param string) (bool, error) {
	b, err := strconv.ParseBool(param)
	if err != nil {
		return false, fmt.Errorf("invalid parameter %q: expected a boolean", param)
	}
	return b, nil
}

// paramDuration mirrors asIntFromTimeDuration: a duration string is parsed as
// nanoseconds, anything else falls back to an integer.
func paramDuration(param string) (int64, error) {
	d, err := time.ParseDuration(param)
	if err != nil {
		return paramInt(param)
	}
	return int64(d), nil
}

func floatLiteral(v float64, bits int) string {
	return strconv.FormatFloat(v, 'g', -1, bits)
}
