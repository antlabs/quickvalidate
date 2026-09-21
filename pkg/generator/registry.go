package generator

import (
	"fmt"
	"strings"

	"github.com/antlabs/quickvalidate/pkg/parser"
)

type emitFunc func(se *structEmitter, ref fieldRef, param string) (string, error)

type tagSpec struct {
	// emit returns a Go expression that is true when the check passes.
	emit emitFunc
}

// registry maps every supported validation tag to its emitter.
var registry = map[string]tagSpec{}

// knownTags returns the full tag set accepted by the parser, including the
// structural tags handled by the traversal itself.
func knownTags() map[string]bool {
	out := make(map[string]bool, len(registry)+16)
	for k := range registry {
		out[k] = true
	}
	for _, t := range []string{
		"dive", "keys", "endkeys", "omitempty", "omitnil", "omitzero",
		"structonly", "nostructlevel", "isdefault", "required",
	} {
		out[t] = true
	}
	for k := range parser.AliasExpansions {
		out[k] = true
	}
	return out
}

// register records the emitter for a tag.
//
// Registering the same tag twice is always a bug: init() runs per file in file
// name order, so a bulk registration can silently shadow a dedicated emitter —
// which is how "iso4217_numeric" lost the kind-aware emitter that reads integer
// fields. Fail loudly instead of dropping one of them.
func register(tag string, spec tagSpec) {
	if _, dup := registry[tag]; dup {
		panic("quickvalidate: duplicate registration for tag " + tag)
	}
	registry[tag] = spec
}

// kindHelper registers a tag whose reference implementation switches on the
// field kind: strings go through the string entry point, integer fields through
// the signed or unsigned one. An empty entry point means the reference rejects
// that kind (it panics on it), which is reported at generation time instead.
func kindHelper(tag, strFn, intFn, uintFn string) {
	register(tag, tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		switch ref.typ.Kind {
		case parser.KindString:
			if strFn != "" {
				return fmt.Sprintf("%s.%s(%s)", qv, strFn, ref.arg()), nil
			}
		case parser.KindInt:
			if intFn != "" {
				return fmt.Sprintf("%s.%s(int64(%s))", qv, intFn, ref.expr), nil
			}
		case parser.KindUint:
			if uintFn != "" {
				return fmt.Sprintf("%s.%s(uint64(%s))", qv, uintFn, ref.expr), nil
			}
		}
		return "", fmt.Errorf("%s is not supported on %s (the reference implementation panics on it)", tag, ref.typ.Expr)
	}})
}

// requireString rejects a string only tag on a field that cannot be read as a
// string. The reference reads such fields through reflect's String(), so it
// would keep failing silently for the rest of the program's life; refusing at
// generation time is the only way the mistake gets noticed.
func requireString(tag string, ref fieldRef) error {
	if ref.typ.Kind == parser.KindString {
		return nil
	}
	return fmt.Errorf("%s requires a string field, got %s", tag, ref.typ.Expr)
}

// helper registers a tag emitting a call to a runtime helper package function.
func helper(tag, fn string, argCount int) {
	register(tag, tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		if err := requireString(tag, ref); err != nil {
			return "", err
		}
		if argCount == 0 {
			return fmt.Sprintf("%s.%s(%s)", qv, fn, ref.arg()), nil
		}
		return fmt.Sprintf("%s.%s(%s, %q)", qv, fn, ref.arg(), param), nil
	}})
}

// stringHelper registers a tag that only applies to string fields.
func stringHelper(tag, fn string) {
	register(tag, tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		if err := requireString(tag, ref); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%s(%s)", qv, fn, ref.arg()), nil
	}})
}

