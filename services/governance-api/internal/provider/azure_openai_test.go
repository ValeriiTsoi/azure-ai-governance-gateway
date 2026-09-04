package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func TestAzureOpenAIInvoke(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.Method != http.MethodPost {
					t.Fatalf(
						"expected POST, got %s",
						r.Method,
					)
				}

				if r.URL.Path !=
					"/openai/v1/responses" {
					t.Fatalf(
						"unexpected path: %s",
						r.URL.Path,
					)
				}

				var body map[string]any

				if err := json.NewDecoder(
					r.Body,
				).Decode(&body); err != nil {
					t.Fatalf(
						"decode request: %v",
						err,
					)
				}

				if body["model"] != "gpt-5-mini" {
					t.Fatalf(
						"unexpected model: %v",
						body["model"],
					)
				}

				if body["input"] !=
					"Explain governed AI." {
					t.Fatalf(
						"unexpected input: %v",
						body["input"],
					)
				}

				if body["store"] != false {
					t.Fatalf(
						"expected store=false, got %v",
						body["store"],
					)
				}

				if body["max_output_tokens"] !=
					float64(
						azureOpenAIMaxOutputTokens,
					) {
					t.Fatalf(
						"unexpected max_output_tokens: %v",
						body["max_output_tokens"],
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write([]byte(`{
					"id": "resp_test",
					"object": "response",
					"created_at": 1,
					"status": "completed",
					"model": "gpt-5-mini",
					"output": [
						{
							"id": "msg_test",
							"type": "message",
							"status": "completed",
							"role": "assistant",
							"content": [
								{
									"type": "output_text",
									"text": "Governed AI response.",
									"annotations": []
								}
							]
						}
					],
					"usage": {
						"input_tokens": 7,
						"output_tokens": 11,
						"total_tokens": 18,
						"input_tokens_details": {
							"cached_tokens": 0
						},
						"output_tokens_details": {
							"reasoning_tokens": 4
						}
					}
				}`))
			},
		),
	)
	defer server.Close()

	client := openai.NewClient(
		option.WithBaseURL(
			server.URL+"/openai/v1/",
		),
		option.WithAPIKey("unit-test"),
	)

	azureProvider := newAzureOpenAIWithClient(
		client,
	)

	response, err := azureProvider.Invoke(
		context.Background(),
		InvokeRequest{
			Model:  "gpt-5-mini",
			Prompt: "Explain governed AI.",
		},
	)
	if err != nil {
		t.Fatalf(
			"invoke Azure OpenAI: %v",
			err,
		)
	}

	if azureProvider.Name() != "azure-openai" {
		t.Fatalf(
			"unexpected provider name: %s",
			azureProvider.Name(),
		)
	}

	if response.Content !=
		"Governed AI response." {
		t.Fatalf(
			"unexpected content: %q",
			response.Content,
		)
	}

	if response.Model != "gpt-5-mini" {
		t.Fatalf(
			"unexpected model: %s",
			response.Model,
		)
	}

	if response.Usage.InputTokens != 7 {
		t.Fatalf(
			"unexpected input tokens: %d",
			response.Usage.InputTokens,
		)
	}

	if response.Usage.OutputTokens != 11 {
		t.Fatalf(
			"unexpected output tokens: %d",
			response.Usage.OutputTokens,
		)
	}

	if response.Usage.EstimatedCostUSD != 0 {
		t.Fatalf(
			"unexpected estimated cost: %f",
			response.Usage.EstimatedCostUSD,
		)
	}
}

func TestAzureOpenAIRejectsEmptyPrompt(
	t *testing.T,
) {
	client := openai.NewClient(
		option.WithAPIKey("unit-test"),
	)

	azureProvider := newAzureOpenAIWithClient(
		client,
	)

	_, err := azureProvider.Invoke(
		context.Background(),
		InvokeRequest{
			Model: "gpt-5-mini",
		},
	)

	if err == nil {
		t.Fatal(
			"expected empty prompt error",
		)
	}
}

func TestAzureOpenAIManagedIdentityRequiresClientID(
	t *testing.T,
) {
	_, err := NewAzureOpenAIManagedIdentity(
		"https://example.openai.azure.com/",
		"",
	)

	if err == nil {
		t.Fatal(
			"expected missing managed identity client ID error",
		)
	}
}
