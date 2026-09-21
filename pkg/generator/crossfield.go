package generator

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/antlabs/quickvalidate/pkg/parser"
)

const qvPath = "github.com/antlabs/quickvalidate/pkg/validators"

func init() {
	// --- ordering & length comparisons ---------------------------------------
	registry["min"] = cmpTag(">=") // hasMinOf is isGte
	registry["gte"] = cmpTag(">=")
	registry["max"] = cmpTag("<=") // hasMaxOf is isLte
	registry["lte"] = cmpTag("<=")
	registry["gt"] = cmpTag(">")
	registry["lt"] = cmpTag("<")
	registry["len"] = cmpTag("==") // hasLengthOf

	// --- equality -------------------------------------------------------------
	registry["eq"] = eqTag("==")
	registry["ne"] = eqTag("!=")
	registry["eq_ignore_case"] = foldTag(false)
	registry["ne_ignore_case"] = foldTag(true)
	registry["oneof"] = oneOfTag(false)
	registry["oneofci"] = oneOfTag(true)
	registry["unique"] = tagSpec{emit: uniqueTag}

	// --- cross field comparisons ---------------------------------------------
	registry["eqfield"] = fieldCmpTag("==")
	registry["nefield"] = fieldCmpTag("!=")
	registry["gtfield"] = fieldCmpTag(">")
	registry["gtefield"] = fieldCmpTag(">=")
	registry["ltfield"] = fieldCmpTag("<")
	registry["ltefield"] = fieldCmpTag("<=")
	// The cross-struct variants walk the same dotted paths and behave the same
	// way for in-package structs.
	registry["eqcsfield"] = fieldCmpTag("==")
	registry["necsfield"] = fieldCmpTag("!=")
	registry["gtcsfield"] = fieldCmpTag(">")
	registry["gtecsfield"] = fieldCmpTag(">=")
	registry["ltcsfield"] = fieldCmpTag("<")
	registry["ltecsfield"] = fieldCmpTag("<=")
	registry["fieldcontains"] = fieldContainsTag(true)
	registry["fieldexcludes"] = fieldContainsTag(false)

	// --- substring / prefix / suffix -----------------------------------------
	registry["contains"] = strTag(func(e, p string) string { return fmt.Sprintf("strings.Contains(%s, %q)", e, p) })
	registry["containsany"] = strTag(func(e, p string) string { return fmt.Sprintf("strings.ContainsAny(%s, %q)", e, p) })
	registry["containsrune"] = strTag(func(e, p string) string {
		r, _ := utf8.DecodeRuneInString(p)
		return fmt.Sprintf("strings.ContainsRune(%s, %q)", e, r)
	})
	registry["excludes"] = strTag(func(e, p string) string { return fmt.Sprintf("!strings.Contains(%s, %q)", e, p) })
	registry["excludesall"] = strTag(func(e, p string) string { return fmt.Sprintf("!strings.ContainsAny(%s, %q)", e, p) })
	registry["excludesrune"] = strTag(func(e, p string) string {
		r, _ := utf8.DecodeRuneInString(p)
		return fmt.Sprintf("!strings.ContainsRune(%s, %q)", e, r)
	})
	registry["startswith"] = strTag(func(e, p string) string { return fmt.Sprintf("strings.HasPrefix(%s, %q)", e, p) })
	registry["endswith"] = strTag(func(e, p string) string { return fmt.Sprintf("strings.HasSuffix(%s, %q)", e, p) })
	registry["startsnotwith"] = strTag(func(e, p string) string { return fmt.Sprintf("!strings.HasPrefix(%s, %q)", e, p) })
	registry["endsnotwith"] = strTag(func(e, p string) string { return fmt.Sprintf("!strings.HasSuffix(%s, %q)", e, p) })

	// --- conditional presence -------------------------------------------------
	registry["required_if"] = condTag(condRequiredIf)
	registry["required_unless"] = condTag(condRequiredUnless)
	registry["skip_unless"] = condTag(condSkipUnless)
	registry["required_with"] = condTag(condRequiredWith)
	registry["required_with_all"] = condTag(condRequiredWithAll)
	registry["required_without"] = condTag(condRequiredWithout)
	registry["required_without_all"] = condTag(condRequiredWithoutAll)
	registry["excluded_if"] = condTag(condExcludedIf)
	registry["excluded_unless"] = condTag(condExcludedUnless)
	registry["excluded_with"] = condTag(condExcludedWith)
	registry["excluded_with_all"] = condTag(condExcludedWithAll)
	registry["excluded_without"] = condTag(condExcludedWithout)
	registry["excluded_without_all"] = condTag(condExcludedWithoutAll)
}

