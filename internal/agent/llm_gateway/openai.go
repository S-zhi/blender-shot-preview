package llmgateway

import (
	"fmt"
	"strings"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

// openAIAdapter owns only OpenAI credential conventions. Model calls are out
// of scope for this gateway and will be added by a later client layer.
type openAIAdapter struct{}

func (openAIAdapter) provider() Provider { return ProviderOpenAI }

func (openAIAdapter) normalizeBaseURL(baseURL string) (string, error) {
	return normalizeHTTPURL(baseURL, defaultOpenAIBaseURL)
}

func (openAIAdapter) validateAPIKey(apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("%w: OpenAI API key is required", ErrInvalidCommand)
	}
	return nil
}

func (openAIAdapter) usableKey(record CredentialRecord, apiKey string) (UsableKey, error) {
	return UsableKey{
		KeyID:    record.ID,
		Provider: record.Provider,
		BaseURL:  record.BaseURL,
		APIKey:   apiKey,
		Headers: map[string]string{
			"Authorization": "Bearer " + apiKey,
		},
	}, nil
}
