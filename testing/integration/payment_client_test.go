package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/example/order-processing-service/testing/setup"
)

// PaymentClient calls an external payment HTTP API (stubbed by WireMock).
type PaymentClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ChargeResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ChargePOST POSTs /v1/charges and decodes JSON.
func (c *PaymentClient) ChargePOST(ctx context.Context, body map[string]any) (*ChargeResponse, error) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/v1/charges", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 500 {
		return nil, errors.New("payment service error")
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New("payment rejected")
	}
	var out ChargeResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func registerStub(t *testing.T, baseURL, mapping string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/__admin/mappings", strings.NewReader(mapping))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("wiremock admin: %d %s", resp.StatusCode, string(b))
	}
}

func TestPaymentClient_ExternalAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	wc, err := setup.StartWireMock(ctx)
	if err != nil {
		t.Skipf("integration: wiremock container (need Docker): %v", err)
	}
	defer func() { _ = wc.Container.Terminate(ctx) }()

	// Success
	okBody := `{"id":"ch_1","status":"succeeded"}`
	mappingOK := `{
  "request": { "method": "POST", "urlPath": "/v1/charges" },
  "response": {
    "status": 200,
    "headers": { "Content-Type": "application/json" },
    "jsonBody": ` + okBody + `
  }
}`
	registerStub(t, wc.BaseURL, mappingOK)

	client := &PaymentClient{
		BaseURL: wc.BaseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	out, err := client.ChargePOST(ctx, map[string]any{"amount": 100})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "ch_1" || out.Status != "succeeded" {
		t.Fatalf("got %+v", out)
	}

	// Failure: 500 from payment — reset stubs so only the error mapping applies
	del, _ := http.NewRequest(http.MethodDelete, wc.BaseURL+"/__admin/mappings", nil)
	respDel, _ := http.DefaultClient.Do(del)
	if respDel != nil {
		respDel.Body.Close()
	}

	mapping500 := `{
  "request": { "method": "POST", "urlPath": "/v1/charges" },
  "response": { "status": 500, "body": "internal error" }
}`
	registerStub(t, wc.BaseURL, mapping500)
	_, err = client.ChargePOST(ctx, map[string]any{"amount": 200})
	if err == nil || err.Error() != "payment service error" {
		t.Fatalf("want payment service error, got %v", err)
	}

	// Timeout: unroutable host + short client timeout (no WireMock involved)
	del2, _ := http.NewRequest(http.MethodDelete, wc.BaseURL+"/__admin/mappings", nil)
	resp2, _ := http.DefaultClient.Do(del2)
	if resp2 != nil {
		resp2.Body.Close()
	}
	clientSlow := &PaymentClient{
		BaseURL: "http://192.0.2.1:9", // TEST-NET-1 — should not respond
		HTTPClient: &http.Client{
			Timeout: 50 * time.Millisecond,
		},
	}
	_, err = clientSlow.ChargePOST(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected timeout / connection error")
	}
}