// strTag renders a strings.* based check; the emitter is string only.
func strTag(fn func(expr, param string) string) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		se.g.addImport("strings", "")
		if ref.knownNil {
			return "false", nil
		}
		if ref.typ.Kind != parser.KindString {
			return "", fmt.Errorf("%s requires a string field, got %s", ref.typ.Expr, ref.typ.Expr)
		}
		return fn(ref.arg(), param), nil
	}}
}

// ---------------------------------------------------------------------------
// comparisons
// ---------------------------------------------------------------------------

func cmpTag(op string) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		return se.compareExpr(ref, op, param)
	}}
}

// compareExpr mirrors isGt/isGte/isLt/isLte/hasLengthOf, where the parameter is
// parsed according to the field's own kind.
func (se *structEmitter) compareExpr(ref fieldRef, op, param string) (string, error) {
	if ref.knownNil {
		return "false", nil
	}
	t := ref.typ
	if t == nil {
		return "", fmt.Errorf("cannot compare an unknown type")
	}

	switch t.Kind {
	case parser.KindString:
		n, err := paramInt(param)
		if err != nil {
			return "", err
		}
		se.g.addImport("unicode/utf8", "")
		return fmt.Sprintf("int64(utf8.RuneCountInString(%s)) %s %d", ref.arg(), op, n), nil

	case parser.KindSlice, parser.KindArray, parser.KindMap:
		n, err := paramInt(param)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int64(len(%s)) %s %d", ref.expr, op, n), nil

	case parser.KindInt:
		var (
			n   int64
			err error
		)
		if t.IsDuration {
			n, err = paramDuration(param)
		} else {
			n, err = paramInt(param)
		}
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int64(%s) %s %d", ref.expr, op, n), nil

	case parser.KindUint:
		n, err := paramUint(param)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("uint64(%s) %s %d", ref.expr, op, n), nil

	case parser.KindFloat:
		v, err := paramFloat(param, t.Bits)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", ref.expr, op, floatLiteral(v, t.Bits)), nil

	case parser.KindStruct:
		if t.IsTime {
			se.g.addImport("time", "")
			switch op {
			case ">":
				return fmt.Sprintf("%s.After(time.Now().UTC())", ref.expr), nil
			case ">=":
				return fmt.Sprintf("!%s.Before(time.Now().UTC())", ref.expr), nil
			case "<":
				return fmt.Sprintf("%s.Before(time.Now().UTC())", ref.expr), nil
			case "<=":
				return fmt.Sprintf("!%s.After(time.Now().UTC())", ref.expr), nil
			}
			return "", fmt.Errorf("comparison %q is not supported on time.Time", op)
		}
	}

	return "", fmt.Errorf("tag is not supported on type %s", t.Expr)
}

// eqTag mirrors isEq/isNe: strings and bools compare by value, containers by
// length, numbers by value.
func eqTag(op string) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		if ref.knownNil {
			return "false", nil
		}
		t := ref.typ
		switch t.Kind {
		case parser.KindString:
			return fmt.Sprintf("%s %s %q", ref.expr, op, param), nil
		case parser.KindBool:
			b, err := paramBool(param)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s %s %t", ref.expr, op, b), nil
		default:
			return se.compareExpr(ref, op, param)
		}
	}}
}

