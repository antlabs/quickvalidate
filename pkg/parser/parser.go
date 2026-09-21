package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Kind is the static kind of a field type, mirroring the reflect.Kind subset
// that matters for validation semantics.
type Kind int

const (
	KindInvalid Kind = iota
	KindString
	KindBool
	KindInt
	KindUint
	KindFloat
	KindSlice
	KindArray
	KindMap
	KindPtr
	KindStruct
	KindInterface
)

func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindUint:
		return "uint"
	case KindFloat:
		return "float"
	case KindSlice:
		return "slice"
	case KindArray:
		return "array"
	case KindMap:
		return "map"
	case KindPtr:
		return "ptr"
	case KindStruct:
		return "struct"
	case KindInterface:
		return "interface"
	}
	return "invalid"
}

// TypeInfo describes a field's static type.
type TypeInfo struct {
	// Expr is the rendered Go type expression, e.g. "[]*Address".
	Expr string
	Kind Kind
	// Bits is the width for numeric kinds (8/16/32/64); 0 when not applicable.
	Bits int
	// Name is the type name for named types ("Address", "time.Time").
	Name string
	// Named marks a type declared in the parsed package (or a selector type such
	// as time.Duration). Its Kind and Bits describe the underlying type, but Go
	// still requires an explicit conversion before it can be passed to something
	// expecting the predeclared type.
	Named bool
	// builtin is the predeclared type the value converts to, when it has one:
	// "string" for string and for `type Email string`, "int64" for int64 and for
	// time.Duration. It is empty for the composites. Kind and Bits alone cannot
	// express this — `int` and `int64` share a width but are different kinds to
	// reflect.
	builtin string
	// IsTime / IsDuration mark time.Time and time.Duration, which validator
	// treats specially (comparisons against now / duration params).
	IsTime     bool
	IsDuration bool
	// Elem is the element type for slice/array/ptr, the value type for map.
	Elem *TypeInfo
	// Key is the key type for maps.
	Key *TypeInfo
}

// BuiltinName returns the predeclared type with the same underlying kind and
// width, which is what a value of this type must be converted to before it can
// be handed to something expecting the builtin. It returns "" for types with no
// scalar builtin (slices, maps, structs, pointers, interfaces).
func (t *TypeInfo) BuiltinName() string {
	if t == nil {
		return ""
	}
	if t.builtin != "" {
		return t.builtin
	}

	switch t.Kind {
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindInt:
		switch t.Bits {
		case 8:
			return "int8"
		case 16:
			return "int16"
		case 32:
			return "int32"
		default:
			return "int"
		}
	case KindUint:
		switch t.Bits {
		case 8:
			return "uint8"
		case 16:
			return "uint16"
		case 32:
			return "uint32"
		default:
			return "uint"
		}
	case KindFloat:
		if t.Bits == 32 {
			return "float32"
		}
		return "float64"
	}

	return ""
}

// IsBytes reports whether the type is []byte (validator accepts []byte for the
// file/image tags).
func (t *TypeInfo) IsBytes() bool {
	return t != nil && t.Kind == KindSlice && t.Elem != nil &&
		t.Elem.Kind == KindUint && t.Elem.Bits == 8
}

// FieldInfo describes one struct field carrying a `validate` tag.
type FieldInfo struct {
	Name string
	Type *TypeInfo
	// Tag is the raw `validate` struct tag value.
	Tag string
	// Tags is the parsed tag chain.
	Tags *TagNode
	// Anonymous marks embedded fields.
	Anonymous bool
}

// StructInfo describes a struct that needs a generated validator.
type StructInfo struct {
	Name   string
	Fields []FieldInfo
	// File is the source file the struct was declared in.
	File string
}

// Package is the parsed unit handed to the generator.
type Package struct {
	Name    string
	Structs []*StructInfo
}

