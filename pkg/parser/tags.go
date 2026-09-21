package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// TagType mirrors the `typeof` of go-playground/validator's internal cTag.
type TagType int

const (
	TypeDefault TagType = iota
	TypeDive
	TypeKeys
	TypeEndKeys
	TypeOmitEmpty
	TypeOmitNil
	TypeOmitZero
	TypeIsDefault
	TypeStructOnly
	TypeNoStructLevel
	TypeOr
)

const (
	diveTag          = "dive"
	keysTag          = "keys"
	endKeysTag       = "endkeys"
	omitemptyTag     = "omitempty"
	omitnilTag       = "omitnil"
	omitzeroTag      = "omitzero"
	isdefaultTag     = "isdefault"
	structOnlyTag    = "structonly"
	noStructLevelTag = "nostructlevel"

	tagSeparator    = ","
	orSeparator     = "|"
	tagKeySeparator = "="

	utf8HexComma = "0x2C"
	utf8Pipe     = "0x7C"
)

// splitParamsRegexString mirrors go-playground/validator's splitParamsRegexString.
const splitParamsRegexString = `'[^']*'|\S+`

var splitParams = regexp.MustCompile(splitParamsRegexString)

// TagNode is one node of a parsed `validate` tag chain. It mirrors the shape of
// go-playground/validator's internal cTag so that traversal semantics can be
// reproduced exactly.
type TagNode struct {
	Type TagType

	// Tag is the validation tag name, e.g. "min" (never contains "=" or "|").
	Tag string
	// Param is the unescaped parameter, e.g. "3" for `min=3`.
	Param string
	// HasParam reports whether the original tag carried "=".
	HasParam bool

	// AliasTag is the tag name reported on failure. It is the alias name when the
	// tag came from an alias (e.g. "iscolor"), "" for plain or-blocks (the full
	// expression is reported instead, see OrExpr), otherwise equal to Tag.
	AliasTag   string
	ActualTag  string
	HasAlias   bool
	IsBlockEnd bool

	// RunWhenNil mirrors registerValidation's runValidationOnNil flag: the check
	// still runs when the field is nil (the conditional required_*/excluded_* family).
	RunWhenNil bool

	// Keys holds the tag chain applied to map keys (set on TypeKeys nodes).
	Keys *TagNode

	Next *TagNode
}

// OrExpr renders the tag as validator would report it for a failed or-block,
// e.g. "rgb|rgba" or "oneof=a b|oneof=c d".
func (n *TagNode) OrExpr() string {
	var b strings.Builder
	for c := n; c != nil && (c.Type == TypeOr || c == n); c = c.Next {
		if b.Len() > 0 {
			b.WriteString(orSeparator)
		}
		b.WriteString(c.Tag)
		if c.HasParam {
			b.WriteString(tagKeySeparator)
			b.WriteString(c.Param)
		}
		if c.IsBlockEnd {
			break
		}
	}
	return b.String()
}

// AliasExpansions mirrors go-playground/validator's bakedInAliases.
var AliasExpansions = map[string]string{
	"iscolor":         "hexcolor|rgb|rgba|hsl|hsla",
	"country_code":    "iso3166_1_alpha2|iso3166_1_alpha3|iso3166_1_alpha_numeric",
	"eu_country_code": "iso3166_1_alpha2_eu|iso3166_1_alpha3_eu|iso3166_1_alpha_numeric_eu",
}

// RunWhenNilTags mirrors the validators registered with runValidationOnNil=true.
var RunWhenNilTags = map[string]bool{
	"required_if": true, "required_unless": true, "required_with": true,
	"required_with_all": true, "required_without": true, "required_without_all": true,
	"excluded_if": true, "excluded_unless": true, "excluded_with": true,
	"excluded_with_all": true, "excluded_without": true, "excluded_without_all": true,
	"skip_unless": true,
}

// KnownTags is the set of validation tags the generator can emit. The generator
// fills it from its registry so that unknown tags are rejected at parse time.
var KnownTags = map[string]bool{}

// ParseTags parses a `validate` struct tag into a tag chain, mirroring
// go-playground/validator's parseFieldTagsRecursive.
func ParseTags(tag string, fieldName string) (*TagNode, error) {
	first, _, err := parseTagsRecursive(tag, fieldName, "", false)
	return first, err
}