// foldTag mirrors isEqIgnoreCase/isNeIgnoreCase.
func foldTag(negate bool) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		se.g.addImport("strings", "")
		if ref.knownNil {
			return "false", nil
		}
		if ref.typ.Kind != parser.KindString {
			return "", fmt.Errorf("ignore case comparison requires a string field")
		}
		if negate {
			return fmt.Sprintf("!strings.EqualFold(%s, %q)", ref.arg(), param), nil
		}
		return fmt.Sprintf("strings.EqualFold(%s, %q)", ref.arg(), param), nil
	}}
}

// oneOfTag mirrors isOneOf/isOneOfCI.
func oneOfTag(fold bool) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport(qvPath, "qv")
		if ref.knownNil {
			return "false", nil
		}
		vals := parser.SplitParam(param)
		key := fmt.Sprintf("oneof:%t:%s", fold, param)
		name := se.g.helperVar(key, quoteJoin(vals))

		t := ref.typ
		switch t.Kind {
		case parser.KindString:
			fn := "IsOneOf"
			if fold {
				fn = "IsOneOfFold"
			}
			return fmt.Sprintf("%s.%s(%s, %s)", qv, fn, ref.arg(), name), nil
		case parser.KindInt:
			se.g.addImport("strconv", "")
			if fold {
				return "", fmt.Errorf("oneofci requires a string field")
			}
			return fmt.Sprintf("%s.IsOneOf(strconv.FormatInt(int64(%s), 10), %s)", qv, ref.expr, name), nil
		case parser.KindUint:
			se.g.addImport("strconv", "")
			if fold {
				return "", fmt.Errorf("oneofci requires a string field")
			}
			return fmt.Sprintf("%s.IsOneOf(strconv.FormatUint(uint64(%s), 10), %s)", qv, ref.expr, name), nil
		}
		return "", fmt.Errorf("oneof is not supported on type %s", t.Expr)
	}}
}

// uniqueTag mirrors isUnique.
func uniqueTag(se *structEmitter, ref fieldRef, param string) (string, error) {
	qv := se.g.addImport(qvPath, "qv")
	if ref.knownNil {
		return "false", nil
	}
	t := ref.typ

	switch t.Kind {
	case parser.KindSlice, parser.KindArray:
		elem := t.Elem
		if param == "" {
			if elem.Kind == parser.KindPtr && elem.Elem != nil {
				return fmt.Sprintf("%s.UniquePtr(%s)", qv, ref.expr), nil
			}
			return fmt.Sprintf("%s.Unique(%s)", qv, ref.expr), nil
		}
		// unique=Field: deduplicate by a field of the element struct.
		structType := elem
		ptr := false
		if structType.Kind == parser.KindPtr && structType.Elem != nil {
			structType, ptr = structType.Elem, true
		}
		if structType.Kind != parser.KindStruct {
			return "", fmt.Errorf("unique=%s requires a slice of structs", param)
		}
		st := se.g.Pkg.StructByName(structType.Name)
		if st == nil {
			return "", fmt.Errorf("unique=%s: struct %s is not generated in this package", param, structType.Name)
		}
		var field *parser.FieldInfo
		for i := range st.Fields {
			if st.Fields[i].Name == param {
				field = &st.Fields[i]
				break
			}
		}
		if field == nil {
			return "", fmt.Errorf("unique=%s: no such field on %s", param, structType.Name)
		}
		elemExpr := "e"
		keyExpr := "e." + param
		if ptr {
			elemExpr = "e"
			keyExpr = "e." + param
			return fmt.Sprintf("%s.UniqueBy(%s, func(e *%s) (%s, bool) {\nif e == nil { return *new(%s), false }\nreturn %s, true\n})",
				qv, ref.expr, structType.Name, field.Type.Expr, field.Type.Expr, keyExpr), nil
		}
		_ = elemExpr
		return fmt.Sprintf("%s.UniqueBy(%s, func(e %s) (%s, bool) {\nreturn %s, true\n})",
			qv, ref.expr, structType.Name, field.Type.Expr, keyExpr), nil

	case parser.KindMap:
		if param != "" {
			return "", fmt.Errorf("unique=%s is not supported on maps", param)
		}
		if t.Elem.Kind == parser.KindPtr && t.Elem.Elem != nil {
			return fmt.Sprintf("%s.UniqueMapPtr(%s)", qv, ref.expr), nil
		}
		return fmt.Sprintf("%s.UniqueMap(%s)", qv, ref.expr), nil
	}

	// Scalar: `unique=Other` means "differs from the other field".
	if param == "" {
		return "", fmt.Errorf("unique on a scalar field requires a field parameter")
	}
	other, otherType, guard, ok := se.otherRef(param)
	if !ok || otherType == nil || otherType.Kind != t.Kind || guard != "" {
		return "false", nil
	}
	if t.Name != otherType.Name {
		// isUnique compares the two values as interface{}: a named type and its
		// builtin are different dynamic types, so they are never equal.
		return "true", nil
	}
	return fmt.Sprintf("%s != %s", ref.expr, other), nil
}

