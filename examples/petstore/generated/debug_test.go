// Debug test file for the generated petstore client.
// This file is in the same package as the generated code (internal tests),
// giving it access to both exported and unexported symbols.
package generated

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	runtime "github.com/fred29910/gowasm/pkg/runtime"
)

// ---------------------------------------------------------------------------
// a. TestCreatePetRequestToRequest
// ---------------------------------------------------------------------------

func TestCreatePetRequestToRequest(t *testing.T) {
	t.Run("sets method and path", func(t *testing.T) {
		params := CreatePetRequest{Body: Pet{ID: 1, Name: "Fluffy", Status: "available"}}
		req := CreatePetRequestToRequest(params)

		if req.Method != "POST" {
			t.Errorf("Method = %q, want %q", req.Method, "POST")
		}
		if req.Path != "/pet" {
			t.Errorf("Path = %q, want %q", req.Path, "/pet")
		}
	})

	t.Run("passes body as Pet struct", func(t *testing.T) {
		p := Pet{ID: 42, Name: "Rex", Status: "pending"}
		params := CreatePetRequest{Body: p}
		req := CreatePetRequestToRequest(params)

		if req.Body == nil {
			t.Fatal("Body is nil, expected Pet struct")
		}
		bodyPet, ok := req.Body.(Pet)
		if !ok {
			t.Fatalf("Body type = %T, want Pet", req.Body)
		}
		if bodyPet.ID != 42 || bodyPet.Name != "Rex" || bodyPet.Status != "pending" {
			t.Errorf("Body = %+v, want {ID:42 Name:Rex Status:pending}", bodyPet)
		}
	})

	t.Run("path params are empty", func(t *testing.T) {
		params := CreatePetRequest{}
		req := CreatePetRequestToRequest(params)

		if len(req.PathParams) != 0 {
			t.Errorf("PathParams = %v, want empty", req.PathParams)
		}
	})

	t.Run("query params are empty", func(t *testing.T) {
		params := CreatePetRequest{}
		req := CreatePetRequestToRequest(params)

		if req.Query != nil {
			t.Errorf("Query = %v, want nil", req.Query)
		}
	})
}

// ---------------------------------------------------------------------------
// b. TestGetPetByIDRequestToRequest
// ---------------------------------------------------------------------------

func TestGetPetByIDRequestToRequest(t *testing.T) {
	t.Run("sets method and path", func(t *testing.T) {
		params := GetPetByIDRequest{PetID: 7}
		req := GetPetByIDRequestToRequest(params)

		if req.Method != "GET" {
			t.Errorf("Method = %q, want %q", req.Method, "GET")
		}
		if req.Path != "/pet/{petId}" {
			t.Errorf("Path = %q, want %q", req.Path, "/pet/{petId}")
		}
	})

	t.Run("sets pathParams petId from PetID", func(t *testing.T) {
		params := GetPetByIDRequest{PetID: 99}
		req := GetPetByIDRequestToRequest(params)

		if got, ok := req.PathParams["petId"]; !ok {
			t.Error("PathParams[\"petId\"] missing")
		} else if got != "99" {
			t.Errorf("PathParams[\"petId\"] = %q, want %q", got, "99")
		}
	})

	t.Run("body is nil", func(t *testing.T) {
		params := GetPetByIDRequest{PetID: 1}
		req := GetPetByIDRequestToRequest(params)

		if req.Body != nil {
			t.Errorf("Body = %v, want nil", req.Body)
		}
	})

	t.Run("query params are empty", func(t *testing.T) {
		params := GetPetByIDRequest{PetID: 1}
		req := GetPetByIDRequestToRequest(params)

		if req.Query != nil {
			t.Errorf("Query = %v, want nil", req.Query)
		}
	})
}

// ---------------------------------------------------------------------------
// c. TestFindPetsByStatusRequestToRequest
// ---------------------------------------------------------------------------

