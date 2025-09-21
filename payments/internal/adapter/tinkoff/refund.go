package tinkoffadp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shortlink-org/billing/payments/internal/application/payments/ports"
	"github.com/shortlink-org/billing/payments/internal/dto"
)

// RefundPayment creates a refund for a payment through Tinkoff API.
func (p *Provider) RefundPayment(ctx context.Context, in ports.RefundPaymentIn) (ports.RefundPaymentOut, error) {
	// Convert amount to minor units (kopecks for RUB)
	amount, err := dto.FormatAmountForTinkoff(in.Amount)
	if err != nil {
		return ports.RefundPaymentOut{}, fmt.Errorf("format amount: %w", err)
	}

	// Build Tinkoff refund request
	req := dto.TinkoffRefundRequest{
		PaymentID: in.ProviderID,
		Amount:    amount,
		Reason:    in.Reason,
		Meta:      make(map[string]interface{}),
	}

	// Add metadata
	for k, v := range in.Metadata {
		req.Meta[k] = v
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return ports.RefundPaymentOut{}, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/payment/refund", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return ports.RefundPaymentOut{}, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return ports.RefundPaymentOut{}, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ports.RefundPaymentOut{}, fmt.Errorf("read response: %w", err)
	}

	// Parse response
	var tinkoffResp dto.TinkoffRefundResponse
	if err := json.Unmarshal(respBody, &tinkoffResp); err != nil {
		return ports.RefundPaymentOut{}, fmt.Errorf("parse response: %w", err)
	}

	// Check for API errors
	if !tinkoffResp.Success {
		return ports.RefundPaymentOut{}, fmt.Errorf("tinkoff api error: %s - %s", tinkoffResp.Error.Code, tinkoffResp.Error.Message)
	}

	// Map response to our output format
	out := ports.RefundPaymentOut{
		Provider: ports.ProviderTinkoff,
		RefundID: tinkoffResp.Data.RefundID,
		Status:   dto.MapTinkoffRefundStatus(tinkoffResp.Data.Status),
		Amount:   dto.FromMinorTinkoff(in.Currency, tinkoffResp.Data.Amount),
	}

	return out, nil
}