// fieldCmpTag mirrors isEqField/isNeField/isGtField/... including the quirk
// that ordering comparisons on strings use byte length (not rune count).
func fieldCmpTag(op string) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		if ref.knownNil {
			return "false", nil
		}
		other, otherType, guard, ok := se.otherRef(param)
		if !ok || otherType == nil {
			return notFoundResult(op), nil
		}
		t := ref.typ
		// The reference compares reflect kinds: a named type shares the kind of
		// what it wraps, but int and int64 (or int8 and int64) are different
		// kinds to reflect, and a mismatch counts as "not found".
		if t.Kind != otherType.Kind || t.BuiltinName() != otherType.BuiltinName() {
			return notFoundResult(op), nil
		}

		switch t.Kind {
		case parser.KindString:
			if op == "==" || op == "!=" {
				// isEqField reads both sides through reflect and compares the
				// underlying values, so a named type and its builtin compare by
				// content.
				l, r := cmpOperands(ref.expr, t, other, otherType)
				return withGuard(fmt.Sprintf("%s %s %s", l, op, r), guard, op == "!="), nil
			}
			// isGtField & friends compare byte length for strings.
			return withGuard(fmt.Sprintf("len(%s) %s len(%s)", ref.expr, op, other), guard, false), nil

		case parser.KindInt, parser.KindUint, parser.KindFloat:
			// field.Int()/Uint()/Float() read both sides as one common type.
			l, r := cmpOperands(ref.expr, t, other, otherType)
			return withGuard(fmt.Sprintf("%s %s %s", l, op, r), guard, op == "!="), nil

		case parser.KindSlice, parser.KindArray, parser.KindMap:
			if op == "==" || op == "!=" {
				return fmt.Sprintf("len(%s) %s len(%s)", ref.expr, op, other), nil
			}
			return "", fmt.Errorf("%s is not supported on type %s", op, t.Expr)

		case parser.KindBool:
			if op == "==" || op == "!=" {
				l, r := cmpOperands(ref.expr, t, other, otherType)
				return withGuard(fmt.Sprintf("%s %s %s", l, op, r), guard, op == "!="), nil
			}
			return "", fmt.Errorf("%s is not supported on bool", op)

		case parser.KindStruct:
			if t.IsTime && otherType.IsTime {
				switch op {
				case "==":
					return fmt.Sprintf("%s.Equal(%s)", ref.expr, other), nil
				case "!=":
					return fmt.Sprintf("!%s.Equal(%s)", ref.expr, other), nil
				case ">":
					return fmt.Sprintf("%s.After(%s)", ref.expr, other), nil
				case ">=":
					return fmt.Sprintf("!%s.Before(%s)", ref.expr, other), nil
				case "<":
					return fmt.Sprintf("%s.Before(%s)", ref.expr, other), nil
				case "<=":
					return fmt.Sprintf("!%s.After(%s)", ref.expr, other), nil
				}
			}
		}
		return "", fmt.Errorf("cross field comparison is not supported on type %s", t.Expr)
	}}
}

