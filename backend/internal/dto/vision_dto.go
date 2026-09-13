package dto

// GeminiInlineData defines the inline image payload for Gemini API.
type GeminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GeminiPart defines a single part (text or media) in a Gemini content request.
type GeminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inline_data,omitempty"`
}

// GeminiContent defines the content object in Gemini request/response.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPropertySchema defines schema properties for Gemini JSON generation config.
type GeminiPropertySchema struct {
	Type        string                          `json:"type"`
	Description string                          `json:"description,omitempty"`
	Items       *GeminiPropertySchema           `json:"items,omitempty"`
	Properties  map[string]GeminiPropertySchema `json:"properties,omitempty"`
	Required    []string                        `json:"required,omitempty"`
}

// GeminiGenerationConfig defines generation config including structured JSON output schema.
type GeminiGenerationConfig struct {
	ResponseMimeType string               `json:"response_mime_type"`
	ResponseSchema   GeminiPropertySchema `json:"response_schema"`
}

// GeminiGenerateContentRequest defines the full request payload sent to Gemini API.
type GeminiGenerateContentRequest struct {
	Contents         []GeminiContent        `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig"`
}

// GeminiCandidate defines candidate in Gemini API response.
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// GeminiGenerateContentResponse defines response body returned from Gemini API.
type GeminiGenerateContentResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// GeminiIntermediateItem defines individual item extracted in Gemini OCR raw response.
type GeminiIntermediateItem struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
}

// GeminiIntermediateReceipt defines raw JSON structure extracted from receipt images.
type GeminiIntermediateReceipt struct {
	MerchantName string                   `json:"merchant_name"`
	Date         string                   `json:"date"`
	Items        []GeminiIntermediateItem `json:"items"`
	TotalAmount  float64                  `json:"total_amount"`
}