func init() {
	// required is hasValue, except on (non time.Time) struct fields: validator
	// skips it unless WithRequiredStructEnabled is used, which the generator
	// deliberately mirrors.
	registry["required"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		if ref.typ.Kind == parser.KindStruct && !ref.typ.IsTime {
			return "true", nil
		}
		return se.hasValueExpr(ref), nil
	}}

	// isdefault is the opposite of required.
	registry["isdefault"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		return "!(" + se.hasValueExpr(ref) + ")", nil
	}}

	// datetime carries the layout as its parameter: `datetime=2006-01-02`.
	registry["datetime"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		if err := requireString("datetime", ref); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.IsDatetime(%s, %q)", qv, ref.arg(), param), nil
	}}

	// --- container agnostic helpers -------------------------------------------------
	helper("email", "IsEmail", 0)
	helper("url", "IsURL", 0)
	helper("http_url", "IsHTTPURL", 0)
	helper("uri", "IsURI", 0)
	helper("urn_rfc2141", "IsUrnRFC2141", 0)
	helper("base32", "IsBase32", 0)
	helper("base64", "IsBase64", 0)
	helper("base64url", "IsBase64URL", 0)
	helper("base64rawurl", "IsBase64RawURL", 0)
	helper("jwt", "IsJWT", 0)
	helper("html", "IsHTML", 0)
	helper("html_encoded", "IsHTMLEncoded", 0)
	helper("url_encoded", "IsURLEncoded", 0)

	// --- string only helpers --------------------------------------------------------
	for tag, fn := range map[string]string{
		"alpha":                     "IsAlpha",
		"alphanum":                  "IsAlphanum",
		"alphaunicode":              "IsAlphaUnicode",
		"alphanumunicode":           "IsAlphanumUnicode",
		"boolean":                   "IsBoolean",
		"numeric":                   "IsNumeric",
		"number":                    "IsNumber",
		"hexadecimal":               "IsHexadecimal",
		"hexcolor":                  "IsHEXColor",
		"rgb":                       "IsRGB",
		"rgba":                      "IsRGBA",
		"hsl":                       "IsHSL",
		"hsla":                      "IsHSLA",
		"e164":                      "IsE164",
		"ascii":                     "IsASCII",
		"printascii":                "IsPrintableASCII",
		"multibyte":                 "IsMultiByteCharacter",
		"datauri":                   "IsDataURI",
		"latitude":                  "IsLatitude",
		"longitude":                 "IsLongitude",
		"ssn":                       "IsSSN",
		"ipv4":                      "IsIPv4",
		"ipv6":                      "IsIPv6",
		"ip":                        "IsIP",
		"cidrv4":                    "IsCIDRv4",
		"cidrv6":                    "IsCIDRv6",
		"cidr":                      "IsCIDR",
		"mac":                       "IsMAC",
		"hostname":                  "IsHostnameRFC952",
		"hostname_rfc1123":          "IsHostnameRFC1123",
		"fqdn":                      "IsFQDN",
		"hostname_port":             "IsHostnamePort",
		"lowercase":                 "IsLowercase",
		"uppercase":                 "IsUppercase",
		"timezone":                  "IsTimeZone",
		"isbn":                      "IsISBN",
		"isbn10":                    "IsISBN10",
		"isbn13":                    "IsISBN13",
		"issn":                      "IsISSN",
		"eth_addr":                  "IsEthereumAddress",
		"eth_addr_checksum":         "IsEthereumAddressChecksum",
		"btc_addr":                  "IsBitcoinAddress",
		"btc_addr_bech32":           "IsBitcoinBech32Address",
		"uuid":                      "IsUUID",
		"uuid3":                     "IsUUID3",
		"uuid4":                     "IsUUID4",
		"uuid5":                     "IsUUID5",
		"uuid_rfc4122":              "IsUUIDRFC4122",
		"uuid3_rfc4122":             "IsUUID3RFC4122",
		"uuid4_rfc4122":             "IsUUID4RFC4122",
		"uuid5_rfc4122":             "IsUUID5RFC4122",
		"ulid":                      "IsULID",
		"md4":                       "IsMD4",
		"md5":                       "IsMD5",
		"sha256":                    "IsSHA256",
		"sha384":                    "IsSHA384",
		"sha512":                    "IsSHA512",
		"ripemd128":                 "IsRIPEMD128",
		"ripemd160":                 "IsRIPEMD160",
		"tiger128":                  "IsTIGER128",
		"tiger160":                  "IsTIGER160",
		"tiger192":                  "IsTIGER192",
		"iso3166_1_alpha2":          "IsIso3166Alpha2",
		"iso3166_1_alpha2_eu":       "IsIso3166Alpha2EU",
		"iso3166_1_alpha3":          "IsIso3166Alpha3",
		"iso3166_1_alpha3_eu":       "IsIso3166Alpha3EU",
		"iso3166_2":                 "IsIso31662",
		"iso4217":                   "IsIso4217",
		"bcp47_language_tag":        "IsBCP47LanguageTag",
		"bic":                       "IsBIC",
		"semver":                    "IsSemver",
		"cve":                       "IsCVE",
		"dns_rfc1035_label":         "IsDnsRFC1035Label",
		"credit_card":               "IsCreditCard",
		"luhn_checksum":             "IsLuhnChecksum",
		"mongodb":                   "IsMongoDBObjectID",
		"mongodb_connection_string": "IsMongoDBConnectionString",
		"cron":                      "IsCron",
		"tcp4_addr":                 "IsTCP4AddrResolvable",
		"tcp6_addr":                 "IsTCP6AddrResolvable",
		"tcp_addr":                  "IsTCPAddrResolvable",
		"udp4_addr":                 "IsUDP4AddrResolvable",
		"udp6_addr":                 "IsUDP6AddrResolvable",
		"udp_addr":                  "IsUDPAddrResolvable",
		"ip4_addr":                  "IsIP4AddrResolvable",
		"ip6_addr":                  "IsIP6AddrResolvable",
		"ip_addr":                   "IsIPAddrResolvable",
		"unix_addr":                 "IsUnixAddrResolvable",
	} {
		stringHelper(tag, fn)
	}

	// json reads a string, or a byte slice (the reference switches on the field
	// kind and panics on anything else).
	register("json", tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		switch {
		case ref.typ.Kind == parser.KindString:
			return fmt.Sprintf("%s.IsJSON(%s)", qv, ref.arg()), nil
		case ref.typ.IsBytes():
			expr := ref.expr
			if ref.typ.Named {
				expr = "[]byte(" + expr + ")"
			}
			return fmt.Sprintf("%s.IsJSONBytes(%s)", qv, expr), nil
		}
		return "", fmt.Errorf("json requires a string or []byte field, got %s", ref.typ.Expr)
	}})

	// Numeric codes read the field as a string or as an integer, depending on
	// its kind. iso4217_numeric has no string branch at all: the reference
	// panics on string fields, so that combination is a generation-time error.
	kindHelper("iso3166_1_alpha_numeric", "IsIso3166AlphaNumeric",
		"IsIso3166AlphaNumericInt", "IsIso3166AlphaNumericUint")
	kindHelper("iso3166_1_alpha_numeric_eu", "IsIso3166AlphaNumericEU",
		"IsIso3166AlphaNumericEUInt", "IsIso3166AlphaNumericEUUint")
	kindHelper("iso4217_numeric", "", "IsIso4217NumericInt", "IsIso4217NumericUint")

	// file / dir family accept a path string.
	register("file", pathHelper("IsFile"))
	register("filepath", pathHelper("IsFilePath"))
	register("dir", pathHelper("IsDir"))
	register("dirpath", pathHelper("IsDirPath"))

	// image accepts both []byte content and a path.
	registry["image"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		if ref.typ.IsBytes() {
			// v10.25.0's isImage only implements the string branch: byte slices
			// always fail, mirror that instead of sniffing content.
			return "false", nil
		}
		return fmt.Sprintf("%s.IsImageFile(%s)", qv, ref.arg()), nil
	}}

	// postcode_iso3166_alpha2=US / =OtherField
	registry["postcode_iso3166_alpha2"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if err := requireString("postcode_iso3166_alpha2", ref); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.IsPostcodeByIso3166Alpha2(%s, %q)", qv, ref.arg(), strings.ToUpper(param)), nil
	}}
	registry["postcode_iso3166_alpha2_field"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		other, otherType, guard, ok := se.otherRef(param)
		if !ok {
			return "", fmt.Errorf("postcode_iso3166_alpha2_field: cannot resolve field %q", param)
		}
		if guard != "" {
			// The country code has to be read to compare the postcode against it,
			// and a nil pointer on the way makes that impossible.
			return "", fmt.Errorf("postcode_iso3166_alpha2_field: cannot traverse a pointer to reach %q", param)
		}
		if err := requireString("postcode_iso3166_alpha2_field", ref); err != nil {
			return "", err
		}
		if err := requireString("postcode_iso3166_alpha2_field", fieldRef{typ: otherType}); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.IsPostcodeByIso3166Alpha2Field(%s, %s)",
			qv, ref.arg(), convert(other, otherType)), nil
	}}
}

func pathHelper(fn string) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport("github.com/antlabs/quickvalidate/pkg/validators", "qv")
		if ref.knownNil {
			return "false", nil
		}
		if ref.typ.Kind != parser.KindString {
			return "", fmt.Errorf("%s requires a string path field, got %s", fn, ref.typ.Expr)
		}
		return fmt.Sprintf("%s.%s(%s)", qv, fn, ref.arg()), nil
	}}
}