// fieldContainsTag mirrors fieldContains/fieldExcludes.
func fieldContainsTag(contains bool) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		se.g.addImport("strings", "")
		if ref.knownNil {
			return "false", nil
		}
		other, otherType, guard, ok := se.otherRef(param)
		if !ok || otherType == nil || otherType.Kind != parser.KindString || ref.typ.Kind != parser.KindString {
			// fieldexcludes passes when the other field cannot be resolved.
			if contains {
				return "false", nil
			}
			return "true", nil
		}
		if contains {
			return withGuard(fmt.Sprintf("strings.Contains(%s, %s)", ref.arg(), convert(other, otherType)), guard, false), nil
		}
		return withGuard(fmt.Sprintf("!strings.Contains(%s, %s)", ref.arg(), convert(other, otherType)), guard, true), nil
	}}
}

// ---------------------------------------------------------------------------
// conditional presence
// ---------------------------------------------------------------------------

type condFunc func(se *structEmitter, ref fieldRef, param string) (string, error)

func condTag(fn condFunc) tagSpec {
	return tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		return fn(se, ref, param)
	}}
}

func condRequiredIf(se *structEmitter, ref fieldRef, param string) (string, error) {
	matches, err := se.matchPairs(param)
	if err != nil {
		return "", err
	}
	// valid when any pair does not match, otherwise the field must have a value
	return fmt.Sprintf("!(%s) || %s", strings.Join(matches, " && "), se.hasValueExpr(ref)), nil
}

func condRequiredUnless(se *structEmitter, ref fieldRef, param string) (string, error) {
	matches, err := se.matchPairs(param)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s) || %s", strings.Join(matches, " || "), se.hasValueExpr(ref)), nil
}

// condSkipUnless mirrors skipUnless: valid as soon as one pair does not match.
func condSkipUnless(se *structEmitter, ref fieldRef, param string) (string, error) {
	matches, err := se.matchPairs(param)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("!(%s) || %s", strings.Join(matches, " && "), se.hasValueExpr(ref)), nil
}

func condExcludedIf(se *structEmitter, ref fieldRef, param string) (string, error) {
	matches, err := se.matchPairs(param)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("!(%s) || !(%s)", strings.Join(matches, " && "), se.hasValueExpr(ref)), nil
}

func condExcludedUnless(se *structEmitter, ref fieldRef, param string) (string, error) {
	matches, err := se.matchPairs(param)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s) || !(%s)", strings.Join(matches, " && "), se.hasValueExpr(ref)), nil
}

func condRequiredWith(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(param, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s) || %s", strings.Join(zeros, " && "), se.hasValueExpr(ref)), nil
}

func condRequiredWithAll(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(param, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s) || %s", strings.Join(zeros, " || "), se.hasValueExpr(ref)), nil
}

