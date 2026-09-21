// Ported from github.com/go-playground/validator/v10 (baked_in.go).
// Copyright (c) 2015 Dean Karn. MIT License — see LICENSE.

package validators

import (
	"strconv"

	"golang.org/x/text/language"
)

// IsIso3166Alpha2 validates if the string is a valid iso3166-1 alpha-2 country code.
func IsIso3166Alpha2(str string) bool {
	_, ok := iso3166_1_alpha2[str]
	return ok
}

// IsIso3166Alpha2EU validates if the string is a valid iso3166-1 alpha-2 European Union country code.
func IsIso3166Alpha2EU(str string) bool {
	_, ok := iso3166_1_alpha2_eu[str]
	return ok
}

// IsIso3166Alpha3 validates if the string is a valid iso3166-1 alpha-3 country code.
func IsIso3166Alpha3(str string) bool {
	_, ok := iso3166_1_alpha3[str]
	return ok
}

// IsIso3166Alpha3EU validates if the string is a valid iso3166-1 alpha-3 European Union country code.
func IsIso3166Alpha3EU(str string) bool {
	_, ok := iso3166_1_alpha3_eu[str]
	return ok
}

// IsIso3166AlphaNumeric validates if the string is a valid iso3166-1 alpha-numeric country code.
//
// As in the reference implementation (which accepts both string and integer
// fields), the parsed value is reduced modulo 1000 before the lookup, and
// strings that are not valid integers are rejected.
func IsIso3166AlphaNumeric(str string) bool {
	i, err := strconv.Atoi(str)
	if err != nil {
		return false
	}

	_, ok := iso3166_1_alpha_numeric[i%1000]
	return ok
}

// IsIso3166AlphaNumericEU validates if the string is a valid iso3166-1 alpha-numeric European Union country code.
//
// As in the reference implementation (which accepts both string and integer
// fields), the parsed value is reduced modulo 1000 before the lookup, and
// strings that are not valid integers are rejected.
func IsIso3166AlphaNumericEU(str string) bool {
	i, err := strconv.Atoi(str)
	if err != nil {
		return false
	}

	_, ok := iso3166_1_alpha_numeric_eu[i%1000]
	return ok
}

// IsIso3166AlphaNumericInt is the signed integer branch of the reference
// implementation: it reads the field as an integer and reduces it modulo 1000.
func IsIso3166AlphaNumericInt(code int64) bool {
	_, ok := iso3166_1_alpha_numeric[int(code%1000)]
	return ok
}

// IsIso3166AlphaNumericUint is the unsigned integer branch of the reference
// implementation. The modulo happens before the conversion to int, so values
// above MaxInt64 wrap exactly like the reference does.
func IsIso3166AlphaNumericUint(code uint64) bool {
	_, ok := iso3166_1_alpha_numeric[int(code%1000)]
	return ok
}

// IsIso3166AlphaNumericEUInt is the signed integer branch of the reference
// implementation for European Union country codes.
func IsIso3166AlphaNumericEUInt(code int64) bool {
	_, ok := iso3166_1_alpha_numeric_eu[int(code%1000)]
	return ok
}

// IsIso3166AlphaNumericEUUint is the unsigned integer branch of the reference
// implementation for European Union country codes.
func IsIso3166AlphaNumericEUUint(code uint64) bool {
	_, ok := iso3166_1_alpha_numeric_eu[int(code%1000)]
	return ok
}

// IsIso31662 validates if the string is a valid iso3166-2 code.
func IsIso31662(str string) bool {
	_, ok := iso3166_2[str]
	return ok
}

// IsIso4217 validates if the string is a valid iso4217 currency code.
func IsIso4217(str string) bool {
	_, ok := iso4217[str]
	return ok
}

// IsIso4217NumericInt is the signed integer branch of the reference
// implementation: the value is truncated to int before the lookup.
func IsIso4217NumericInt(code int64) bool {
	_, ok := iso4217_numeric[int(code)]
	return ok
}

// IsIso4217NumericUint is the unsigned integer branch of the reference
// implementation.
func IsIso4217NumericUint(code uint64) bool {
	_, ok := iso4217_numeric[int(code)]
	return ok
}

// IsBCP47LanguageTag validates if the string is a valid BCP 47 language tag, as parsed by language.Parse.
func IsBCP47LanguageTag(str string) bool {
	_, err := language.Parse(str)
	return err == nil
}

// IsPostcodeByIso3166Alpha2 validates the postcode by the given country code in iso 3166 alpha 2,
// example: `postcode_iso3166_alpha2=US`.
// Unknown country codes (including the empty string) are rejected.
func IsPostcodeByIso3166Alpha2(postcode, countryCode string) bool {
	postcodeRegexInit.Do(initPostcodes)
	reg, found := postCodeRegexDict[countryCode]
	if !found {
		return false
	}

	return reg.MatchString(postcode)
}

// IsPostcodeByIso3166Alpha2Field validates the postcode by a separately supplied value
// representing a country code in iso 3166 alpha 2, example: `postcode_iso3166_alpha2_field=CountryCode`.
// Unknown country codes (including the empty string) are rejected.
func IsPostcodeByIso3166Alpha2Field(postcode, countryCode string) bool {
	postcodeRegexInit.Do(initPostcodes)
	reg, found := postCodeRegexDict[countryCode]
	if !found {
		return false
	}

	return reg.MatchString(postcode)
}
