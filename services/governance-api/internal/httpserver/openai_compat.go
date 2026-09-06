package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"governance-api/internal/airouter"
)

const (
	openAIExternalModel        = "aigov-fast-general"
	openAIInternalModel        = "fast-general"
	openAICostCenterHeader     = "X-AIGOV-Cost-Center"
	openAIClassificationHeader = "X-AIGOV-Data-Classification"
)

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type openAIChatRequest struct {
	Model      string              `json:"model"`
	Messages   []openAIChatMessage `json:"messages"`
	Stream     bool                `json:"stream,omitempty"`
	Tools      json.RawMessage     `json:"tools,omitempty"`
	ToolChoice json.RawMessage     `json:"tool_choice,omitempty"`
}

type openAIChatChoice struct {
	Index        int               `json:"index"`
	Message      openAIChatMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type openAIChatResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []openAIChatChoice `json:"choices"`
	Usage   openAIUsage        `json:"usage"`
}

func (s *Server) listOpenAIModels(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(w, http.StatusOK, map[string]any{
		"object": "list",
		"data": []map[string]any{
			{
				"id":       openAIExternalModel,
				"object":   "model",
				"created":  int64(0),
				"owned_by": "aigov",
			},
		},
	})
}

func (s *Server) createOpenAIChatCompletion(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	decoder := json.NewDecoder(r.Body)

	var request openAIChatRequest

	if err := decoder.Decode(&request); err != nil {
		writeOpenAICompatError(
			w,
			http.StatusBadRequest,
			"invalid JSON request body",
			"invalid_request_error",
			"invalid_json",
		)
		return
	}

	var extra any

	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeOpenAICompatError(
			w,
			http.StatusBadRequest,
			"request body must contain exactly one JSON object",
			"invalid_request_error",
			"invalid_json",
		)
		return
	}

	if request.Model != openAIExternalModel {
		writeOpenAICompatError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"unsupported model %q",
				request.Model,
			),
			"invalid_request_error",
			"model_not_found",
		)
		return
	}

	if request.Stream {
		writeOpenAICompatError(
			w,
			http.StatusBadRequest,
			"stream=true is not supported yet",
			"invalid_request_error",
			"stream_not_supported",
		)
		return
	}

	if hasOpenAIJSONValue(request.Tools) ||
		hasOpenAIJSONValue(request.ToolChoice) {
		writeOpenAICompatError(
			w,
			http.StatusBadRequest,
			"tools and function calling are not supported yet",
			"invalid_request_error",
			"tools_not_supported",
		)
		return
	}

	prompt, err := openAIMessagesPrompt(
		request.Messages,
	)
	if err != nil {
		writeOpenAICompatError(
			w,
			http.StatusBadRequest,
			err.Error(),
			"invalid_request_error",
			"invalid_messages",
		)
		return
	}

	classification := strings.TrimSpace(
		r.Header.Get(
			openAIClassificationHeader,
		),
	)

	if classification == "" {
		classification = "internal"
	}

	result, err := s.ai.Invoke(
		r.Context(),
		airouter.InvokeInput{
			CallerSubject: trustedCallerSubject(
				r,
				"openai-compatible-client",
			),
			CostCenter: strings.TrimSpace(
				r.Header.Get(
					openAICostCenterHeader,
				),
			),
			UseCase:            "openai-chat-completions",
			DataClassification: classification,
			RequestedModel:     openAIInternalModel,
			Prompt:             prompt,
			Metadata: map[string]any{
				"interface":      "openai-chat-completions",
				"external_model": request.Model,
				"message_count":  len(request.Messages),
			},
		},
	)
	if err != nil {
		s.handleAIError(w, err)
		return
	}

	effectiveDecision :=
		result.Governance.Policy.Decision

	if result.Budget != nil &&
		effectiveDecision == "allow" {
		effectiveDecision =
			result.Budget.Decision
	}

	switch effectiveDecision {
	case "review":
		writeOpenAICompatError(
			w,
			http.StatusConflict,
			"request requires governance review",
			"governance_error",
			"review_required",
		)
		return

	case "deny":
		writeOpenAICompatError(
			w,
			http.StatusForbidden,
			"request denied by governance",
			"governance_error",
			"request_denied",
		)
		return

	case "allow":
		// Continue.

	default:
		writeOpenAICompatError(
			w,
			http.StatusInternalServerError,
			"internal server error",
			"server_error",
			"unexpected_decision",
		)
		return
	}

	if !result.ProviderCalled ||
		result.Response == nil ||
		result.Usage == nil {
		writeOpenAICompatError(
			w,
			http.StatusBadGateway,
			"AI provider returned an incomplete response",
			"server_error",
			"incomplete_provider_response",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		openAIChatResponse{
			ID: "chatcmpl-" +
				result.Governance.RequestID,
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   openAIExternalModel,
			Choices: []openAIChatChoice{
				{
					Index: 0,
					Message: openAIChatMessage{
						Role:    "assistant",
						Content: result.Response.Content,
					},
					FinishReason: "stop",
				},
			},
			Usage: openAIUsage{
				PromptTokens:     result.Usage.InputTokens,
				CompletionTokens: result.Usage.OutputTokens,
				TotalTokens: result.Usage.InputTokens +
					result.Usage.OutputTokens,
			},
		},
	)
}

func openAIMessagesPrompt(
	messages []openAIChatMessage,
) (string, error) {
	if len(messages) == 0 {
		return "", errors.New(
			"messages must not be empty",
		)
	}

	var builder strings.Builder

	for index, message := range messages {
		role := strings.ToLower(
			strings.TrimSpace(message.Role),
		)

		switch role {
		case "system",
			"developer",
			"user",
			"assistant":
		default:
			return "", fmt.Errorf(
				"unsupported message role %q",
				message.Role,
			)
		}

		if message.Content == "" {
			return "", fmt.Errorf(
				"message %d content must not be empty",
				index,
			)
		}

		if index > 0 {
			builder.WriteString("\n\n")
		}

		builder.WriteString(
			strings.ToUpper(role),
		)
		builder.WriteString(":\n")
		builder.WriteString(
			message.Content,
		)
	}

	return builder.String(), nil
}

func hasOpenAIJSONValue(
	raw json.RawMessage,
) bool {
	value := strings.TrimSpace(
		string(raw),
	)

	return value != "" &&
		value != "null" &&
		value != "[]"
}

func writeOpenAICompatError(
	w http.ResponseWriter,
	status int,
	message string,
	errorType string,
	code string,
) {
	writeJSON(
		w,
		status,
		map[string]any{
			"error": map[string]any{
				"message": message,
				"type":    errorType,
				"code":    code,
			},
		},
	)
}
