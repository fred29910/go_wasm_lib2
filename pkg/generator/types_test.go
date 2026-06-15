package generator

import (
	"testing"
)

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic cases
		{"", "Value"},
		{"name", "Name"},
		{"pet_id", "PetID"},
		{"pet-id", "PetID"},
		{"pet.id", "PetID"},
		{"Pet", "Pet"},
		{"petId", "PetID"},
		{"123abc", "N123abc"},
		{"_private", "Private"},
		// Edge cases: single character, all caps, all lowercase
		{"a", "A"},
		{"z", "Z"},
		{"AB", "AB"},
		{"ABC", "ABC"},
		{"a_b_c", "ABC"},
		{"abc", "Abc"},
		{"ABC_DEF", "ABCDEF"},
		// All acronym mappings
		{"id", "ID"},
		{"url", "URL"},
		{"http", "HTTP"},
		{"https", "HTTPS"},
		{"json", "JSON"},
		{"api", "API"},
		{"uuid", "UUID"},
		{"jwt", "JWT"},
		{"html", "HTML"},
		{"xml", "XML"},
		{"sql", "SQL"},
		{"rest", "REST"},
		{"grpc", "gRPC"},
		{"tls", "TLS"},
		{"ssh", "SSH"},
		{"csv", "CSV"},
		{"pdf", "PDF"},
		{"utf", "UTF"},
		{"ascii", "ASCII"},
		{"db", "DB"},
		{"ios", "iOS"},
		{"oauth", "OAuth"},
		// Standalone acronyms (already uppercase)
		{"ID", "ID"},
		{"URL", "URL"},
		{"HTTP", "HTTP"},
		{"HTTPS", "HTTPS"},
		{"JSON", "JSON"},
		{"API", "API"},
		{"UUID", "UUID"},
		{"JWT", "JWT"},
		{"HTML", "HTML"},
		{"XML", "XML"},
		{"SQL", "SQL"},
		{"REST", "REST"},
		{"gRPC", "GRPC"},
		{"TLS", "TLS"},
		{"SSH", "SSH"},
		{"CSV", "CSV"},
		{"PDF", "PDF"},
		{"UTF", "UTF"},
		{"ASCII", "ASCII"},
		{"DB", "DB"},
		{"iOS", "IOS"},
		{"OAuth", "OAuth"},
		// Acronyms with context
		{"user_id", "UserID"},
		{"userId", "UserID"},
		{"api_key", "APIKey"},
		{"apiKey", "APIKey"},
		{"url_path", "URLPath"},
		{"urlPath", "URLPath"},
		{"http_method", "HTTPMethod"},
		{"httpMethod", "HTTPMethod"},
		{"json_data", "JSONData"},
		{"jsonData", "JSONData"},
		{"uuid_string", "UUIDString"},
		{"uuidString", "UUIDString"},
		{"jwt_token", "JWTToken"},
		{"jwtToken", "JWTToken"},
		{"html_content", "HTMLContent"},
		{"xml_parser", "XMLParser"},
		{"sql_query", "SQLQuery"},
		{"rest_api", "RESTAPI"},
		{"grpc_service", "gRPCService"},
		{"grpcService", "gRPCService"},
		{"tls_config", "TLSConfig"},
		{"ssh_key", "SSHKey"},
		{"csv_file", "CSVFile"},
		{"pdf_document", "PDFDocument"},
		{"utf_encoding", "UTFEncoding"},
		{"ascii_art", "ASCIIArt"},
		{"db_connection", "DBConnection"},
		{"ios_app", "iOSApp"},
		{"oauth_token", "OAuthToken"},
		// Plurals of acronyms
		{"ids", "IDs"},
		{"urls", "URLs"},
		{"apis", "APIs"},
		{"https", "HTTPS"},
		{"jsons", "JSONs"},
		{"uuids", "UUIDs"},
		{"jwts", "JWTs"},
		{"htmls", "HTMLs"},
		{"xmls", "XMLs"},
		{"sqls", "SQLs"},
		{"csvs", "CSVs"},
		{"pdfs", "PDFs"},
		// Mixed acronyms
		{"user_urls", "UserURLs"},
		{"userIdUrl", "UserIDURL"},
		{"xml_id", "XMLID"},
		{"xmlParser", "XMLParser"},
		{"xmlParserId", "XMLParserID"},
		{"userIdApi", "UserIDAPI"},
		{"XMLParser", "XMLParser"},
		{"MyXMLParser", "MyXMLParser"},
		{"myXMLParserID", "MyXMLParserID"},
		{"apiEndpointUrl", "APIEndpointURL"},
		// Complex mixed cases
		{"getUserID", "GetUserID"},
		{"parseJSONData", "ParseJSONData"},
		{"sendHTTPRequest", "SendHTTPRequest"},
		{"validateJWTToken", "ValidateJWTToken"},
		{"fetchUserURLs", "FetchUserURLs"},
		{"loadCSVFile", "LoadCSVFile"},
		{"renderHTMLPage", "RenderHTMLPage"},
		{"parseXMLDocument", "ParseXMLDocument"},
		{"connectToDB", "ConnectToDB"},
		{"generatePDFReport", "GeneratePDFReport"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToGoName(tt.input)
			if got != tt.expected {
				t.Errorf("ToGoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToTSName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "value"},
		{"name", "name"},
		{"pet_id", "petID"},
		{"pet-id", "petID"},
		{"Pet", "pet"},
		{"createPet", "createPet"},
		{"user_id", "userID"},
		{"apiKey", "aPIKey"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToTSName(tt.input)
			if got != tt.expected {
				t.Errorf("ToTSName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPrivateGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "value"},
		{"Name", "name"},
		{"PetId", "petID"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPrivateGoName(tt.input)
			if got != tt.expected {
				t.Errorf("ToPrivateGoName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