func condRequiredWithout(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(strings.TrimSpace(param), true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("!(%s) || %s", strings.Join(zeros, " && "), se.hasValueExpr(ref)), nil
}

func condRequiredWithoutAll(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(param, false)
	if err != nil {
		return "", err
	}
	// valid as soon as one of the listed fields is present
	return fmt.Sprintf("!(%s) || %s", strings.Join(zeros, " && "), se.hasValueExpr(ref)), nil
}

func condExcludedWith(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(param, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s) || !(%s)", strings.Join(zeros, " && "), se.hasValueExpr(ref)), nil
}

func condExcludedWithAll(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(param, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s) || !(%s)", strings.Join(zeros, " || "), se.hasValueExpr(ref)), nil
}

func condExcludedWithout(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(strings.TrimSpace(param), true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("!(%s) || !(%s)", strings.Join(zeros, " && "), se.hasValueExpr(ref)), nil
}

func condExcludedWithoutAll(se *structEmitter, ref fieldRef, param string) (string, error) {
	zeros, err := se.zeroFields(param, false)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("!(%s) || !(%s)", strings.Join(zeros, " && "), se.hasValueExpr(ref)), nil
}

// matchPairs renders `field value field value` comparisons as an AND list.
func (se *structEmitter) matchPairs(param string) ([]string, error) {
	vals := parser.SplitParam(param)
	if len(vals)%2 != 0 {
		return nil, fmt.Errorf("bad param number for conditional tag %q", param)
	}
	var out []string
	for i := 0; i < len(vals); i += 2 {
		expr, typ, guard, ok := se.otherRef(vals[i])
		if !ok || typ == nil {
			out = append(out, "false")
			continue
		}
		e, err := se.matchValue(expr, typ, vals[i+1])
		if err != nil {
			return nil, err
		}
		out = append(out, withGuard(e, guard, false))
	}
	if len(out) == 0 {
		out = append(out, "true")
	}
	return out, nil
}

// matchValue mirrors requireCheckFieldValue.
func (se *structEmitter) matchValue(expr string, typ *parser.TypeInfo, value string) (string, error) {
	switch typ.Kind {
	case parser.KindInt:
		n, err := paramInt(value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int64(%s) == %d", expr, n), nil

	case parser.KindUint:
		n, err := paramUint(value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("uint64(%s) == %d", expr, n), nil

	case parser.KindFloat:
		v, err := paramFloat(value, typ.Bits)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s == %s", expr, floatLiteral(v, typ.Bits)), nil

	case parser.KindSlice, parser.KindArray, parser.KindMap:
		n, err := paramInt(value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int64(len(%s)) == %d", expr, n), nil

	case parser.KindBool:
		return fmt.Sprintf("%s == %t", expr, value == "true"), nil

	case parser.KindPtr:
		if typ.Elem == nil {
			return "false", nil
		}
		inner, err := se.matchValue("(*"+expr+")", typ.Elem, value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s == nil && %t) || (%s != nil && %s)", expr, value == "nil", expr, inner), nil
	}

	return fmt.Sprintf("%s == %q", expr, value), nil
}

// zeroFields renders requireCheckFieldKind for each listed field: true when the
// field holds its zero value.
func (se *structEmitter) zeroFields(param string, defaultNotFound bool) ([]string, error) {
	vals := parser.SplitParam(param)
	var out []string
	for _, name := range vals {
		expr, typ, guard, ok := se.otherRef(name)
		if !ok || typ == nil {
			out = append(out, fmt.Sprintf("%t", defaultNotFound))
			continue
		}
		switch typ.Kind {
		case parser.KindSlice, parser.KindMap, parser.KindPtr, parser.KindInterface:
			out = append(out, withGuard(expr+" == nil", guard, defaultNotFound))
		default:
			out = append(out, withGuard("!("+se.zeroCheck(expr, typ, false)+")", guard, defaultNotFound))
		}
	}
	if len(out) == 0 {
		out = append(out, "true")
	}
	return out, nil
}

func init() {
	// port reads the field as an unsigned integer (validator panics on strings).
	registry["port"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport(qvPath, "qv")
		if ref.knownNil {
			return "false", nil
		}
		switch ref.typ.Kind {
		case parser.KindUint:
			return fmt.Sprintf("%s.IsPort(uint64(%s))", qv, ref.expr), nil
		case parser.KindInt:
			return fmt.Sprintf("%s.IsPort(uint64(%s))", qv, ref.expr), nil
		}
		return "", fmt.Errorf("port requires an integer field, got %s", ref.typ.Expr)
	}}

	// spicedb carries a parameter naming the relation to check. The reference
	// panics on anything outside this set, so an unknown one is a generation
	// time error.
	registry["spicedb"] = tagSpec{emit: func(se *structEmitter, ref fieldRef, param string) (string, error) {
		qv := se.g.addImport(qvPath, "qv")
		switch param {
		case "", "id", "type", "permission":
		default:
			return "", fmt.Errorf("spicedb: unknown parameter %q (want permission, type or id)", param)
		}
		if ref.knownNil {
			return "false", nil
		}
		if err := requireString("spicedb", ref); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.IsSpiceDB(%s, %q)", qv, ref.arg(), param), nil
	}}
}