// StructByName returns the struct with the given name, or nil.
func (p *Package) StructByName(name string) *StructInfo {
	for _, s := range p.Structs {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// ParseFiles parses the given Go files and returns the package they belong to.
func ParseFiles(paths []string, filter func(*StructInfo) bool) (*Package, error) {
	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(paths))
	for _, path := range paths {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		files = append(files, f)
	}

	pkg := &Package{}
	if len(files) > 0 {
		pkg.Name = files[0].Name.Name
	}

	// First pass: collect every type declaration so that named types can be
	// resolved without full type checking.
	decls := map[string]ast.Expr{}
	for _, f := range files {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					decls[ts.Name.Name] = ts.Type
				}
			}
		}
	}

	r := &resolver{decls: decls, seen: map[string]bool{}}

	for i, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok || typeSpec.Type == nil {
				return true
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			info := &StructInfo{Name: typeSpec.Name.Name, File: paths[i]}
			for _, field := range structType.Fields.List {
				anonymous := len(field.Names) == 0

				var name string
				if anonymous {
					// An embedded field is keyed by its type name, exactly as
					// reflect.StructField.Name reports it, so it lands in the
					// namespace as Embed.Base.Common.
					name = embeddedFieldName(field.Type)
				} else {
					name = field.Names[0].Name
					if !ast.IsExported(name) {
						continue
					}
				}
				if name == "" {
					continue
				}

				fi := FieldInfo{Name: name, Type: r.resolve(field.Type), Anonymous: anonymous}
				if field.Tag != nil {
					tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
					if validateTag := tag.Get("validate"); validateTag != "" && validateTag != "-" {
						chain, err := ParseTags(validateTag, name)
						if err != nil {
							return true // reported by the generator with full context
						}
						fi.Tag, fi.Tags = validateTag, chain
					}
				}

				// Untagged fields are kept: they can be recursion targets and
				// cross field references.
				info.Fields = append(info.Fields, fi)
			}

			if filter == nil || filter(info) {
				pkg.Structs = append(pkg.Structs, info)
			}
			return true
		})
	}

	return pkg, nil
}

// ParseDir parses every non-test Go file in dir (non recursive).
func ParseDir(dir string) (*Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if name := e.Name(); strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	return ParseFiles(paths, nil)
}

type resolver struct {
	decls map[string]ast.Expr
	seen  map[string]bool
}

// resolve turns an AST type expression into a TypeInfo.
func (r *resolver) resolve(expr ast.Expr) *TypeInfo {
	switch t := expr.(type) {
	case *ast.Ident:
		return r.resolveIdent(t.Name)
	case *ast.StarExpr:
		return &TypeInfo{Expr: "*" + r.resolve(t.X).Expr, Kind: KindPtr, Elem: r.resolve(t.X)}
	case *ast.ArrayType:
		elem := r.resolve(t.Elt)
		if t.Len == nil {
			return &TypeInfo{Expr: "[]" + elem.Expr, Kind: KindSlice, Elem: elem}
		}
		return &TypeInfo{Expr: "[N]" + elem.Expr, Kind: KindArray, Elem: elem}
	case *ast.MapType:
		key, val := r.resolve(t.Key), r.resolve(t.Value)
		return &TypeInfo{Expr: fmt.Sprintf("map[%s]%s", key.Expr, val.Expr), Kind: KindMap, Key: key, Elem: val}
	case *ast.SelectorExpr:
		name := r.resolve(t.X).Expr + "." + t.Sel.Name
		return selectorType(name)
	case *ast.StructType:
		return &TypeInfo{Expr: "struct{...}", Kind: KindStruct}
	case *ast.InterfaceType:
		return &TypeInfo{Expr: "interface{}", Kind: KindInterface}
	default:
		return &TypeInfo{Expr: "interface{}", Kind: KindInterface}
	}
}

// embeddedFieldName mirrors reflect.StructField.Name for an embedded field: the
// type name with any pointer or package qualifier stripped.
func embeddedFieldName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return embeddedFieldName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.IndexExpr: // Base[T]
		return embeddedFieldName(t.X)
	case *ast.IndexListExpr: // Base[T1, T2]
		return embeddedFieldName(t.X)
	}
	return ""
}

