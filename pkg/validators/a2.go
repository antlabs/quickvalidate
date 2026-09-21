// Ported from github.com/go-playground/validator/v10 v10.25.0 (baked_in.go).
// Copyright (c) 2015 Dean Karn. MIT License — see LICENSE.
//
// Concrete-typed, reflect-free equivalents of the baked-in validators. Every
// reference function takes a validator.FieldLevel and reads exactly one string
// out of it (fl.Field().String()), so each one is ported here as a plain
// `string` parameter. Semantics, boundaries, empty-string handling and length
// limits are identical to the reference; the regular expressions live in
// regexes.go and are shared with the rest of the package.

package validators

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// fieldMatchesRegexByStringerValOrString is the port of the reference helper of
// the same name (v10.25.0 util.go).
//
// The reference switches on fl.Field().Kind(): for reflect.String it matches the
// field's string value, for any other kind it first tries the value's
// fmt.Stringer form and falls back to fl.Field().String(). With a concrete
// string input only the reflect.String branch remains reachable, so that is all
// this port does. (The reference uses it for the UUID family; it is kept here so
// porting that family needs no extra helper.)
func fieldMatchesRegexByStringerValOrString(regexFn func() *regexp.Regexp, s string) bool {
	return regexFn().MatchString(s)
}

// IsEmail validates if s is a valid email address (reference: isEmail, line 1697).
//
// NOTE: v10.25.0 performs no length check here — the value is matched against
// the (very permissive, RFC 5322 style) emailRegex only. Callers that want the
// 6..254 guard seen in some other libraries must add it themselves.
func IsEmail(s string) bool {
	return emailRegex().MatchString(s)
}

// IsAlpha validates if s contains only alphabetic (a-zA-Z) characters
// (reference: isAlpha, line 1757).
func IsAlpha(s string) bool {
	return alphaRegex().MatchString(s)
}

// IsAlphanum validates if s contains only alphanumeric (a-zA-Z0-9) characters
// (reference: isAlphanum, line 1752).
func IsAlphanum(s string) bool {
	return alphaNumericRegex().MatchString(s)
}

// IsAlphaUnicode validates if s contains only unicode letters (\p{L})
// (reference: isAlphaUnicode, line 1767).
func IsAlphaUnicode(s string) bool {
	return alphaUnicodeRegex().MatchString(s)
}

// IsAlphanumUnicode validates if s contains only unicode letters and numbers
// (\p{L}\p{N}) (reference: isAlphanumUnicode, line 1762).
func IsAlphanumUnicode(s string) bool {
	return alphaUnicodeNumericRegex().MatchString(s)
}

// IsNumeric validates if s is a valid numeric value, i.e. an optionally signed
// decimal number with an optional fractional part (reference: isNumeric, line 1742).
func IsNumeric(s string) bool {
	return numericRegex().MatchString(s)
}

// IsNumber validates if s consists of digits only (reference: isNumber, line 1732).
func IsNumber(s string) bool {
	return numberRegex().MatchString(s)
}

// IsHexadecimal validates if s is a hexadecimal number with an optional 0x/0X
// prefix (reference: isHexadecimal, line 1727).
func IsHexadecimal(s string) bool {
	return hexadecimalRegex().MatchString(s)
}

// IsHEXColor validates if s is a valid hex color, i.e. '#' followed by 3, 4, 6
// or 8 hex digits (reference: isHEXColor, line 1722).
func IsHEXColor(s string) bool {
	return hexColorRegex().MatchString(s)
}

// IsRGB validates if s is a valid rgb() color, either as three 0..255 integers
// or as three percentages (reference: isRGB, line 1717).
func IsRGB(s string) bool {
	return rgbRegex().MatchString(s)
}

// IsRGBA validates if s is a valid rgba() color: the rgb() body plus an alpha
// value of 0, 1 or 0.x (reference: isRGBA, line 1712).
func IsRGBA(s string) bool {
	return rgbaRegex().MatchString(s)
}

// IsHSL validates if s is a valid hsl() color with a 0..360 hue and 0..100%
// saturation and lightness (reference: isHSL, line 1707).
func IsHSL(s string) bool {
	return hslRegex().MatchString(s)
}

// IsHSLA validates if s is a valid hsla() color: the hsl() body plus an alpha
// value of 0, 1 or 0.x (reference: isHSLA, line 1702).
func IsHSLA(s string) bool {
	return hslaRegex().MatchString(s)
}

// IsASCII validates if s contains only ASCII characters (\x00-\x7F)
// (reference: isASCII, line 533).
func IsASCII(s string) bool {
	return aSCIIRegex().MatchString(s)
}

// IsPrintableASCII validates if s contains only printable ASCII characters
// (\x20-\x7E) (reference: isPrintableASCII, line 528).
func IsPrintableASCII(s string) bool {
	return printableASCIIRegex().MatchString(s)
}

// IsMultiByteCharacter validates if s contains at least one non-ASCII (multi
// byte) character; the empty string is considered valid (reference:
// hasMultiByteCharacter, line 517).
func IsMultiByteCharacter(s string) bool {
	if len(s) == 0 {
		return true
	}

	return multibyteRegex().MatchString(s)
}

// IsLowercase validates if s is non-empty and equal to its lower-cased self
// (reference: isLowercase, line 2757).
func IsLowercase(s string) bool {
	if s == "" {
		return false
	}

	return s == strings.ToLower(s)
}

// IsUppercase validates if s is non-empty and equal to its upper-cased self
// (reference: isUppercase, line 2771).
func IsUppercase(s string) bool {
	if s == "" {
		return false
	}

	return s == strings.ToUpper(s)
}

