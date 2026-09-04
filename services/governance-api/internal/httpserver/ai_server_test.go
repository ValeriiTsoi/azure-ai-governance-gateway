package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"governance-api/internal/airouter"
	"governance-api/internal/governance"
)

type fakeAIService struct{}

func (f *fakeAIService) Invoke(
	_ context.Context,
	input airouter.InvokeInput,
) (airouter.Result, error) {
	decision := "allow"

	switch input.DataClassification {
	case "confidential":
		decision = "review"
	case "restricted":
		decision = "deny"
	}

	result := airouter.Result{
		Governance: governance.Request{
			RequestID:          "req_http_test",
			CallerSubject:      input.CallerSubject,
			DataClassification: input.DataClassification,
			RequestedModel:     input.RequestedModel,
			PromptHash:         "test-hash",
			PromptChars:        len(input.Prompt),
			Metadata:           map[string]any{},
			Policy: governance.PolicyDecision{
				PolicyName: "test-policy",
				Decision:   decision,
				Reason:     "test decision",
			},
		},
		ProviderCalled: decision == "allow",
	}

	if decision == "allow" {
		result.Route = &airouter.Route{
			RequestedModel: input.RequestedModel,
			RoutedModel:    "mock-fast-general",
			Provider:       "mock",
			Reason:         "test route",
		}

		result.Response = &airouter.ModelResponse{
			Provider: "mock",
			Model:    "mock-fast-general",
			Content:  "test response",
		}

		result.Usage = &airouter.Usage{
			Provider:         "mock",
			Model:            "mock-fast-general",
			InputTokens:      10,
			OutputTokens:     5,
			EstimatedCostUSD: 0,
		}
	}

	return result, nil
}

func newAIServer() *Server {
	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	return NewWithAIRouter(
		logger,
		nil,
		nil,
		&fakeAIService{},
	)
}

func invokeAIRequest(
	t *testing.T,
	server *Server,
	classification string,
) *httptest.ResponseRecorder {
	t.Helper()

	payload := map[string]any{
		"caller_subject":      "test@example.com",
		"cost_center":         "AI-PLATFORM",
		"use_case":            "http-test",
		"data_classification": classification,
		"requested_model":     "fast-general",
		"prompt":              "hello governed AI",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/ai/invoke",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		recorder,
		request,
	)

	return recorder
}

func TestInvokeAIAllowReturns200(
	t *testing.T,
) {
	recorder := invokeAIRequest(
		t,
		newAIServer(),
		"internal",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	var result airouter.Result

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !result.ProviderCalled {
		t.Fatal("expected provider_called=true")
	}

	if result.Route == nil ||
		result.Route.Provider != "mock" {
		t.Fatal("expected mock route")
	}

	if result.Usage == nil {
		t.Fatal("expected usage")
	}
}

func TestInvokeAIReviewReturns202WithoutProvider(
	t *testing.T,
) {
	recorder := invokeAIRequest(
		t,
		newAIServer(),
		"confidential",
	)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status 202, got %d",
			recorder.Code,
		)
	}

	var result airouter.Result

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be called for review")
	}

	if result.Response != nil {
		t.Fatal("response must be absent for review")
	}
}

func TestInvokeAIDenyReturns403WithoutProvider(
	t *testing.T,
) {
	recorder := invokeAIRequest(
		t,
		newAIServer(),
		"restricted",
	)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			recorder.Code,
		)
	}

	var result airouter.Result

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if result.ProviderCalled {
		t.Fatal("provider must not be called for deny")
	}

	if result.Response != nil {
		t.Fatal("response must be absent for deny")
	}
}

func TestInvokeAITrustedCallerOverridesRequestBody(
	t *testing.T,
) {
	server := newAIServer()

	payload := map[string]any{
		"caller_subject":      "spoofed-attacker@example.com",
		"data_classification": "internal",
		"requested_model":     "fast-general",
		"prompt":              "trusted caller test",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/ai/invoke",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	request.Header.Set(
		trustedCallerSubjectHeader,
		"oid:trusted-directory-object",
	)

	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	var result airouter.Result

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if result.Governance.CallerSubject !=
		"oid:trusted-directory-object" {
		t.Fatalf(
			"expected trusted caller subject, got %q",
			result.Governance.CallerSubject,
		)
	}

	if result.Governance.CallerSubject ==
		"spoofed-attacker@example.com" {
		t.Fatal(
			"request body must not override trusted caller identity",
		)
	}
}
