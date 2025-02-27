package validators

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// Common regular expressions for validation
	emailRegex    = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	ipv4Regex     = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$`)
	ipv6Regex     = regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRegex  = regexp.MustCompile(`^[0-9]+$`)
	hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	rgbRegex      = regexp.MustCompile(`^rgb\(\s*(?:(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])|(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])%\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])%\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])%)\s*\)$`)
	rgbaRegex     = regexp.MustCompile(`^rgba\(\s*(?:(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])|(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])%\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])%\s*,\s*(?:0|[1-9]\d?|1\d\d?|2[0-4]\d|25[0-5])%)\s*,\s*(?:(?:0.[1-9]*)|[01])\s*\)$`)
	e164Regex     = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
)

// IsEmail validates if the string is a valid email address
func IsEmail(email string) bool {
	if len(email) < 6 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// IsURL validates if the string is a valid URL
func IsURL(str string) bool {
	if str == "" {
		return false
	}
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsIP validates if the string is a valid IP address
func IsIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsIPv4 validates if the string is a valid IPv4 address
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}
	return ipv4Regex.MatchString(str)
}

// IsIPv6 validates if the string is a valid IPv6 address
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	if ip == nil {
		return false
	}
	return ipv6Regex.MatchString(str)
}

// IsHostname validates if the string is a valid hostname
func IsHostname(str string) bool {
	if len(str) > 255 {
		return false
	}
	return hostnameRegex.MatchString(str)
}

// IsUUID validates if the string is a valid UUID
func IsUUID(str string) bool {
	return uuidRegex.MatchString(strings.ToLower(str))
}

// IsAlpha validates if the string contains only alpha characters
func IsAlpha(str string) bool {
	return alphaRegex.MatchString(str)
}

// IsAlphanumeric validates if the string contains only alphanumeric characters
func IsAlphanumeric(str string) bool {
	return alphanumRegex.MatchString(str)
}

// IsNumeric validates if the string contains only numeric characters
func IsNumeric(str string) bool {
	return numericRegex.MatchString(str)
}

// IsHexColor validates if the string is a valid hex color
func IsHexColor(str string) bool {
	return hexColorRegex.MatchString(str)
}

// IsRGB validates if the string is a valid RGB color
func IsRGB(str string) bool {
	return rgbRegex.MatchString(str)
}

// IsRGBA validates if the string is a valid RGBA color
func IsRGBA(str string) bool {
	return rgbaRegex.MatchString(str)
}

// IsColor validates if the string is a valid color (hex, rgb, or rgba)
func IsColor(str string) bool {
	return IsHexColor(str) || IsRGB(str) || IsRGBA(str)
}

// IsE164 validates if the string is a valid E.164 formatted phone number
func IsE164(str string) bool {
	return e164Regex.MatchString(str)
}

// IsOneOf validates if the string is one of the allowed values
func IsOneOf(str string, allowedValues []string) bool {
	for _, v := range allowedValues {
		if str == v {
			return true
		}
	}
	return false
}

// IsOneOfCaseInsensitive validates if the string is one of the allowed values (case insensitive)
func IsOneOfCaseInsensitive(str string, allowedValues []string) bool {
	strLower := strings.ToLower(str)
	for _, v := range allowedValues {
		if strLower == strings.ToLower(v) {
			return true
		}
	}
	return false
}

// IsLengthLessThan validates if the string length is less than max
func IsLengthLessThan(str string, max int) bool {
	return utf8.RuneCountInString(str) < max
}

// IsLengthGreaterThan validates if the string length is greater than min
func IsLengthGreaterThan(str string, min int) bool {
	return utf8.RuneCountInString(str) > min
}

// IsLengthBetween validates if the string length is between min and max
func IsLengthBetween(str string, min, max int) bool {
	count := utf8.RuneCountInString(str)
	return count >= min && count <= max
}

// IsLengthEqual validates if the string length is equal to length
func IsLengthEqual(str string, length int) bool {
	return utf8.RuneCountInString(str) == length
}

// ParseTime parses a time string in multiple formats
func ParseTime(str string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time: %s", str)
}

// IsDatetime validates if the string is a valid datetime
func IsDatetime(str string) bool {
	_, err := ParseTime(str)
	return err == nil
}

// IsJSON validates if the string is valid JSON
func IsJSON(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

// ContainsString validates if the string contains the substring
func ContainsString(str, substr string) bool {
	return strings.Contains(str, substr)
}

// StartsWith validates if the string starts with the prefix
func StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith validates if the string ends with the suffix
func EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}
