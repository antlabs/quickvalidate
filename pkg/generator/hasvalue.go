package generator

import (
	"fmt"
	"strings"

	"github.com/antlabs/quickvalidate/pkg/parser"
)

// hasValueExpr renders validator's hasValue() for the given ref: true when the
// value is not the zero value for its type.
func (se *structEmitter) hasValueExpr(ref fieldRef) string {
	if ref.knownNil {
		return "false"
	}
	if ref.fromPointer {
		// validator: a non-nil pointer is "set" no matter what it points at.
		return "true"
	}
	t := ref.typ
	if t == nil {
		return "false"
	}
	if t.Kind == parser.KindPtr || t.Kind == parser.KindInterface {
		// validator returns !field.IsNil() for these kinds.
		return ref.expr + " != nil"
	}
	return se.zeroCheck(ref.expr, t, false)
}

// hasNotZeroValueExpr renders validator's hasNotZeroValue(), used by omitzero.
// It only differs from hasValue for pointer fields: a non-nil pointer is not
// automatically "set", the pointed-to value must be non-zero too.
func (se *structEmitter) hasNotZeroValueExpr(ref fieldRef) string {
	if ref.knownNil {
		return "false"
	}
	t := ref.typ
	if t == nil {
		return "false"
	}
	if t.Kind == parser.KindPtr && t.Elem != nil {
		return fmt.Sprintf("%s != nil && %s", ref.expr, se.zeroCheck("(*"+ref.expr+")", t.Elem, false))
	}
	if t.Kind == parser.KindInterface {
		return ref.expr + " != nil"
	}
	return se.zeroCheck(ref.expr, t, false)
}

// zeroCheck renders "the value is not zero". isPointer mirrors the
// fldIsPointer branch of hasValue: a non-nil pointer to a value is always set.
func (se *structEmitter) zeroCheck(expr string, t *parser.TypeInfo, isPointer bool) string {
	if isPointer {
		return "true"
	}

	switch t.Kind {
	case parser.KindString:
		return expr + ` != ""`
	case parser.KindBool:
		return expr
	case parser.KindInt, parser.KindUint, parser.KindFloat:
		return expr + " != 0"
	case parser.KindSlice, parser.KindMap:
		return expr + " != nil"
	case parser.KindArray:
		se.g.addImport("reflect", "")
		return "!reflect.ValueOf(" + expr + ").IsZero()"
	case parser.KindStruct:
		if t.IsTime {
			return "!" + expr + ".IsZero()"
		}
		se.g.addImport("reflect", "")
		return "!reflect.ValueOf(" + expr + ").IsZero()"
	case parser.KindPtr, parser.KindInterface:
		return expr + " != nil"
	}
	return "false"
}

// isZeroExpr is the negation helper used by isdefault and omitempty handling.
func (se *structEmitter) isZeroExpr(ref fieldRef) string {
	return "!(" + se.hasValueExpr(ref) + ")"
}

// quoteJoin renders a []string literal with the given values.
func quoteJoin(vals []string) string {
	quoted := make([]string, len(vals))
	for i, v := range vals {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[]string{" + strings.Join(quoted, ", ") + "}"
}
