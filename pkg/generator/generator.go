package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/antlabs/quickvalidate/pkg/parser"
)

// Generator generates validation code
type Generator struct {
	PackageName string
	Structs     []parser.StructInfo
	Imports     map[string]bool
}

// NewGenerator creates a new Generator
func NewGenerator(packageName string, structs []parser.StructInfo) *Generator {
	return &Generator{
		PackageName: packageName,
		Structs:     structs,
		Imports:     make(map[string]bool),
	}
}

// Generate generates validation code
func (g *Generator) Generate() ([]byte, error) {
	var buf bytes.Buffer

	// Add package declaration
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.PackageName))

	// Add imports
	g.addRequiredImports()
	buf.WriteString("import (\n")
	for imp := range g.Imports {
		buf.WriteString(fmt.Sprintf("\t%q\n", imp))
	}
	buf.WriteString(")\n\n")

	// Generate validation functions for each struct
	for _, structInfo := range g.Structs {
		validatorCode, err := g.generateStructValidator(structInfo)
		if err != nil {
			return nil, err
		}
		buf.WriteString(validatorCode)
		buf.WriteString("\n\n")
	}

	// Format the generated code
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("error formatting generated code: %w", err)
	}

	return formattedCode, nil
}

// addRequiredImports adds required imports
func (g *Generator) addRequiredImports() {
	g.Imports["github.com/antlabs/quickvalidate/pkg/errors"] = true
	g.Imports["fmt"] = true
	g.Imports["regexp"] = true
	g.Imports["strings"] = true
	g.Imports["reflect"] = true
	g.Imports["net/url"] = true
}