func parseTagsRecursive(tag string, fieldName string, alias string, hasAlias bool) (firstCtag *TagNode, lastCtag *TagNode, err error) {
	var t string
	noAlias := len(alias) == 0
	tags := strings.Split(tag, tagSeparator)

	for i := 0; i < len(tags); i++ {
		t = tags[i]
		if noAlias {
			alias = t
		}

		if tagsVal, found := AliasExpansions[t]; found {
			if i == 0 {
				firstCtag, lastCtag, err = parseTagsRecursive(tagsVal, fieldName, t, true)
				if err != nil {
					return nil, nil, err
				}
			} else {
				next, curr, err := parseTagsRecursive(tagsVal, fieldName, t, true)
				if err != nil {
					return nil, nil, err
				}
				lastCtag.Next, lastCtag = next, curr
			}
			continue
		}

		var prevTag TagType

		if i == 0 {
			lastCtag = &TagNode{AliasTag: alias, HasAlias: hasAlias}
			firstCtag = lastCtag
		} else {
			prevTag = lastCtag.Type
			lastCtag.Next = &TagNode{AliasTag: alias, HasAlias: hasAlias}
			lastCtag = lastCtag.Next
		}

		switch t {
		case diveTag:
			lastCtag.Type = TypeDive
			continue

		case keysTag:
			if i == 0 || prevTag != TypeDive {
				return nil, nil, fmt.Errorf("'%s' tag must be immediately preceded by the '%s' tag", keysTag, diveTag)
			}

			lastCtag.Type = TypeKeys

			// pass along only the keys tags
			b := make([]byte, 0, 64)
			i++
			for ; i < len(tags); i++ {
				b = append(b, tags[i]...)
				b = append(b, ',')
				if tags[i] == endKeysTag {
					break
				}
			}
			lastCtag.Keys, _, err = parseTagsRecursive(string(b[:len(b)-1]), fieldName, "", false)
			if err != nil {
				return nil, nil, err
			}
			continue

		case endKeysTag:
			lastCtag.Type = TypeEndKeys
			if i != len(tags)-1 {
				return nil, nil, fmt.Errorf("'%s' tag must be immediately preceded by the '%s' tag", keysTag, diveTag)
			}
			return firstCtag, lastCtag, nil

		case omitzeroTag:
			lastCtag.Type = TypeOmitZero
			continue

		case omitemptyTag:
			lastCtag.Type = TypeOmitEmpty
			continue

		case omitnilTag:
			lastCtag.Type = TypeOmitNil
			continue

		case structOnlyTag:
			lastCtag.Type = TypeStructOnly
			continue

		case noStructLevelTag:
			lastCtag.Type = TypeNoStructLevel
			continue

		default:
			if t == isdefaultTag {
				lastCtag.Type = TypeIsDefault
			}

			orVals := strings.Split(t, orSeparator)

			for j := 0; j < len(orVals); j++ {
				vals := strings.SplitN(orVals[j], tagKeySeparator, 2)
				if noAlias {
					alias = vals[0]
					lastCtag.AliasTag = alias
				} else {
					lastCtag.ActualTag = t
				}

				if j > 0 {
					lastCtag.Next = &TagNode{AliasTag: alias, ActualTag: lastCtag.ActualTag, HasAlias: hasAlias}
					lastCtag = lastCtag.Next
				}
				lastCtag.HasParam = len(vals) > 1

				lastCtag.Tag = vals[0]
				if len(lastCtag.Tag) == 0 {
					return nil, nil, fmt.Errorf("invalid validation tag on field '%s'", fieldName)
				}

				lastCtag.RunWhenNil = RunWhenNilTags[lastCtag.Tag]
				if len(KnownTags) > 0 && !KnownTags[lastCtag.Tag] {
					return nil, nil, fmt.Errorf("undefined validation function '%s' on field '%s'", lastCtag.Tag, fieldName)
				}

				if len(orVals) > 1 {
					lastCtag.Type = TypeOr
				}

				if len(vals) > 1 {
					lastCtag.Param = strings.Replace(strings.Replace(vals[1], utf8HexComma, ",", -1), utf8Pipe, "|", -1)
				}
			}
			lastCtag.IsBlockEnd = true
		}
	}

	return firstCtag, lastCtag, nil
}

// SplitParam splits a whitespace separated parameter list the same way
// validator's parseOneOfParam2 does (single quoted values are kept whole).
func SplitParam(param string) []string {
	return splitParams.FindAllString(param, -1)
}