func TestFindPetsByStatusRequestToRequest(t *testing.T) {
	t.Run("sets method and path", func(t *testing.T) {
		params := FindPetsByStatusRequest{}
		req := FindPetsByStatusRequestToRequest(params)

		if req.Method != "GET" {
			t.Errorf("Method = %q, want %q", req.Method, "GET")
		}
		if req.Path != "/pet/findByStatus" {
			t.Errorf("Path = %q, want %q", req.Path, "/pet/findByStatus")
		}
	})

	t.Run("query contains status filter", func(t *testing.T) {
		params := FindPetsByStatusRequest{
			Query: url.Values{"status": {"available"}},
		}
		req := FindPetsByStatusRequestToRequest(params)

		if got := req.Query.Get("status"); got != "available" {
			t.Errorf("Query[\"status\"] = %q, want %q", got, "available")
		}
	})

	t.Run("body is nil", func(t *testing.T) {
		params := FindPetsByStatusRequest{}
		req := FindPetsByStatusRequestToRequest(params)

		if req.Body != nil {
			t.Errorf("Body = %v, want nil", req.Body)
		}
	})

	t.Run("path params are empty", func(t *testing.T) {
		params := FindPetsByStatusRequest{}
		req := FindPetsByStatusRequestToRequest(params)

		if len(req.PathParams) != 0 {
			t.Errorf("PathParams = %v, want empty", req.PathParams)
		}
	})

	t.Run("multiple query values", func(t *testing.T) {
		params := FindPetsByStatusRequest{
			Query: url.Values{"status": {"available", "pending"}},
		}
		req := FindPetsByStatusRequestToRequest(params)

		vals := req.Query["status"]
		if len(vals) != 2 {
			t.Fatalf("Query[\"status\"] has %d values, want 2", len(vals))
		}
		if vals[0] != "available" {
			t.Errorf("Query[\"status\"][0] = %q, want %q", vals[0], "available")
		}
		if vals[1] != "pending" {
			t.Errorf("Query[\"status\"][1] = %q, want %q", vals[1], "pending")
		}
	})
}

// ---------------------------------------------------------------------------
// d. TestCreatePetRequestValidation — Pet struct validation
// ---------------------------------------------------------------------------