// IsE164 validates if s is an E.164 formatted phone number: '+' followed by an
// optional non-zero leading digit and 7 to 14 more digits (reference: isE164,
// line 1692).
func IsE164(s string) bool {
	return e164Regex().MatchString(s)
}

// IsSSN validates if s is a valid US Social Security Number; it must be exactly
// 11 bytes long before the pattern is applied (reference: isSSN, line 445).
func IsSSN(s string) bool {
	if len(s) != 11 {
		return false
	}

	return sSNRegex().MatchString(s)
}

// IsLatitude validates if s is a valid latitude coordinate in [-90, 90]
// (reference: isLatitude, line 479).
func IsLatitude(s string) bool {
	return latitudeRegex().MatchString(s)
}

// IsLongitude validates if s is a valid longitude coordinate in [-180, 180]
// (reference: isLongitude, line 456).
func IsLongitude(s string) bool {
	return longitudeRegex().MatchString(s)
}

// IsHTML validates if s contains an HTML tag (reference: isHTML, line 277).
func IsHTML(s string) bool {
	return hTMLRegex().MatchString(s)
}

// IsHTMLEncoded validates if s contains an HTML encoded entity (reference:
// isHTMLEncoded, line 273). The underlying pattern is unanchored, so a match
// anywhere in s is enough.
func IsHTMLEncoded(s string) bool {
	return hTMLEncodedRegex().MatchString(s)
}

// IsURLEncoded validates if s is percent encoded, i.e. every '%' is followed by
// two hex digits (reference: isURLEncoded, line 269).
func IsURLEncoded(s string) bool {
	return uRLEncodedRegex().MatchString(s)
}

// IsBase32 validates if s is a valid base32 string (RFC 4648 alphabet with
// padding); the empty string is valid (reference: isBase32, line 1431).
func IsBase32(s string) bool {
	return base32Regex().MatchString(s)
}

// IsBase64 validates if s is a valid base64 string (standard alphabet with '='
// padding); the empty string is valid (reference: isBase64, line 1436).
func IsBase64(s string) bool {
	return base64Regex().MatchString(s)
}

// IsBase64URL validates if s is a valid URL-safe base64 string (with '='
// padding) (reference: isBase64URL, line 1441).
func IsBase64URL(s string) bool {
	return base64URLRegex().MatchString(s)
}

// IsBase64RawURL validates if s is a valid URL-safe base64 string without '='
// padding (reference: isBase64RawURL, line 1446).
func IsBase64RawURL(s string) bool {
	return base64RawURLRegex().MatchString(s)
}

// IsDataURI validates if s is a data URI: "mediatype[;base64],<base64 payload>".
// The value is split on the first ',' and the payload must be valid base64
// (reference: isDataURI, line 502).
func IsDataURI(s string) bool {
	uri := strings.SplitN(s, ",", 2)

	if len(uri) != 2 {
		return false
	}

	if !dataURIRegex().MatchString(uri[0]) {
		return false
	}

	return base64Regex().MatchString(uri[1])
}

// IsJSON validates if s is valid JSON, using encoding/json's validity check
// (reference: isJSON, line 2704).
func IsJSON(s string) bool {
	return json.Valid([]byte(s))
}

// IsJSONBytes is the byte slice branch of the reference implementation, which
// validates the slice as it is (reference: isJSON, line 2710).
func IsJSONBytes(b []byte) bool {
	return json.Valid(b)
}

// IsJWT validates if s is a JSON Web Token, i.e. two non-empty dot separated
// base64url parts followed by a (possibly empty) third one (reference: isJWT,
// line 2724).
func IsJWT(s string) bool {
	return jWTRegex().MatchString(s)
}

// IsDatetime validates if s parses with the given Go time layout, exactly like
// the reference isDatetime (line 2785) which calls time.Parse(param, value):
// the layout is the tag parameter ("datetime=2006-01-02" -> "2006-01-02").
//
// No layouts means the same as the reference tag "datetime" without a parameter,
// where the layout is the empty string — time.Parse("", s) succeeds only for
// s == "" — and is reproduced here verbatim.
//
// Any layout accepted by time.Parse works, including timezone offset layouts
// such as time.RFC3339 / "2006-01-02T15:04:05Z07:00" and the "2006-01-02
// 15:04:05" form. The reference has no time.Duration based form here.
func IsDatetime(s string, layout ...string) bool {
	if len(layout) == 0 {
		_, err := time.Parse("", s)
		return err == nil
	}

	// The reference only ever gets one layout (the tag parameter), so for a
	// single layout this is exactly equivalent. Several layouts are accepted as
	// a convenience: s is valid if any one of them parses.
	for _, l := range layout {
		if _, err := time.Parse(l, s); err == nil {
			return true
		}
	}

	return false
}

// IsTimeZone validates if s is a time zone name known to time.LoadLocation
// (reference: isTimeZone, line 2799).
//
// Like the reference this uses the standard library only — there is no external
// timezone dependency — so the result depends on the zoneinfo database available
// at runtime (system zoneinfo, GOROOT/lib/time/zoneinfo.zip, $ZONEINFO, or a
// time/tzdata import by the main program). Although time.LoadLocation accepts
// them, the empty string and "local" (case-insensitive) are rejected because
// they are not real time zone names.
func IsTimeZone(s string) bool {
	if s == "" {
		return false
	}

	if strings.ToLower(s) == "local" {
		return false
	}

	_, err := time.LoadLocation(s)
	return err == nil
}
