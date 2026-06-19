// End-to-end tests for the generated petstore client.
// Tests cover: type correctness, validation, request conversion,
// operation registration, helper functions, and edge cases.
package generated

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	runtime "github.com/fred29910/gowasm/pkg/runtime"
)

// ===========================================================================
// 1. Pet schema — type & validation
// ===========================================================================

func TestPet_TypeFields(t *testing.T) {
	t.Run("Name is *string pointer", func(t *testing.T) {
		name := "Fluffy"
		p := Pet{Name: &name}
		if p.Name == nil || *p.Name != "Fluffy" {
			t.Errorf("Name = %v, want &\"Fluffy\"", p.Name)
		}
	})

	t.Run("ID is int64", func(t *testing.T) {
		p := Pet{ID: 42}
		if p.ID != 42 {
			t.Errorf("ID = %d, want 42", p.ID)
		}
	})

	t.Run("Status is string", func(t *testing.T) {
		p := Pet{Status: "available"}
		if p.Status != "available" {
			t.Errorf("Status = %q, want %q", p.Status, "available")
		}
	})
}

func TestPet_Validate(t *testing.T) {
	t.Run("valid Pet passes", func(t *testing.T) {
		name := "Buddy"
		p := Pet{ID: 1, Name: &name, Status: "available"}
		if err := p.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("nil Name fails (required)", func(t *testing.T) {
		p := Pet{ID: 1, Name: nil, Status: "available"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for nil Name, got nil")
		}
	})

	t.Run("invalid Status fails", func(t *testing.T) {
		name := "Buddy"
		p := Pet{ID: 1, Name: &name, Status: "unknown"}
		if err := p.Validate(); err == nil {
			t.Error("expected error for invalid status, got nil")
		}
	})

	t.Run("all valid enum values pass", func(t *testing.T) {
		for _, status := range []string{"available", "pending", "sold"} {
			name := "Pet"
			p := Pet{Name: &name, Status: status}
			if err := p.Validate(); err != nil {
				t.Errorf("status=%q: unexpected error: %v", status, err)
			}
		}
	})

	t.Run("zero ID passes (not required)", func(t *testing.T) {
		name := "Buddy"
		p := Pet{ID: 0, Name: &name, Status: "available"}
		if err := p.Validate(); err != nil {
			t.Errorf("zero ID should pass, got: %v", err)
		}
	})
}

// ===========================================================================
  // 2. CreatePetRequest — ToRequest conversion
// ===========================================================================

func TestCreatePetRequestToRequest(t *testing.T) {
	t.Run("sets method POST and path /pet", func(t *testing.T) {
		params := CreatePetRequest{}
		req := CreatePetRequestToRequest(params)
		if req.Method != "POST" {
			t.Errorf("Method = %q, want %q", req.Method, "POST")
		}
		if req.Path != "/pet" {
			t.Errorf("Path = %q, want %q", req.Path, "/pet")
		}
	})

	t.Run("passes body as Pet struct", func(t *testing.T) {
		name := "Rex"
		body := Pet{ID: 42, Name: &name, Status: "pending"}
		params := CreatePetRequest{Body: body}
		req := CreatePetRequestToRequest(params)

		if req.Body == nil {
			t.Fatal("Body is nil, expected Pet struct")
		}
		bodyPet, ok := req.Body.(Pet)
		if !ok {
			t.Fatalf("Body type = %T, want Pet", req.Body)
		}
		if bodyPet.ID != 42 || *bodyPet.Name != "Rex" || bodyPet.Status != "pending" {
			t.Errorf("Body = %+v, want {ID:42 Name:Rex Status:pending}", bodyPet)
		}
	})

	t.Run("path params are empty (no path params)", func(t *testing.T) {
		req := CreatePetRequestToRequest(CreatePetRequest{})
		if len(req.PathParams) != 0 {
			t.Errorf("PathParams = %v, want empty", req.PathParams)
		}
	})

	t.Run("query is nil when empty", func(t *testing.T) {
		req := CreatePetRequestToRequest(CreatePetRequest{})
		if req.Query != nil {
			t.Errorf("Query = %v, want nil", req.Query)
		}
	})

	t.Run("headers are preserved", func(t *testing.T) {
		params := CreatePetRequest{
			Headers: map[string]string{"Authorization": "Bearer token123"},
		}
		req := CreatePetRequestToRequest(params)
		if req.Headers["Authorization"] != "Bearer token123" {
			t.Errorf("Header Authorization = %q, want %q",
				req.Headers["Authorization"], "Bearer token123")
		}
	})

	t.Run("query passthrough", func(t *testing.T) {
		params := CreatePetRequest{
			Query: url.Values{"extra": {"val"}},
		}
		req := CreatePetRequestToRequest(params)
		if req.Query.Get("extra") != "val" {
			t.Errorf("Query[\"extra\"] = %q, want %q", req.Query.Get("extra"), "val")
		}
	})
}

func TestCreatePetRequest_Validate(t *testing.T) {
	// CreatePetRequest has no required fields; Validate should always pass.
	t.Run("empty request passes", func(t *testing.T) {
		if err := (CreatePetRequest{}).Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

// ===========================================================================
// 3. GetPetByIDRequest — ToRequest conversion
// ===========================================================================

func TestGetPetByIDRequestToRequest(t *testing.T) {
	t.Run("sets method GET and path /pet/{petId}", func(t *testing.T) {
		id := int64(7)
		params := GetPetByIDRequest{PetID: &id}
		req := GetPetByIDRequestToRequest(params)
		if req.Method != "GET" {
			t.Errorf("Method = %q, want %q", req.Method, "GET")
		}
		if req.Path != "/pet/{petId}" {
			t.Errorf("Path = %q, want %q", req.Path, "/pet/{petId}")
		}
	})

	t.Run("sets pathParams petId from PetID pointer", func(t *testing.T) {
		id := int64(99)
		params := GetPetByIDRequest{PetID: &id}
		req := GetPetByIDRequestToRequest(params)
		if got, ok := req.PathParams["petId"]; !ok {
			t.Error("PathParams[\"petId\"] missing")
		} else if got != "99" {
			t.Errorf("PathParams[\"petId\"] = %q, want %q", got, "99")
		}
	})

	t.Run("body is nil (GET has no body)", func(t *testing.T) {
		id := int64(1)
		req := GetPetByIDRequestToRequest(GetPetByIDRequest{PetID: &id})
		if req.Body != nil {
			t.Errorf("Body = %v, want nil", req.Body)
		}
	})

	t.Run("query is nil when empty", func(t *testing.T) {
		id := int64(1)
		req := GetPetByIDRequestToRequest(GetPetByIDRequest{PetID: &id})
		if req.Query != nil {
			t.Errorf("Query = %v, want nil", req.Query)
		}
	})

	t.Run("negative petId", func(t *testing.T) {
		id := int64(-1)
		req := GetPetByIDRequestToRequest(GetPetByIDRequest{PetID: &id})
		if req.PathParams["petId"] != "-1" {
			t.Errorf("PathParams[\"petId\"] = %q, want %q", req.PathParams["petId"], "-1")
		}
	})

	t.Run("max int64 petId", func(t *testing.T) {
		id := int64(9223372036854775807)
		req := GetPetByIDRequestToRequest(GetPetByIDRequest{PetID: &id})
		if req.PathParams["petId"] != "9223372036854775807" {
			t.Errorf("PathParams[\"petId\"] = %q, want %q", req.PathParams["petId"], "9223372036854775807")
		}
	})

	t.Run("user-provided PathParams are NOT merged", func(t *testing.T) {
		id := int64(5)
		params := GetPetByIDRequest{
			PetID:      &id,
			PathParams: map[string]string{"extra": "value"},
		}
		req := GetPetByIDRequestToRequest(params)
		if _, exists := req.PathParams["extra"]; exists {
			t.Error("user-provided PathParams should NOT be merged")
		}
	})
}

func TestGetPetByIDRequest_Validate(t *testing.T) {
	t.Run("valid PetID passes", func(t *testing.T) {
		id := int64(1)
		req := GetPetByIDRequest{PetID: &id}
		if err := req.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("nil PetID fails (required)", func(t *testing.T) {
		req := GetPetByIDRequest{PetID: nil}
		if err := req.Validate(); err == nil {
			t.Error("expected error for nil PetID, got nil")
		}
	})

	t.Run("zero PetID passes (pointer is non-nil)", func(t *testing.T) {
		id := int64(0)
		req := GetPetByIDRequest{PetID: &id}
		if err := req.Validate(); err != nil {
			t.Errorf("pointer-to-zero should pass, got: %v", err)
		}
	})
}

// ===========================================================================
// 4. FindPetsByStatusRequest — ToRequest conversion
// ===========================================================================

func TestFindPetsByStatusRequestToRequest(t *testing.T) {
	t.Run("sets method GET and path /pet/findByStatus", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{})
		if req.Method != "GET" {
			t.Errorf("Method = %q, want %q", req.Method, "GET")
		}
		if req.Path != "/pet/findByStatus" {
			t.Errorf("Path = %q, want %q", req.Path, "/pet/findByStatus")
		}
	})

	t.Run("query passthrough with status filter", func(t *testing.T) {
		params := FindPetsByStatusRequest{
			Query: url.Values{"status": {"available"}},
		}
		req := FindPetsByStatusRequestToRequest(params)
		if req.Query.Get("status") != "available" {
			t.Errorf("Query[\"status\"] = %q, want %q", req.Query.Get("status"), "available")
		}
	})

	t.Run("body is nil (GET has no body)", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{})
		if req.Body != nil {
			t.Errorf("Body = %v, want nil", req.Body)
		}
	})

	t.Run("path params are empty", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{})
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
		if vals[0] != "available" || vals[1] != "pending" {
			t.Errorf("Query[\"status\"] = %v, want [available, pending]", vals)
		}
	})

	t.Run("empty query returns nil", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{})
		if req.Query != nil {
			t.Errorf("Query = %v, want nil", req.Query)
		}
	})

	t.Run("headers passthrough", func(t *testing.T) {
		params := FindPetsByStatusRequest{
			Headers: map[string]string{"X-Trace": "abc123"},
		}
		req := FindPetsByStatusRequestToRequest(params)
		if req.Headers["X-Trace"] != "abc123" {
			t.Errorf("Header X-Trace = %q, want %q", req.Headers["X-Trace"], "abc123")
		}
	})
}

