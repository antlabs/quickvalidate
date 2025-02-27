package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// TagInfo represents a parsed validation tag
type TagInfo struct {
	Name       string
	Params     string
	IsRequired bool
	IsDive     bool
}

// FieldInfo represents a field with validation tags
type FieldInfo struct {
	Name      string
	Type      string
	Tags      []TagInfo
	IsComplex bool
	IsSlice   bool
	IsMap     bool
	IsPtr     bool
	ElemType  string
}

// StructInfo represents a struct with validation fields
type StructInfo struct {
	Name   string
	Fields []FieldInfo
}

// ParseFile parses a Go source file and extracts struct validation information
func ParseFile(filePath string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structs []StructInfo

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Type == nil {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structInfo := StructInfo{
			Name: typeSpec.Name.Name,
		}

		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}

			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			validateTag := tag.Get("validate")
			if validateTag == "" {
				continue
			}

			fieldType := formatFieldType(field.Type)
			fieldInfo := FieldInfo{
				Name:      field.Names[0].Name,
				Type:      fieldType,
				IsComplex: isComplexType(fieldType),
				IsSlice:   isSliceType(fieldType),
				IsMap:     isMapType(fieldType),
				IsPtr:     isPtrType(fieldType),
				ElemType:  getElemType(fieldType),
			}

			tagParts := strings.Split(validateTag, ",")
			for _, part := range tagParts {
				tagInfo := parseTagPart(part)
				fieldInfo.Tags = append(fieldInfo.Tags, tagInfo)
			}

			structInfo.Fields = append(structInfo.Fields, fieldInfo)
		}

		structs = append(structs, structInfo)
		return true
	})

	return structs, nil
}

// parseTagPart parses a single validation tag part
func parseTagPart(part string) TagInfo {
	tagInfo := TagInfo{}
	
	if part == "required" {
		tagInfo.Name = "required"
		tagInfo.IsRequired = true
		return tagInfo
	}
	
	if part == "dive" {
		tagInfo.Name = "dive"
		tagInfo.IsDive = true
		return tagInfo
	}
	
	if idx := strings.Index(part, "="); idx != -1 {
		tagInfo.Name = part[:idx]
		tagInfo.Params = part[idx+1:]
	} else {
		tagInfo.Name = part
	}
	
	return tagInfo
}

// formatFieldType formats the AST field type as a string
func formatFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatFieldType(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatFieldType(t.Elt)
		}
		return "[N]" + formatFieldType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatFieldType(t.Key), formatFieldType(t.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", formatFieldType(t.X), t.Sel.Name)
	default:
		return "interface{}"
	}
}

// isComplexType checks if the field type is complex (struct, map, slice)
func isComplexType(fieldType string) bool {
	return isSliceType(fieldType) || isMapType(fieldType) || isPtrType(fieldType)
}

// isSliceType checks if the field type is a slice
func isSliceType(fieldType string) bool {
	return strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "[N]")
}

// isMapType checks if the field type is a map
func isMapType(fieldType string) bool {
	return strings.HasPrefix(fieldType, "map[")
}

// isPtrType checks if the field type is a pointer
func isPtrType(fieldType string) bool {
	return strings.HasPrefix(fieldType, "*")
}

// getElemType gets the element type of a complex type
func getElemType(fieldType string) string {
	if isSliceType(fieldType) {
		return strings.TrimPrefix(strings.TrimPrefix(fieldType, "[]"), "[N]")
	}
	
	if isMapType(fieldType) {
		parts := strings.SplitN(fieldType, "]", 2)
		if len(parts) > 1 {
			return parts[1]
		}
	}
	
	if isPtrType(fieldType) {
		return strings.TrimPrefix(fieldType, "*")
	}
	
	return fieldType
}
