package generator

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	splitNameRE  = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	leadingNumRE = regexp.MustCompile(`^[0-9]+`)
)

type GeneratedOperation struct {
	ID                  string
	TSName              string
	Method              string
	Path                string
	Summary             string
	RequestStructName   string
	RequestInterface    string
	ResponseStructName  string
	ResponseInterface   string
	PathParams          []GeneratedParameter
	QueryParams         []GeneratedParameter
	HasBody             bool
	BodyType            string
	ResponseType        string
	BodyParamName       string
}

type GeneratedParameter struct {
	Name     string
	GoName   string
	TSName   string
	GoType   string
	TSType   string
	In       string
	Required bool
}

type GeneratedSchema struct {
	Name       string
	GoName     string
	TSName     string
	Properties []GeneratedProperty
}

type GeneratedProperty struct {
	Name     string
	GoName   string
	TSName   string
	GoType   string
	TSType   string
	Required bool
}

// ToGoName converts arbitrary names to exported Go identifiers
func ToGoName(name string) string {
	if name == "" {
		return "Value"
	}

	parts := splitNameRE.Split(name, -1)
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}

	result := b.String()
	if result == "" {
		result = "Value"
	}
	if !isASCIIUpper(result[0]) {
		result = string(unicode.ToUpper(rune(result[0]))) + result[1:]
	}
	if leadingNumRE.MatchString(result) {
		result = "N" + result
	}
	return result
}

func ToPrivateGoName(name string) string {
	goName := ToGoName(name)
	if goName == "" {
		return "value"
	}
	return string(unicode.ToLower(rune(goName[0]))) + goName[1:]
}

func ToTSName(name string) string {
	goName := ToGoName(name)
	if goName == "" {
		return "Value"
	}
	return string(unicode.ToLower(rune(goName[0]))) + goName[1:]
}

func isASCIIUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

