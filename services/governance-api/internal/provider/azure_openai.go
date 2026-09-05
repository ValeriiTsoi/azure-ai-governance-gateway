package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

const (
	azureOpenAIScope = "https://cognitiveservices.azure.com/.default"

	// GPT-5-mini uses reasoning tokens as part of its output-token budget.
	// Keep enough room for reasoning plus a short visible answer.
	azureOpenAIMaxOutputTokens int64 = 1000
)

type AzureOpenAI struct {
	client openai.Client
}

func NewAzureOpenAIManagedIdentity(
	endpoint string,
	clientID string,
) (*AzureOpenAI, error) {
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return nil, errors.New(
			"Azure managed identity client ID is required",
		)
	}

	credential, err := azidentity.NewManagedIdentityCredential(
		&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ClientID(clientID),
		},
	)
	if err != nil {
		return nil, err
	}

	return NewAzureOpenAI(
		endpoint,
		credential,
	)
}

func NewAzureOpenAI(
	endpoint string,
	credential azcore.TokenCredential,
) (*AzureOpenAI, error) {
	endpoint = strings.TrimSpace(endpoint)

	if endpoint == "" {
		return nil, errors.New(
			"azure OpenAI endpoint is required",
		)
	}

	if credential == nil {
		return nil, errors.New(
			"azure OpenAI token credential is required",
		)
	}

	authenticate := option.WithMiddleware(
		func(
			req *http.Request,
			next option.MiddlewareNext,
		) (*http.Response, error) {
			accessToken, err := credential.GetToken(
				req.Context(),
				policy.TokenRequestOptions{
					Scopes: []string{
						azureOpenAIScope,
					},
				},
			)
			if err != nil {
				return nil, fmt.Errorf(
					"get Azure OpenAI access token: %w",
					err,
				)
			}

			authenticatedRequest :=
				req.Clone(req.Context())

			authenticatedRequest.Header =
				req.Header.Clone()

			authenticatedRequest.Header.Set(
				"Authorization",
				"Bearer "+accessToken.Token,
			)

			return next(authenticatedRequest)
		},
	)

	client := openai.NewClient(
		// Explicitly clear any accidental OPENAI_API_KEY
		// inherited from the runtime environment.
		option.WithAPIKey(""),

		option.WithBaseURL(
			strings.TrimRight(endpoint, "/")+
				"/openai/v1/",
		),

		authenticate,

		option.WithMaxRetries(2),
		option.WithRequestTimeout(60*time.Second),
	)

	return &AzureOpenAI{
		client: client,
	}, nil
}

func newAzureOpenAIWithClient(
	client openai.Client,
) *AzureOpenAI {
	return &AzureOpenAI{
		client: client,
	}
}

func (p *AzureOpenAI) Name() string {
	return "azure-openai"
}

func (p *AzureOpenAI) Invoke(
	ctx context.Context,
	request InvokeRequest,
) (InvokeResponse, error) {
	model := strings.TrimSpace(request.Model)
	prompt := strings.TrimSpace(request.Prompt)

	if model == "" {
		return InvokeResponse{}, errors.New(
			"azure OpenAI model is required",
		)
	}

	if prompt == "" {
		return InvokeResponse{}, errors.New(
			"azure OpenAI prompt is required",
		)
	}

	response, err := p.client.Responses.New(
		ctx,
		responses.ResponseNewParams{
			Model: model,
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(prompt),
			},
			MaxOutputTokens: openai.Int(
				azureOpenAIMaxOutputTokens,
			),
		},

		// Keep the provider call stateless. The gateway itself
		// persists only prompt hash/length, route and usage.
		option.WithJSONSet("store", false),
	)
	if err != nil {
		return InvokeResponse{}, err
	}

	content := strings.TrimSpace(
		response.OutputText(),
	)

	if content == "" {
		return InvokeResponse{}, errors.New(
			"azure OpenAI returned no text output",
		)
	}

	return InvokeResponse{
		Content: content,
		Model:   model,
		Usage: Usage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
		},
	}, nil
}