// generateStructValidator generates validation code for a struct
func (g *Generator) generateStructValidator(structInfo parser.StructInfo) (string, error) {
	tmpl := template.Must(template.New("structValidator").Parse(`
// Validate validates the {{ .Name }} struct
func (s *{{ .Name }}) Validate() error {
	var errs errors.ValidationErrors

	{{ range .Fields }}
	{{ $fieldName := .Name }}
	{{ range .Tags }}
	{{ if eq .Name "required" }}
	// Validate {{ $fieldName }} is required
	if {{ $.GenerateRequiredCheck $fieldName .Type }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "required",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "email" }}
	// Validate {{ $fieldName }} is a valid email
	if s.{{ $fieldName }} != "" && !{{ $.GenerateEmailCheck $fieldName }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "email",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "min" }}
	// Validate {{ $fieldName }} is at least {{ .Params }}
	if {{ $.GenerateMinCheck $fieldName .Type .Params }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "min",
			Param: "{{ .Params }}",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "max" }}
	// Validate {{ $fieldName }} is at most {{ .Params }}
	if {{ $.GenerateMaxCheck $fieldName .Type .Params }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "max",
			Param: "{{ .Params }}",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "gt" }}
	// Validate {{ $fieldName }} is greater than {{ .Params }}
	if {{ $.GenerateGtCheck $fieldName .Type .Params }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "gt",
			Param: "{{ .Params }}",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "gte" }}
	// Validate {{ $fieldName }} is greater than or equal to {{ .Params }}
	if {{ $.GenerateGteCheck $fieldName .Type .Params }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "gte",
			Param: "{{ .Params }}",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "oneof" }}
	// Validate {{ $fieldName }} is one of {{ .Params }}
	if {{ $.GenerateOneOfCheck $fieldName .Params }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "oneof",
			Param: "{{ .Params }}",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "url" }}
	// Validate {{ $fieldName }} is a valid URL
	if s.{{ $fieldName }} != "" && !{{ $.GenerateURLCheck $fieldName }} {
		errs = append(errs, errors.ValidationError{
			Field: "{{ $fieldName }}",
			Tag: "url",
			Value: s.{{ $fieldName }},
			Namespace: "{{ $.Name }}.{{ $fieldName }}",
		})
	}
	{{ else if eq .Name "dive" }}
	// Validate {{ $fieldName }} elements
	{{ if .IsSlice }}
	for i, elem := range s.{{ $fieldName }} {
		if elem != nil {
			if err := elem.Validate(); err != nil {
				if valErrs, ok := err.(errors.ValidationErrors); ok {
					for _, valErr := range valErrs {
						valErr.Namespace = fmt.Sprintf("{{ $.Name }}.{{ $fieldName }}[%d].%s", i, valErr.Field)
						errs = append(errs, valErr)
					}
				}
			}
		}
	}
	{{- else if .IsMap }}
	for key, elem := range s.{{ $fieldName }} {
		if elem != nil {
			if err := elem.Validate(); err != nil {
				if valErrs, ok := err.(errors.ValidationErrors); ok {
					for _, valErr := range valErrs {
						valErr.Namespace = fmt.Sprintf("{{ $.Name }}.{{ $fieldName }}[%v].%s", key, valErr.Field)
						errs = append(errs, valErr)
					}
				}
			}
		}
	}
	{{- end }}
	{{- end }}
	{{- end }}
	{{- end }}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
`))

	funcMap := template.FuncMap{
		"GenerateRequiredCheck": func(fieldName, fieldType string) string {
			switch {
			case strings.HasPrefix(fieldType, "string"):
				return fmt.Sprintf("s.%s == \"\"", fieldName)
			case strings.HasPrefix(fieldType, "int") || strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "float"):
				return fmt.Sprintf("s.%s == 0", fieldName)
			case strings.HasPrefix(fieldType, "bool"):
				return fmt.Sprintf("!s.%s", fieldName)
			case strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "map["):
				return fmt.Sprintf("len(s.%s) == 0", fieldName)
			case strings.HasPrefix(fieldType, "*"):
				return fmt.Sprintf("s.%s == nil", fieldName)
			default:
				return fmt.Sprintf("reflect.ValueOf(s.%s).IsZero()", fieldName)
			}
		},
		"GenerateEmailCheck": func(fieldName string) string {
			g.Imports["regexp"] = true
			return fmt.Sprintf("emailRegex.MatchString(s.%s)", fieldName)
		},
		"GenerateMinCheck": func(fieldName, fieldType, minVal string) string {
			switch {
			case strings.HasPrefix(fieldType, "string"):
				return fmt.Sprintf("len(s.%s) < %s", fieldName, minVal)
			case strings.HasPrefix(fieldType, "int") || strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "float"):
				return fmt.Sprintf("s.%s < %s", fieldName, minVal)
			case strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "map["):
				return fmt.Sprintf("len(s.%s) < %s", fieldName, minVal)
			default:
				return "false"
			}
		},
		"GenerateMaxCheck": func(fieldName, fieldType, maxVal string) string {
			switch {
			case strings.HasPrefix(fieldType, "string"):
				return fmt.Sprintf("len(s.%s) > %s", fieldName, maxVal)
			case strings.HasPrefix(fieldType, "int") || strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "float"):
				return fmt.Sprintf("s.%s > %s", fieldName, maxVal)
			case strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "map["):
				return fmt.Sprintf("len(s.%s) > %s", fieldName, maxVal)
			default:
				return "false"
			}
		},
		"GenerateGtCheck": func(fieldName, fieldType, gtVal string) string {
			switch {
			case strings.HasPrefix(fieldType, "string"):
				return fmt.Sprintf("len(s.%s) <= %s", fieldName, gtVal)
			case strings.HasPrefix(fieldType, "int") || strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "float"):
				return fmt.Sprintf("s.%s <= %s", fieldName, gtVal)
			case strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "map["):
				return fmt.Sprintf("len(s.%s) <= %s", fieldName, gtVal)
			default:
				return "false"
			}
		},
		"GenerateGteCheck": func(fieldName, fieldType, gteVal string) string {
			switch {
			case strings.HasPrefix(fieldType, "string"):
				return fmt.Sprintf("len(s.%s) < %s", fieldName, gteVal)
			case strings.HasPrefix(fieldType, "int") || strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "float"):
				return fmt.Sprintf("s.%s < %s", fieldName, gteVal)
			case strings.HasPrefix(fieldType, "[]") || strings.HasPrefix(fieldType, "map["):
				return fmt.Sprintf("len(s.%s) < %s", fieldName, gteVal)
			default:
				return "false"
			}
		},
		"GenerateOneOfCheck": func(fieldName, values string) string {
			g.Imports["strings"] = true
			return fmt.Sprintf("!isOneOf(s.%s, []string{%s})", fieldName, formatOneOfValues(values))
		},
		"GenerateURLCheck": func(fieldName string) string {
			g.Imports["net/url"] = true
			return fmt.Sprintf("isValidURL(s.%s)", fieldName)
		},
	}

	tmpl = tmpl.Funcs(funcMap)

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		Name                 string
		Fields               []parser.FieldInfo
		GenerateRequiredCheck func(string, string) string
		GenerateEmailCheck    func(string) string
		GenerateMinCheck      func(string, string, string) string
		GenerateMaxCheck      func(string, string, string) string
		GenerateGtCheck       func(string, string, string) string
		GenerateGteCheck      func(string, string, string) string
		GenerateOneOfCheck    func(string, string) string
		GenerateURLCheck      func(string) string
	}{
		Name:                 structInfo.Name,
		Fields:               structInfo.Fields,
		GenerateRequiredCheck: funcMap["GenerateRequiredCheck"].(func(string, string) string),
		GenerateEmailCheck:    funcMap["GenerateEmailCheck"].(func(string) string),
		GenerateMinCheck:      funcMap["GenerateMinCheck"].(func(string, string, string) string),
		GenerateMaxCheck:      funcMap["GenerateMaxCheck"].(func(string, string, string) string),
		GenerateGtCheck:       funcMap["GenerateGtCheck"].(func(string, string, string) string),
		GenerateGteCheck:      funcMap["GenerateGteCheck"].(func(string, string, string) string),
		GenerateOneOfCheck:    funcMap["GenerateOneOfCheck"].(func(string, string) string),
		GenerateURLCheck:      funcMap["GenerateURLCheck"].(func(string) string),
	})
	if err != nil {
		return "", err
	}

	// Add helper functions
	helpers := `
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isOneOf(value string, allowedValues []string) bool {
	for _, v := range allowedValues {
		if value == v {
			return true
		}
	}
	return false
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
`

	return buf.String() + helpers, nil
}

// formatOneOfValues formats the values for the oneof validator
func formatOneOfValues(values string) string {
	parts := strings.Split(values, " ")
	for i, part := range parts {
		parts[i] = fmt.Sprintf("%q", part)
	}
	return strings.Join(parts, ", ")
}