func (r *resolver) resolveIdent(name string) *TypeInfo {
	switch name {
	case "string":
		return &TypeInfo{Expr: "string", Kind: KindString, Name: "string", builtin: "string"}
	case "bool":
		return &TypeInfo{Expr: "bool", Kind: KindBool, Name: "bool", builtin: "bool"}
	case "int":
		return &TypeInfo{Expr: "int", Kind: KindInt, Bits: 64, Name: "int", builtin: "int"}
	case "int8":
		return &TypeInfo{Expr: "int8", Kind: KindInt, Bits: 8, Name: "int8", builtin: "int8"}
	case "int16":
		return &TypeInfo{Expr: "int16", Kind: KindInt, Bits: 16, Name: "int16", builtin: "int16"}
	case "int32":
		return &TypeInfo{Expr: "int32", Kind: KindInt, Bits: 32, Name: "int32", builtin: "int32"}
	case "int64":
		return &TypeInfo{Expr: "int64", Kind: KindInt, Bits: 64, Name: "int64", builtin: "int64"}
	case "uint":
		return &TypeInfo{Expr: "uint", Kind: KindUint, Bits: 64, Name: "uint", builtin: "uint"}
	case "uint8", "byte":
		return &TypeInfo{Expr: name, Kind: KindUint, Bits: 8, Name: name, builtin: "uint8"}
	case "uint16":
		return &TypeInfo{Expr: "uint16", Kind: KindUint, Bits: 16, Name: "uint16", builtin: "uint16"}
	case "uint32":
		return &TypeInfo{Expr: "uint32", Kind: KindUint, Bits: 32, Name: "uint32", builtin: "uint32"}
	case "uint64":
		return &TypeInfo{Expr: "uint64", Kind: KindUint, Bits: 64, Name: "uint64", builtin: "uint64"}
	case "uintptr":
		return &TypeInfo{Expr: "uintptr", Kind: KindUint, Bits: 64, Name: "uintptr", builtin: "uintptr"}
	case "float32":
		return &TypeInfo{Expr: "float32", Kind: KindFloat, Bits: 32, Name: "float32", builtin: "float32"}
	case "float64":
		return &TypeInfo{Expr: "float64", Kind: KindFloat, Bits: 64, Name: "float64", builtin: "float64"}
	case "interface{}", "any":
		return &TypeInfo{Expr: "interface{}", Kind: KindInterface, Name: name}
	case "error":
		return &TypeInfo{Expr: "error", Kind: KindInterface, Name: "error"}
	}

	decl, ok := r.decls[name]
	if !ok {
		// Unknown identifier: could be a type from a dot-import; treat as struct.
		return &TypeInfo{Expr: name, Kind: KindStruct, Name: name}
	}
	if r.seen[name] {
		return &TypeInfo{Expr: name, Kind: KindStruct, Name: name}
	}
	r.seen[name] = true
	defer delete(r.seen, name)

	underlying := r.resolve(decl)
	out := *underlying
	out.Name = name
	out.Expr = name
	out.Named = true
	out.builtin = underlying.BuiltinName()
	return &out
}

func selectorType(name string) *TypeInfo {
	switch name {
	case "time.Time":
		return &TypeInfo{Expr: name, Kind: KindStruct, Name: name, IsTime: true}
	case "time.Duration":
		// Duration is declared as an int64, which reflect reports as Int64: it
		// compares against int64 fields, not against int ones.
		return &TypeInfo{Expr: name, Kind: KindInt, Bits: 64, Name: name, IsDuration: true, builtin: "int64"}
	}
	return &TypeInfo{Expr: name, Kind: KindStruct, Name: name}
}

// isStructType reports whether validator would recurse into the type.
func isStructType(t *TypeInfo) bool {
	if t == nil {
		return false
	}
	if t.Kind == KindPtr && t.Elem != nil {
		t = t.Elem
	}
	return t.Kind == KindStruct && !t.IsTime
}