func TestCreatePetRequestValidation(t *testing.T) {
	t.Run("valid Pet passes validation", func(t *testing.T) {
		p := Pet{ID: 1, Name: "Buddy", Status: "available"}
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("Pet with empty Name fails validation", func(t *testing.T) {
		p := Pet{ID: 1, Name: "", Status: "available"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty name, got nil")
		} else {
			t.Logf("got expected error: %v", err)
		}
	})

	t.Run("Pet with invalid Status fails validation", func(t *testing.T) {
		p := Pet{ID: 1, Name: "Buddy", Status: "unknown"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid status, got nil")
		} else {
			t.Logf("got expected error: %v", err)
		}
	})

	t.Run("Pet with valid Status available passes", func(t *testing.T) {
		p := Pet{ID: 1, Name: "Buddy", Status: "available"}
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error for status=available, got: %v", err)
		}
	})

	t.Run("Pet with valid Status pending passes", func(t *testing.T) {
		p := Pet{ID: 1, Name: "Buddy", Status: "pending"}
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error for status=pending, got: %v", err)
		}
	})

	t.Run("Pet with valid Status sold passes", func(t *testing.T) {
		p := Pet{ID: 1, Name: "Buddy", Status: "sold"}
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error for status=sold, got: %v", err)
		}
	})

	t.Run("Pet.Validate() — zero ID fails", func(t *testing.T) {
		// int64(0) is the zero value, which isValid() rejects
		p := Pet{ID: 0, Name: "Buddy", Status: "available"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for zero ID, got nil")
		}
	})

	t.Run("CreatePetRequest.Validate() always passes", func(t *testing.T) {
		req := CreatePetRequest{}
		if err := req.Validate(); err != nil {
			t.Errorf("CreatePetRequest.Validate() should always pass, got: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// e. TestGetPetByIDRequestValidation
// ---------------------------------------------------------------------------

func TestGetPetByIDRequestValidation(t *testing.T) {
	t.Run("valid PetID passes", func(t *testing.T) {
		req := GetPetByIDRequest{PetID: 1}
		if err := req.Validate(); err != nil {
			t.Errorf("expected no error for PetID=1, got: %v", err)
		}
	})

	t.Run("zero PetID fails (required)", func(t *testing.T) {
		req := GetPetByIDRequest{PetID: 0}
		if err := req.Validate(); err == nil {
			t.Error("expected error for zero PetID, got nil")
		} else {
			t.Logf("got expected error: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// f. TestFindPetsByStatusRequestValidation
// ---------------------------------------------------------------------------

func TestFindPetsByStatusRequestValidation(t *testing.T) {
	t.Run("empty status passes (optional)", func(t *testing.T) {
		req := FindPetsByStatusRequest{}
		if err := req.Validate(); err != nil {
			t.Errorf("expected no error for empty status, got: %v", err)
		}
	})

	t.Run("with status passes", func(t *testing.T) {
		req := FindPetsByStatusRequest{Status: "available"}
		if err := req.Validate(); err != nil {
			t.Errorf("expected no error with status, got: %v", err)
		}
	})

	t.Run("with query passes", func(t *testing.T) {
		req := FindPetsByStatusRequest{
			Query: url.Values{"status": {"available"}},
		}
		if err := req.Validate(); err != nil {
			t.Errorf("expected no error with query, got: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// g. TestOperationRegistration — verify init() registered all 3 operations
// ---------------------------------------------------------------------------

func TestOperationRegistration(t *testing.T) {
	expected := []string{"createPet", "findPetsByStatus", "getPetById"}

	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			handler, ok := runtime.GetOperation(name)
			if !ok {
				t.Fatalf("operation %q not registered", name)
			}
			if handler == nil {
				t.Fatalf("operation %q handler is nil", name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// h. TestRegisterOperationHandler — custom handler registration
// ---------------------------------------------------------------------------

func TestRegisterOperationHandler(t *testing.T) {
	t.Run("register and retrieve a custom handler", func(t *testing.T) {
		expectedBody := map[string]interface{}{"result": "ok"}
		customHandler := runtime.OperationHandler(
			func(ctx context.Context, req runtime.Request) (*runtime.Response, error) {
				return &runtime.Response{
					Status:  201,
					Headers: map[string]string{"x-custom": "true"},
					Body:    expectedBody,
				}, nil
			},
		)

		const opName = "testCustomOp"
		runtime.RegisterOperation(opName, customHandler)

		handler, ok := runtime.GetOperation(opName)
		if !ok {
			t.Fatal("custom operation not found after registration")
		}

		resp, err := handler(context.Background(), runtime.Request{})
		if err != nil {
			t.Fatalf("handler returned unexpected error: %v", err)
		}
		if resp.Status != 201 {
			t.Errorf("Status = %d, want %d", resp.Status, 201)
		}
		if resp.Headers["x-custom"] != "true" {
			t.Errorf("Header x-custom = %q, want %q", resp.Headers["x-custom"], "true")
		}
		if fmt.Sprintf("%v", resp.Body) != fmt.Sprintf("%v", expectedBody) {
			t.Errorf("Body = %v, want %v", resp.Body, expectedBody)
		}
	})

	t.Run("handler receives the correct request", func(t *testing.T) {
		var capturedReq runtime.Request
		customHandler := runtime.OperationHandler(
			func(ctx context.Context, req runtime.Request) (*runtime.Response, error) {
				capturedReq = req
				return &runtime.Response{Status: 200}, nil
			},
		)

		const opName = "testCaptureReq"
		runtime.RegisterOperation(opName, customHandler)

		handler, ok := runtime.GetOperation(opName)
		if !ok {
			t.Fatal("operation not found")
		}

		input := runtime.Request{
			Method: "PUT",
			Path:   "/test",
			Body:   "hello",
		}
		_, err := handler(context.Background(), input)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		if capturedReq.Method != "PUT" || capturedReq.Path != "/test" {
			t.Errorf("captured request = %+v, want {Method:PUT Path:/test}", capturedReq)
		}
	})

	t.Run("nil and empty operation IDs are silently ignored", func(t *testing.T) {
		// Verify that registering with empty string doesn't panic or register
		runtime.RegisterOperation("", func(ctx context.Context, req runtime.Request) (*runtime.Response, error) {
			return &runtime.Response{Status: 200}, nil
		})
		_, ok := runtime.GetOperation("")
		if ok {
			t.Error("expected empty operation ID not to be registered")
		}

		// nil handler is also ignored
		runtime.RegisterOperation("nilHandler", nil)
		_, ok = runtime.GetOperation("nilHandler")
		if ok {
			t.Error("expected nil handler not to be registered")
		}
	})
}

// ---------------------------------------------------------------------------
// i. TestPetTypeToString — test conversion helper functions
// ---------------------------------------------------------------------------

func TestPetTypeToString(t *testing.T) {
	t.Run("int64ToString", func(t *testing.T) {
		cases := []struct {
			input int64
			want  string
		}{
			{0, "0"},
			{1, "1"},
			{-1, "-1"},
			{9223372036854775807, "9223372036854775807"},
			{-9223372036854775808, "-9223372036854775808"},
		}
		for _, c := range cases {
			got := int64ToString(c.input)
			if got != c.want {
				t.Errorf("int64ToString(%d) = %q, want %q", c.input, got, c.want)
			}
		}
	})

	t.Run("intToString", func(t *testing.T) {
		cases := []struct {
			input int
			want  string
		}{
			{0, "0"},
			{42, "42"},
			{-7, "-7"},
		}
		for _, c := range cases {
			got := intToString(c.input)
			if got != c.want {
				t.Errorf("intToString(%d) = %q, want %q", c.input, got, c.want)
			}
		}
	})

	t.Run("floatToString", func(t *testing.T) {
		cases := []struct {
			input float64
			want  string
		}{
			{0, "0"},
			{3.14, "3.14"},
			{-2.5, "-2.5"},
		}
		for _, c := range cases {
			got := floatToString(c.input)
			if got != c.want {
				t.Errorf("floatToString(%v) = %q, want %q", c.input, got, c.want)
			}
		}
	})

	t.Run("boolToString", func(t *testing.T) {
		if got := boolToString(true); got != "true" {
			t.Errorf("boolToString(true) = %q, want %q", got, "true")
		}
		if got := boolToString(false); got != "false" {
			t.Errorf("boolToString(false) = %q, want %q", got, "false")
		}
	})

	t.Run("stringToString", func(t *testing.T) {
		if got := stringToString("hello"); got != "hello" {
			t.Errorf("stringToString(\"hello\") = %q, want %q", got, "hello")
		}
		if got := stringToString(""); got != "" {
			t.Errorf("stringToString(\"\") = %q, want %q", got, "")
		}
	})
}

// ---------------------------------------------------------------------------
// Helper function validation tests (isValid, isValidEnum)
// ---------------------------------------------------------------------------

func TestIsValid(t *testing.T) {
	t.Run("non-nil non-zero values are valid", func(t *testing.T) {
		if !isValid(int64(1)) {
			t.Error("isValid(int64(1)) = false, want true")
		}
		if !isValid("hello") {
			t.Error("isValid(\"hello\") = false, want true")
		}
		if !isValid(Pet{ID: 1, Name: "x"}) {
			t.Error("isValid(Pet{...}) = false, want true")
		}
	})

	t.Run("zero values are invalid", func(t *testing.T) {
		if isValid(int64(0)) {
			t.Error("isValid(int64(0)) = true, want false")
		}
		if isValid("") {
			t.Error("isValid(\"\") = true, want false")
		}
	})

	t.Run("nil is invalid", func(t *testing.T) {
		if isValid(nil) {
			t.Error("isValid(nil) = true, want false")
		}
	})
}

func TestIsValidEnum(t *testing.T) {
	allowed := []interface{}{"available", "pending", "sold"}

	t.Run("valid enum values", func(t *testing.T) {
		if !isValidEnum("available", allowed) {
			t.Error("isValidEnum(\"available\") = false, want true")
		}
		if !isValidEnum("pending", allowed) {
			t.Error("isValidEnum(\"pending\") = false, want true")
		}
		if !isValidEnum("sold", allowed) {
			t.Error("isValidEnum(\"sold\") = false, want true")
		}
	})

	t.Run("invalid enum values", func(t *testing.T) {
		if isValidEnum("unknown", allowed) {
			t.Error("isValidEnum(\"unknown\") = true, want false")
		}
		if isValidEnum("", allowed) {
			t.Error("isValidEnum(\"\") = true, want false")
		}
	})
}

// ---------------------------------------------------------------------------
// Edge cases and negative tests
// ---------------------------------------------------------------------------

func TestFindPetsByStatusRequestToRequest_QueryPassthrough(t *testing.T) {
	t.Run("preserves empty query", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{})
		if req.Query != nil {
			t.Errorf("Query = %v, want nil", req.Query)
		}
	})

	t.Run("preserves headers from params", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{
			Headers: map[string]string{"X-Trace": "abc123"},
		})
		if req.Headers["X-Trace"] != "abc123" {
			t.Errorf("Header X-Trace = %q, want %q", req.Headers["X-Trace"], "abc123")
		}
	})
}

func TestGetPetByIDRequestToRequest_EdgeCases(t *testing.T) {
	t.Run("minimal int64 value for petId", func(t *testing.T) {
		params := GetPetByIDRequest{PetID: -9223372036854775808}
		req := GetPetByIDRequestToRequest(params)

		if req.PathParams["petId"] != "-9223372036854775808" {
			t.Errorf("PathParams[\"petId\"] = %q, want %q",
				req.PathParams["petId"], "-9223372036854775808")
		}
	})

	t.Run("does not merge user-provided PathParams", func(t *testing.T) {
		// The generated ToRequest creates a fresh map and does NOT copy params.PathParams
		params := GetPetByIDRequest{
			PetID:      5,
			PathParams: map[string]string{"extra": "value"},
		}
		req := GetPetByIDRequestToRequest(params)

		if req.PathParams["petId"] != "5" {
			t.Errorf("PathParams[\"petId\"] = %q, want %q", req.PathParams["petId"], "5")
		}
		if _, exists := req.PathParams["extra"]; exists {
			t.Error("user-provided PathParams should NOT be merged by the generated function")
		}
	})
}

func TestCreatePetRequestToRequest_HeaderPassthrough(t *testing.T) {
	t.Run("headers are preserved", func(t *testing.T) {
		params := CreatePetRequest{
			Body:    Pet{Name: "Test"},
			Headers: map[string]string{"Authorization": "Bearer token123"},
		}
		req := CreatePetRequestToRequest(params)

		if req.Headers["Authorization"] != "Bearer token123" {
			t.Errorf("Header Authorization = %q, want %q",
				req.Headers["Authorization"], "Bearer token123")
		}
	})
}

// ---------------------------------------------------------------------------
// The *Call functions require network — mark with t.Skip in unit tests
// ---------------------------------------------------------------------------

func TestCreatePetRequestCall_RequiresNetwork(t *testing.T) {
	t.Skip("Skipping: TestCreatePetRequestCall requires network access")
	_ = CreatePetRequestCall
}

func TestFindPetsByStatusRequestCall_RequiresNetwork(t *testing.T) {
	t.Skip("Skipping: TestFindPetsByStatusRequestCall requires network access")
	_ = FindPetsByStatusRequestCall
}

func TestGetPetByIDRequestCall_RequiresNetwork(t *testing.T) {
	t.Skip("Skipping: TestGetPetByIDRequestCall requires network access")
	_ = GetPetByIDRequestCall
}

// ---------------------------------------------------------------------------
// Temp directory usage (t.TempDir)
// ---------------------------------------------------------------------------

func TestTempDirUsage(t *testing.T) {
	t.Run("TempDir is accessible", func(t *testing.T) {
		dir := t.TempDir()
		if dir == "" {
			t.Fatal("t.TempDir() returned empty string")
		}
		t.Logf("TempDir: %s", dir)
	})
}

// ---------------------------------------------------------------------------
// copyStringMap and copyValues helper tests
// ---------------------------------------------------------------------------

func TestCopyStringMap(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		if got := copyStringMap(nil); got != nil {
			t.Errorf("copyStringMap(nil) = %v, want nil", got)
		}
	})

	t.Run("returns nil for empty map", func(t *testing.T) {
		if got := copyStringMap(map[string]string{}); got != nil {
			t.Errorf("copyStringMap({}) = %v, want nil", got)
		}
	})

	t.Run("copies all entries", func(t *testing.T) {
		input := map[string]string{"a": "1", "b": "2"}
		got := copyStringMap(input)

		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got["a"] != "1" || got["b"] != "2" {
			t.Errorf("got = %v, want %v", got, input)
		}

		// Ensure it's a deep copy
		input["a"] = "changed"
		if got["a"] != "1" {
			t.Error("copyStringMap did not produce a deep copy")
		}
	})
}

func TestCopyValues(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		if got := copyValues(nil); got != nil {
			t.Errorf("copyValues(nil) = %v, want nil", got)
		}
	})

	t.Run("returns nil for empty values", func(t *testing.T) {
		if got := copyValues(url.Values{}); got != nil {
			t.Errorf("copyValues({}) = %v, want nil", got)
		}
	})

	t.Run("copies all values", func(t *testing.T) {
		input := url.Values{"status": {"available", "pending"}}
		got := copyValues(input)

		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if len(got["status"]) != 2 {
			t.Fatalf("len(status) = %d, want 2", len(got["status"]))
		}
		if got["status"][0] != "available" || got["status"][1] != "pending" {
			t.Errorf("got = %v, want %v", got, input)
		}

		// Ensure deep copy
		input["status"][0] = "sold"
		if got["status"][0] != "available" {
			t.Error("copyValues did not produce a deep copy")
		}
	})
}
