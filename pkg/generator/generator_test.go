package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antlabs/quickvalidate/pkg/parser"
)

// parseSource writes src to a throwaway package and parses it.
func parseSource(t *testing.T, src string) *parser.Package {
	t.Helper()

	path := filepath.Join(t.TempDir(), "models.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	pkg, err := parser.ParseFiles([]string{path}, nil)
	if err != nil {
		t.Fatalf("ParseFiles: %v", err)
	}
	return pkg
}

// The numeric codes are read as strings or as integers depending on the field
// kind, mirroring the reference implementation: every accepted kind has to
// generate code that compiles, and the unsigned branches must keep the
// reference's cast semantics.
func TestGenerateNumericCodeKinds(t *testing.T) {
	code, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
		"\tStr  string `validate:\"iso3166_1_alpha_numeric\"`\n"+
		"\tInt  int    `validate:\"iso3166_1_alpha_numeric\"`\n"+
		"\tUint uint16 `validate:\"iso3166_1_alpha_numeric\"`\n"+
		"\tCurr uint16 `validate:\"iso4217_numeric\"`\n"+
		"}\n")).Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got := string(code)
	for _, want := range []string{
		"qv.IsIso3166AlphaNumeric(s.Str)",
		"qv.IsIso3166AlphaNumericInt(int64(s.Int))",
		"qv.IsIso3166AlphaNumericUint(uint64(s.Uint))",
		"qv.IsIso4217NumericUint(uint64(s.Curr))",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("generated code is missing %q", want)
		}
	}
}

// A string field for iso4217_numeric makes the reference implementation panic,
// so the generator refuses it instead of emitting code that will not compile.
func TestGenerateRejectsUnsupportedNumericKind(t *testing.T) {
	_, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
		"\tCurr string `validate:\"iso4217_numeric\"`\n"+
		"}\n")).Generate()
	if err == nil {
		t.Fatal("expected a generation error for iso4217_numeric on a string field")
	}
	if !strings.Contains(err.Error(), "iso4217_numeric") {
		t.Errorf("the error should name the tag, got: %v", err)
	}
}

// A string only tag on a field that is not a string makes the reference fail
// forever at run time; the generator refuses it instead of emitting code that
// does not compile.
func TestGenerateRejectsStringTagsOnOtherKinds(t *testing.T) {
	for _, tc := range []struct{ field, tag string }{
		{"Email int", `validate:"email"`},
		{"Age int", `validate:"gte=0,datetime=2006-01-02"`},
		{"Code int", `validate:"contains=x"`},
		{"Rel int", `validate:"spicedb=id"`},
		{"ID int", `validate:"uuid"`},
		{"Hex int", `validate:"hexadecimal"`},
	} {
		_, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
			"\t"+tc.field+" `"+tc.tag+"`\n"+
			"}\n")).Generate()
		if err == nil {
			t.Errorf("%s with %s: expected a generation error", tc.field, tc.tag)
			continue
		}
		if !strings.Contains(err.Error(), "string field") {
			t.Errorf("%s with %s: error should say a string field is required, got: %v",
				tc.field, tc.tag, err)
		}
	}
}

// json is the one string tag the reference also reads from a byte slice.
func TestGenerateJSONKinds(t *testing.T) {
	code, err := NewGenerator(parseSource(t, "package models\n\ntype Named []byte\n\ntype M struct {\n"+
		"\tStr string `validate:\"json\"`\n"+
		"\tDoc []byte `validate:\"json\"`\n"+
		"\tRaw Named  `validate:\"json\"`\n"+
		"}\n")).Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got := string(code)
	for _, want := range []string{
		"qv.IsJSON(s.Str)",
		"qv.IsJSONBytes(s.Doc)",
		"qv.IsJSONBytes([]byte(s.Raw))",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("generated code is missing %q", want)
		}
	}
}

// An interface field only says whether it is set, so the tags that depend on the
// dynamic value cannot be emitted — and silently skipping them would be worse
// than refusing.
func TestGenerateRejectsValueTagsOnInterface(t *testing.T) {
	_, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
		"\tAny interface{} `validate:\"min=3\"`\n"+
		"}\n")).Generate()
	if err == nil {
		t.Fatal("expected a generation error for min on an interface field")
	}
	for _, want := range []string{"Any", "min"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should mention %q, got: %v", want, err)
		}
	}
}

func TestGenerateInterfacePresenceTags(t *testing.T) {
	code, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
		"\tReq  interface{} `validate:\"required\"`\n"+
		"\tDef  interface{} `validate:\"isdefault\"`\n"+
		"\tCond interface{} `validate:\"required_if=Req x\"`\n"+
		"\tOdd  interface{}\n"+
		"}\n")).Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got := string(code)
	if !strings.Contains(got, "s.Req != nil") {
		t.Error("required on an interface should check for nil")
	}
	// validator returns early for an untagged interface field.
	for _, field := range []string{"Odd"} {
		if strings.Contains(got, "s."+field+" ") || strings.Contains(got, "s."+field+")") {
			t.Errorf("untagged interface field %s should not be emitted", field)
		}
	}
}

// spicedb only understands a fixed set of parameters and panics on anything
// else, so an unknown one is caught while generating instead.
func TestGenerateRejectsUnknownSpicedbParam(t *testing.T) {
	_, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
		"\tRel string `validate:\"spicedb=bogus\"`\n"+
		"}\n")).Generate()
	if err == nil {
		t.Fatal("expected a generation error for an unknown spicedb parameter")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("the error should name the parameter, got: %v", err)
	}
}

func TestGenerateAcceptsKnownSpicedbParams(t *testing.T) {
	code, err := NewGenerator(parseSource(t, "package models\n\ntype M struct {\n"+
		"\tRel string `validate:\"spicedb=permission\"`\n"+
		"\tID  string `validate:\"spicedb=id\"`\n"+
		"}\n")).Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if n := strings.Count(string(code), "qv.IsSpiceDB("); n != 2 {
		t.Errorf("expected both fields to be validated, got %d calls", n)
	}
}

// init() runs per file in file name order, so a duplicate registration silently
// shadows the earlier emitter — that is how iso4217_numeric lost its integer
// aware emitter. Registering twice has to be loud.
func TestRegisterRejectsDuplicates(t *testing.T) {
	const tag = "duplicate_registration_test_tag"
	defer delete(registry, tag)

	register(tag, tagSpec{})

	defer func() {
		if recover() == nil {
			t.Error("registering the same tag twice should panic")
		}
	}()
	register(tag, tagSpec{})
}
