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

type captureOpenAIService struct {
	calls int
	input airouter.InvokeInput
}

func (f *captureOpenAIService) Invoke(
	_ context.Context,
	input airouter.InvokeInput,
) (airouter.Result, error) {
	f.calls++
	f.input = input

	return airouter.Result{
		Governance: governance.Request{
			RequestID:          "req_openai_test",
			CallerSubject:      input.CallerSubject,
			CostCenter:         input.CostCenter,
			DataClassification: input.DataClassification,
			RequestedModel:     input.RequestedModel,
			PromptHash:         "test-hash",
			PromptChars:        len(input.Prompt),
			Policy: governance.PolicyDecision{
				PolicyName: "test-policy",
				Decision:   "allow",
				Reason:     "test allow",
			},
		},
		ProviderCalled: true,
		Response: &airouter.ModelResponse{
			Provider: "mock",
			Model:    "mock-fast-general",
			Content:  "compatibility response",
		},
		Usage: &airouter.Usage{
			Provider:     "mock",
			Model:        "mock-fast-general",
			InputTokens:  11,
			OutputTokens: 4,
		},
	}, nil
}

func newCaptureOpenAIServer(
	service *captureOpenAIService,
) *Server {
	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	return NewWithAIRouter(
		logger,
		nil,
		nil,
		service,
	)
}

func performOpenAIChatRequest(
	t *testing.T,
	server *Server,
	payload map[string]any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		recorder,
		request,
	)

	return recorder
}

func TestOpenAIModelsListsLogicalModel(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/models",
		nil,
	)

	recorder := httptest.NewRecorder()

	newAIServer().Handler().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	var response struct {
		Object string `json:"object"`
		Data   []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Object != "list" {
		t.Fatalf(
			"expected object=list, got %q",
			response.Object,
		)
	}

	if len(response.Data) != 1 {
		t.Fatalf(
			"expected one model, got %d",
			len(response.Data),
		)
	}

	if response.Data[0].ID != openAIExternalModel {
		t.Fatalf(
			"expected model %q, got %q",
			openAIExternalModel,
			response.Data[0].ID,
		)
	}
}

func TestOpenAIChatMapsIntoExistingAIRouter(
	t *testing.T,
) {
	service := &captureOpenAIService{}

	recorder := performOpenAIChatRequest(
		t,
		newCaptureOpenAIServer(service),
		map[string]any{
			"model": openAIExternalModel,
			"messages": []map[string]any{
				{
					"role":    "system",
					"content": "You are concise.",
				},
				{
					"role":    "user",
					"content": "Say hello.",
				},
			},
			"temperature": 0.2,
			"max_tokens":  64,
		},
		map[string]string{
			trustedCallerSubjectHeader: "oid:cursor-demo",
			openAICostCenterHeader:     "DEMO-CURSOR",
			openAIClassificationHeader: "internal",
		},
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if service.calls != 1 {
		t.Fatalf(
			"expected one AI router call, got %d",
			service.calls,
		)
	}

	if service.input.CallerSubject !=
		"oid:cursor-demo" {
		t.Fatalf(
			"unexpected caller subject %q",
			service.input.CallerSubject,
		)
	}

	if service.input.CostCenter !=
		"DEMO-CURSOR" {
		t.Fatalf(
			"unexpected cost center %q",
			service.input.CostCenter,
		)
	}

	if service.input.RequestedModel !=
		openAIInternalModel {
		t.Fatalf(
			"expected internal model %q, got %q",
			openAIInternalModel,
			service.input.RequestedModel,
		)
	}

	if service.input.UseCase !=
		"openai-chat-completions" {
		t.Fatalf(
			"unexpected use case %q",
			service.input.UseCase,
		)
	}

	expectedPrompt :=
		"SYSTEM:\nYou are concise." +
			"\n\nUSER:\nSay hello."

	if service.input.Prompt != expectedPrompt {
		t.Fatalf(
			"unexpected prompt:\n%q",
			service.input.Prompt,
		)
	}

	if service.input.Metadata["external_model"] !=
		openAIExternalModel {
		t.Fatalf(
			"unexpected external model metadata: %#v",
			service.input.Metadata,
		)
	}

	if service.input.Metadata["message_count"] != 2 {
		t.Fatalf(
			"unexpected message count metadata: %#v",
			service.input.Metadata,
		)
	}

	var response openAIChatResponse

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Object != "chat.completion" {
		t.Fatalf(
			"unexpected object %q",
			response.Object,
		)
	}

	if response.Model != openAIExternalModel {
		t.Fatalf(
			"expected external model %q, got %q",
			openAIExternalModel,
			response.Model,
		)
	}

	if len(response.Choices) != 1 ||
		response.Choices[0].Message.Content !=
			"compatibility response" {
		t.Fatalf(
			"unexpected choices: %#v",
			response.Choices,
		)
	}

	if response.Usage.PromptTokens != 11 ||
		response.Usage.CompletionTokens != 4 ||
		response.Usage.TotalTokens != 15 {
		t.Fatalf(
			"unexpected usage: %#v",
			response.Usage,
		)
	}
}

func TestOpenAIChatRejectsStreamingBeforeAIRouter(
	t *testing.T,
) {
	service := &captureOpenAIService{}

	recorder := performOpenAIChatRequest(
		t,
		newCaptureOpenAIServer(service),
		map[string]any{
			"model": openAIExternalModel,
			"messages": []map[string]any{
				{
					"role":    "user",
					"content": "hello",
				},
			},
			"stream": true,
		},
		nil,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			recorder.Code,
		)
	}

	if service.calls != 0 {
		t.Fatal(
			"AI router must not be called for stream=true",
		)
	}
}

func TestOpenAIChatGovernanceAndBudgetStops(
	t *testing.T,
) {
	tests := []struct {
		name           string
		server         *Server
		classification string
		wantStatus     int
	}{
		{
			name:           "governance review",
			server:         newAIServer(),
			classification: "confidential",
			wantStatus:     http.StatusConflict,
		},
		{
			name:           "governance deny",
			server:         newAIServer(),
			classification: "restricted",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "budget review",
			server:         newBudgetAIServer("review"),
			classification: "internal",
			wantStatus:     http.StatusConflict,
		},
		{
			name:           "budget deny",
			server:         newBudgetAIServer("deny"),
			classification: "internal",
			wantStatus:     http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				recorder := performOpenAIChatRequest(
					t,
					test.server,
					map[string]any{
						"model": openAIExternalModel,
						"messages": []map[string]any{
							{
								"role":    "user",
								"content": "hello",
							},
						},
					},
					map[string]string{
						openAICostCenterHeader:     "DEMO-CURSOR",
						openAIClassificationHeader: test.classification,
					},
				)

				if recorder.Code !=
					test.wantStatus {
					t.Fatalf(
						"expected status %d, got %d: %s",
						test.wantStatus,
						recorder.Code,
						recorder.Body.String(),
					)
				}
			},
		)
	}
}
