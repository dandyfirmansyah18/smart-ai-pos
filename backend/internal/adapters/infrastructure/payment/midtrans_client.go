package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/dto"
	"github.com/pos-backend/internal/ports/outbound"
)

const (
	MidtransSandboxURL    = "https://app.sandbox.midtrans.com/snap/v1/transactions"
	MidtransProductionURL = "https://app.midtrans.com/snap/v1/transactions"
)

type MidtransClient struct {
	serverKey    string
	clientKey    string
	isProduction bool
}

func NewMidtransClient(cfg *config.Config) *MidtransClient {
	return &MidtransClient{
		serverKey:    cfg.MidtransServerKey,
		clientKey:    cfg.MidtransClientKey,
		isProduction: cfg.MidtransIsProduction,
	}
}

func (m *MidtransClient) CreateSnapToken(ctx context.Context, orderID string, amount float64, customerName string) (string, string, error) {
	if m.serverKey == "" {
		mockToken := fmt.Sprintf("MOCK-SNAP-TOKEN-%s", orderID[:8])
		mockURL := fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", mockToken)
		return mockToken, mockURL, nil
	}

	baseURL := MidtransSandboxURL
	if m.isProduction {
		baseURL = MidtransProductionURL
	}

	custName := "Valued Customer"
	if customerName != "" {
		custName = customerName
	}

	reqBody := dto.MidtransSnapRequest{
		TransactionDetails: dto.MidtransTransactionDetails{
			OrderID:     orderID,
			GrossAmount: amount,
		},
		CustomerDetails: dto.MidtransCustomerDetails{
			FirstName: custName,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal midtrans request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", "", fmt.Errorf("failed to create midtrans request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(m.serverKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to call midtrans snap api: %w", err)
	}
	defer resp.Body.Close()

	var snapResp dto.MidtransSnapResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapResp); err != nil {
		return "", "", fmt.Errorf("failed to decode midtrans response: %w", err)
	}

	token := snapResp.Token
	redirectURL := snapResp.RedirectURL

	if token == "" {
		token = fmt.Sprintf("FALLBACK-SNAP-%s", orderID[:8])
	}
	if redirectURL == "" {
		redirectURL = "https://simulator.sandbox.midtrans.com/"
	}

	return token, redirectURL, nil
}

var _ outbound.PaymentGatewayClient = (*MidtransClient)(nil)