func TestFindPetsByStatusRequest_Validate(t *testing.T) {
	t.Run("empty request passes (all optional)", func(t *testing.T) {
		if err := (FindPetsByStatusRequest{}).Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("with status passes", func(t *testing.T) {
		if err := (FindPetsByStatusRequest{Status: "available"}).Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

// ===========================================================================
// 5. Response types
// ===========================================================================

func TestResponseTypes(t *testing.T) {
	t.Run("CreatePetResponse has Data Pet", func(t *testing.T) {
		name := "Test"
		resp := CreatePetResponse{Data: Pet{Name: &name}}
		if resp.Data.Name == nil || *resp.Data.Name != "Test" {
			t.Errorf("Data.Name = %v, want &\"Test\"", resp.Data.Name)
		}
	})

	t.Run("GetPetByIDResponse has Data Pet", func(t *testing.T) {
		name := "Test"
		resp := GetPetByIDResponse{Data: Pet{Name: &name}}
		if resp.Data.Name == nil || *resp.Data.Name != "Test" {
			t.Errorf("Data.Name = %v, want &\"Test\"", resp.Data.Name)
		}
	})

	t.Run("FindPetsByStatusResponse has Data []Pet", func(t *testing.T) {
		name1 := "A"
		name2 := "B"
		resp := FindPetsByStatusResponse{Data: []Pet{{Name: &name1}, {Name: &name2}}}
		if len(resp.Data) != 2 {
			t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
		}
		if *resp.Data[0].Name != "A" || *resp.Data[1].Name != "B" {
			t.Errorf("Data names = [%q, %q], want [A, B]", *resp.Data[0].Name, *resp.Data[1].Name)
		}
	})
}

// ===========================================================================
// 6. APIClient — operation registration & routing
// ===========================================================================

func TestAPIClient_OperationRegistration(t *testing.T) {
	// NewAPIClient calls registerAll, which registers both to the instance
	// map and to the global runtime registry.
	client := NewAPIClient(runtime.DefaultClientConfig())

	// Verify via instance map
	for _, name := range []string{"createPet", "findPetsByStatus", "getPetById"} {
		t.Run("instance_"+name, func(t *testing.T) {
			handler, ok := client.GetOperation(name)
			if !ok {
				t.Fatalf("operation %q not in instance map", name)
			}
			if handler == nil {
				t.Fatalf("operation %q handler is nil", name)
			}
		})
	}

	// Verify via global runtime registry
	for _, name := range []string{"createPet", "findPetsByStatus", "getPetById"} {
		t.Run("global_"+name, func(t *testing.T) {
			handler, ok := runtime.GetOperation(name)
			if !ok {
				t.Fatalf("operation %q not in global registry", name)
			}
			if handler == nil {
				t.Fatalf("operation %q global handler is nil", name)
			}
		})
	}
}

func TestAPIClient_NewAPIClient(t *testing.T) {
	t.Run("creates client with default config", func(t *testing.T) {
		client := NewAPIClient(runtime.DefaultClientConfig())
		if client == nil {
			t.Fatal("NewAPIClient returned nil")
		}
		if client.client == nil {
			t.Fatal("client.client is nil")
		}
		if client.operations == nil {
			t.Fatal("client.operations is nil")
		}
	})

	t.Run("operations map has 3 entries", func(t *testing.T) {
		client := NewAPIClient(runtime.DefaultClientConfig())
		if len(client.operations) != 3 {
			t.Errorf("len(operations) = %d, want 3", len(client.operations))
		}
	})
}

// ===========================================================================
// 7. OperationHandler — custom registration & invocation
// ===========================================================================

func TestOperationHandler_CustomRegistration(t *testing.T) {
	t.Run("register and retrieve custom handler", func(t *testing.T) {
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

		const opName = "testCustomOp_e2e"
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

		const opName = "testCaptureReq_e2e"
		runtime.RegisterOperation(opName, customHandler)

		handler, ok := runtime.GetOperation(opName)
		if !ok {
			t.Fatal("operation not found")
		}

		input := runtime.Request{Method: "PUT", Path: "/test", Body: "hello"}
		_, err := handler(context.Background(), input)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		if capturedReq.Method != "PUT" || capturedReq.Path != "/test" {
			t.Errorf("captured request = %+v, want {Method:PUT Path:/test}", capturedReq)
		}
	})

	t.Run("nil and empty operation IDs are silently ignored", func(t *testing.T) {
		runtime.RegisterOperation("", func(ctx context.Context, req runtime.Request) (*runtime.Response, error) {
			return &runtime.Response{Status: 200}, nil
		})
		_, ok := runtime.GetOperation("")
		if ok {
			t.Error("expected empty operation ID not to be registered")
		}

		runtime.RegisterOperation("nilHandler_e2e", nil)
		_, ok = runtime.GetOperation("nilHandler_e2e")
		if ok {
			t.Error("expected nil handler not to be registered")
		}
	})
}

// ===========================================================================
// 8. Helper functions — type conversions
// ===========================================================================

func TestHelpers_Int64ToString(t *testing.T) {
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
}

func TestHelpers_IntToString(t *testing.T) {
	cases := []struct {
		input int
		want  string
	}{
		{0, "0"}, {42, "42"}, {-7, "-7"},
	}
	for _, c := range cases {
		if got := intToString(c.input); got != c.want {
			t.Errorf("intToString(%d) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestHelpers_FloatToString(t *testing.T) {
	cases := []struct {
		input float64
		want  string
	}{
		{0, "0"}, {3.14, "3.14"}, {-2.5, "-2.5"},
	}
	for _, c := range cases {
		if got := floatToString(c.input); got != c.want {
			t.Errorf("floatToString(%v) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestHelpers_BoolToString(t *testing.T) {
	if got := boolToString(true); got != "true" {
		t.Errorf("boolToString(true) = %q, want %q", got, "true")
	}
	if got := boolToString(false); got != "false" {
		t.Errorf("boolToString(false) = %q, want %q", got, "false")
	}
}

func TestHelpers_StringToString(t *testing.T) {
	if got := stringToString("hello"); got != "hello" {
		t.Errorf("stringToString(\"hello\") = %q, want %q", got, "hello")
	}
	if got := stringToString(""); got != "" {
		t.Errorf("stringToString(\"\") = %q, want %q", got, "")
	}
}

// ===========================================================================
// 9. Helper functions — copyStringMap & copyValues (deep copy)
// ===========================================================================

func TestCopyStringMap(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		if got := copyStringMap(nil); got != nil {
			t.Errorf("copyStringMap(nil) = %v, want nil", got)
		}
	})

	t.Run("empty map returns nil", func(t *testing.T) {
		if got := copyStringMap(map[string]string{}); got != nil {
			t.Errorf("copyStringMap({}) = %v, want nil", got)
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		input := map[string]string{"a": "1", "b": "2"}
		got := copyStringMap(input)
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		input["a"] = "changed"
		if got["a"] != "1" {
			t.Error("copyStringMap did not produce a deep copy")
		}
	})
}

func TestCopyValues(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		if got := copyValues(nil); got != nil {
			t.Errorf("copyValues(nil) = %v, want nil", got)
		}
	})

	t.Run("empty values returns nil", func(t *testing.T) {
		if got := copyValues(url.Values{}); got != nil {
			t.Errorf("copyValues({}) = %v, want nil", got)
		}
	})

	t.Run("deep copy", func(t *testing.T) {
		input := url.Values{"status": {"available", "pending"}}
		got := copyValues(input)
		input["status"][0] = "sold"
		if got["status"][0] != "available" {
			t.Error("copyValues did not produce a deep copy")
		}
	})
}

// ===========================================================================
// 10. isValid / isValidEnum — internal validation helpers
// ===========================================================================

func TestIsValid(t *testing.T) {
	t.Run("non-nil non-zero values are valid", func(t *testing.T) {
		if !runtime.IsValid(int64(1)) {
			t.Error("IsValid(int64(1)) = false, want true")
		}
		if !runtime.IsValid("hello") {
			t.Error("IsValid(\"hello\") = false, want true")
		}
		name := "x"
		if !runtime.IsValid(Pet{ID: 1, Name: &name}) {
			t.Error("IsValid(Pet{...}) = false, want true")
		}
	})

	t.Run("zero values are invalid", func(t *testing.T) {
		if runtime.IsValid(int64(0)) {
			t.Error("IsValid(int64(0)) = true, want false")
		}
		if runtime.IsValid("") {
			t.Error("IsValid(\"\") = true, want false")
		}
	})

	t.Run("nil is invalid", func(t *testing.T) {
		if runtime.IsValid(nil) {
			t.Error("IsValid(nil) = true, want false")
		}
	})
}

func TestIsValidEnum(t *testing.T) {
	allowed := []interface{}{"available", "pending", "sold"}

	t.Run("valid enum values", func(t *testing.T) {
		for _, v := range allowed {
			if !runtime.IsValidEnum(v, allowed) {
				t.Errorf("IsValidEnum(%q) = false, want true", v)
			}
		}
	})

	t.Run("invalid enum values", func(t *testing.T) {
		for _, v := range []string{"unknown", "", "AVAILABLE"} {
			if runtime.IsValidEnum(v, allowed) {
				t.Errorf("IsValidEnum(%q) = true, want false", v)
			}
		}
	})
}

// ===========================================================================
// 11. *Call functions — verify they exist and are callable (compile check)
// ===========================================================================

func TestCallFunctions_Exist(t *testing.T) {
	client := NewAPIClient(runtime.DefaultClientConfig())
	ctx := context.Background()

	// These require network; just verify they compile and are callable.
	// We skip actual invocation to avoid network dependency.
	t.Run("CreatePetRequestCall exists", func(t *testing.T) {
		_ = func() { _, _ = client.CreatePetRequestCall(ctx, CreatePetRequest{}) }
	})

	t.Run("FindPetsByStatusRequestCall exists", func(t *testing.T) {
		_ = func() { _, _ = client.FindPetsByStatusRequestCall(ctx, FindPetsByStatusRequest{}) }
	})

	t.Run("GetPetByIDRequestCall exists", func(t *testing.T) {
		id := int64(1)
		_ = func() { _, _ = client.GetPetByIDRequestCall(ctx, GetPetByIDRequest{PetID: &id}) }
	})
}

// ===========================================================================
// 12. End-to-end: full request flow
// ===========================================================================

func TestE2E_FullRequestFlow(t *testing.T) {
	t.Run("createPet: build request via ToRequest", func(t *testing.T) {
		name := "Fluffy"
		params := CreatePetRequest{
			Body:    Pet{ID: 1, Name: &name, Status: "available"},
			Headers: map[string]string{"Content-Type": "application/json"},
		}
		req := CreatePetRequestToRequest(params)

		if req.Method != "POST" {
			t.Errorf("Method = %q, want POST", req.Method)
		}
		if req.Path != "/pet" {
			t.Errorf("Path = %q, want /pet", req.Path)
		}
		if req.Body == nil {
			t.Fatal("Body is nil")
		}
		if req.Headers["Content-Type"] != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", req.Headers["Content-Type"])
		}
	})

	t.Run("getPetById: path param substitution", func(t *testing.T) {
		id := int64(42)
		req := GetPetByIDRequestToRequest(GetPetByIDRequest{PetID: &id})
		if req.PathParams["petId"] != "42" {
			t.Errorf("petId = %q, want 42", req.PathParams["petId"])
		}
	})

	t.Run("findPetsByStatus: query passthrough", func(t *testing.T) {
		req := FindPetsByStatusRequestToRequest(FindPetsByStatusRequest{
			Query: url.Values{"status": {"sold"}},
		})
		if req.Query.Get("status") != "sold" {
			t.Errorf("status = %q, want sold", req.Query.Get("status"))
		}
	})
}
