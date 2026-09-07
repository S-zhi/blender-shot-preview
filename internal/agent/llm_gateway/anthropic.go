package llmgateway

import (
	"fmt"
	"strings"
)

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com"
	anthropicAPIVersion     = "2023-06-01"
)

// anthropicAdapter owns Anthropic credential conventions. Model calls are out
// of scope for this gateway and will be added by a later client layer.
type anthropicAdapter struct{}

func (anthropicAdapter) provider() Provider { return ProviderAnthropic }

func (anthropicAdapter) normalizeBaseURL(baseURL string) (string, error) {
	return normalizeHTTPURL(baseURL, defaultAnthropicBaseURL)
}

func (anthropicAdapter) validateAPIKey(apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("%w: Anthropic API key is required", ErrInvalidCommand)
	}
	return nil
}

func (anthropicAdapter) usableKey(record CredentialRecord, apiKey string) (UsableKey, error) {
	return UsableKey{
		KeyID:    record.ID,
		Provider: record.Provider,
		BaseURL:  record.BaseURL,
		APIKey:   apiKey,
		Headers: map[string]string{
			"x-api-key":         apiKey,
			"anthropic-version": anthropicAPIVersion,
		},
	}, nil
}
