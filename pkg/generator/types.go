package generator

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	splitNameRE  = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	leadingNumRE = regexp.MustCompile(`^[0-9]+`)

	// acronymMap maps lowercase acronyms to their fully capitalized versions
	acronymMap = map[string]string{
		"id":     "ID",
		"url":    "URL",
		"http":   "HTTP",
		"https":  "HTTPS",
		"json":   "JSON",
		"api":    "API",
		"uuid":   "UUID",
		"jwt":    "JWT",
		"html":   "HTML",
		"xml":    "XML",
		"sql":    "SQL",
		"rest":   "REST",
		"grpc":   "gRPC",
		"grpcs":  "gRPCs",
		"tls":    "TLS",
		"ssh":    "SSH",
		"csv":    "CSV",
		"pdf":    "PDF",
		"utf":    "UTF",
		"ascii":  "ASCII",
		"db":     "DB",
		"ios":    "iOS",
		"oauth":  "OAuth",
	}
)

type GeneratedResponse struct {
	Code        string
	GoType      string
	TSType      string
	Description string
	Primary     bool
	StructName  string
}

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
	Responses           []GeneratedResponse
}

type GeneratedParameter struct {
	Name       string
	GoName     string
	TSName     string
	GoType     string
	TSType     string
	In         string
	Required   bool
	EnumValues []interface{}
	Format     string
}

type GeneratedSchema struct {
	Name       string
	GoName     string
	TSName     string
	Properties []GeneratedProperty
}

type GeneratedProperty struct {
	Name       string
	GoName     string
	TSName     string
	GoType     string
	TSType     string
	Required   bool
	EnumValues []interface{}
	Format     string
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
		words := splitCamelCase(p)
		for _, w := range words {
			if w == "" {
				continue
			}
			if acronym, ok := lookupAcronym(w); ok {
				b.WriteString(acronym)
			} else {
				b.WriteString(strings.ToUpper(w[:1]))
				if len(w) > 1 {
					b.WriteString(w[1:])
				}
			}
		}
	}

	result := b.String()
	if result == "" {
		result = "Value"
	}
	if leadingNumRE.MatchString(result) {
		result = "N" + result
	}
	return result
}

// lookupAcronym checks if a word matches an acronym in the map.
// It handles plurals by stripping trailing 's' if the base form is an acronym.
func lookupAcronym(word string) (string, bool) {
	lower := strings.ToLower(word)
	if val, ok := acronymMap[lower]; ok {
		return val, true
	}
	// Handle plurals: try stripping trailing 's'
	if strings.HasSuffix(lower, "s") && len(lower) > 1 {
		base := lower[:len(lower)-1]
		if val, ok := acronymMap[base]; ok {
			return val + "s", true
		}
	}
	return "", false
}

// splitCamelCase splits a string on camelCase and PascalCase boundaries.
// e.g., "userId" → ["user", "Id"], "XMLParser" → ["XML", "Parser"],
// "MyXMLParser" → ["My", "XML", "Parser"]
func splitCamelCase(s string) []string {
	var words []string
	start := 0

	for i := 1; i < len(s); i++ {
		prevUpper := s[i-1] >= 'A' && s[i-1] <= 'Z'
		currUpper := s[i] >= 'A' && s[i] <= 'Z'
		prevLower := s[i-1] >= 'a' && s[i-1] <= 'z'

		// Split on lowercase→uppercase: "userId" → split before 'I'
		if prevLower && currUpper {
			words = append(words, s[start:i])
			start = i
		} else if prevUpper && currUpper && i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z' {
			// Split on uppercase→uppercase→lowercase: "XMLParser" → split before 'P'
			words = append(words, s[start:i])
			start = i
		}
	}

	if start < len(s) {
		words = append(words, s[start:])
	}
	return words
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



