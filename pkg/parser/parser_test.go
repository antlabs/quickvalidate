package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// parseSource writes src to a throwaway package and parses it.
func parseSource(t *testing.T, src string) *Package {
	t.Helper()

	path := filepath.Join(t.TempDir(), "models.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg, err := ParseFiles([]string{path}, nil)
	if err != nil {
		t.Fatalf("ParseFiles: %v", err)
	}
	return pkg
}

func fields(t *testing.T, src, structName string) map[string]*FieldInfo {
	t.Helper()

	pkg := parseSource(t, src)

	var s *StructInfo
	for _, cand := range pkg.Structs {
		if cand.Name == structName {
			s = cand
			break
		}
	}
	if s == nil {
		t.Fatalf("struct %s not found", structName)
	}

	out := map[string]*FieldInfo{}
	for i := range s.Fields {
		out[s.Fields[i].Name] = &s.Fields[i]
	}
	return out
}

func TestParseTagsChain(t *testing.T) {
	node, err := ParseTags("required,min=3,max=10", "F")
	if err != nil {
		t.Fatalf("ParseTags: %v", err)
	}

	var got []string
	for n := node; n != nil; n = n.Next {
		got = append(got, n.Tag+"="+n.Param)
	}
	want := []string{"required=", "min=3", "max=10"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("chain = %v, want %v", got, want)
	}
}

func TestParseTagsAliasExpands(t *testing.T) {
	// An alias is reported by name but carries what it expands to.
	node, err := ParseTags("iscolor", "F")
	if err != nil {
		t.Fatalf("ParseTags: %v", err)
	}
	if node.AliasTag != "iscolor" || !node.HasAlias {
		t.Errorf("alias = %q (hasAlias=%v), want iscolor", node.AliasTag, node.HasAlias)
	}
	if got := node.ActualTag; got != AliasExpansions["iscolor"] {
		t.Errorf("actual tag = %q, want the expansion %q", got, AliasExpansions["iscolor"])
	}
}

func TestParseTagsOrBranch(t *testing.T) {
	node, err := ParseTags("rgb|rgba", "F")
	if err != nil {
		t.Fatalf("ParseTags: %v", err)
	}
	if node.Type != TypeOr || node.Next == nil || node.Next.Tag != "rgba" {
		t.Fatalf("expected an or-branch, got %+v", node)
	}
	if got := node.OrExpr(); got != "rgb|rgba" {
		t.Errorf("OrExpr = %q, want rgb|rgba", got)
	}
}

func TestParseTagsEscapes(t *testing.T) {
	// A value that needs a comma or a pipe writes them as the literals 0x2C and
	// 0x7C, which are unescaped after the or-branches are split.
	node, err := ParseTags("oneof=a0x2Cb 0x7C", "F")
	if err != nil {
		t.Fatalf("ParseTags: %v", err)
	}
	if node.Param != "a,b |" {
		t.Errorf("param = %q, want %q", node.Param, "a,b |")
	}
}

func TestParseTagsUnknownTag(t *testing.T) {
	KnownTags = map[string]bool{"required": true}
	defer func() { KnownTags = map[string]bool{} }()

	if _, err := ParseTags("required,nosuchtag", "F"); err == nil {
		t.Fatal("expected an error for an undefined tag")
	}
}

func TestParseTagsDiveKeys(t *testing.T) {
	node, err := ParseTags("dive,keys,min=2,endkeys,required", "M")
	if err != nil {
		t.Fatalf("ParseTags: %v", err)
	}
	if node.Type != TypeDive || node.Next == nil || node.Next.Type != TypeKeys {
		t.Fatalf("dive/keys not parsed: %+v", node)
	}
	if node.Next.Keys == nil || node.Next.Keys.Tag != "min" {
		t.Errorf("keys chain = %+v, want min=2", node.Next.Keys)
	}
}

func TestParseTagsKeysWithoutDive(t *testing.T) {
	if _, err := ParseTags("keys,min=2", "M"); err == nil {
		t.Fatal("expected an error: keys must follow dive")
	}
}

// BuiltinName drives every conversion the generator emits, so it has to tell
// apart types that share a width but differ to reflect.
func TestBuiltinName(t *testing.T) {
	f := fields(t, `package models

type Age int
type Wide int64
type Doc []byte

type M struct {
	S    string
	I    int
	I64  int64
	U8   uint8
	Age  Age
	Wide Wide
	Doc  Doc
	Ms   map[string]int
}
`, "M")

	for name, want := range map[string]string{
		"S": "string", "I": "int", "I64": "int64", "U8": "uint8",
		"Age": "int", "Wide": "int64", "Doc": "", "Ms": "",
	} {
		if got := f[name].Type.BuiltinName(); got != want {
			t.Errorf("%s.BuiltinName() = %q, want %q", name, got, want)
		}
	}

	if !f["Age"].Type.Named || !f["Wide"].Type.Named {
		t.Error("declared types should be marked Named")
	}
	if f["I"].Type.Named {
		t.Error("a predeclared type is not Named")
	}
}

// validator keys an embedded field by its type name and still walks it, so the
// parser has to keep it rather than drop it.
func TestEmbeddedFields(t *testing.T) {
	f := fields(t, `package models

type Base struct {
	Common string `+"`validate:\"required\"`"+`
}

type Outer struct {
	Base
	*PtrBase
	Named  Base   `+"`validate:\"required\"`"+`
	Skip   string `+"`validate:\"-\"`"+`
	hidden string `+"`validate:\"required\"`"+`
}

type PtrBase struct{ Note string }
`, "Outer")

	for _, want := range []string{"Base", "PtrBase", "Named"} {
		if f[want] == nil {
			t.Fatalf("embedded field %s was dropped", want)
		}
	}
	if !f["Base"].Anonymous || !f["PtrBase"].Anonymous {
		t.Error("embedded fields should be marked Anonymous")
	}
	if f["Base"].Type.Kind != KindStruct {
		t.Errorf("Base kind = %v, want struct", f["Base"].Type.Kind)
	}
	if f["PtrBase"].Type.Kind != KindPtr || f["PtrBase"].Type.Elem.Name != "PtrBase" {
		t.Errorf("PtrBase type = %+v, want a pointer to PtrBase", f["PtrBase"].Type)
	}

	// An untagged field stays so it can be recursed into and referenced.
	if f["Named"].Tag != "required" {
		t.Errorf("Named tag = %q, want required", f["Named"].Tag)
	}
	if f["Skip"].Tag != "" || f["Skip"].Tags != nil {
		t.Error(`validate:"-" should leave the field untagged`)
	}
	if f["hidden"] != nil {
		t.Error("unexported fields are skipped unless private field validation is on")
	}
}

func TestSplitParam(t *testing.T) {
	got := SplitParam("red green blue")
	if strings.Join(got, ",") != "red,green,blue" {
		t.Errorf("SplitParam = %v", got)
	}
}
