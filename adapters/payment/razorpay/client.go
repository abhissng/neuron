package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/razorpay/razorpay-go"
)

// Client implements Service using the official Razorpay SDK.
type Client struct {
	rz     *razorpay.Client
	log    *log.Log
	key    string
	secret string
}

// NewClient returns a new payment service client. Key and secret are used for Razorpay API auth.
// Options can override the default SDK client (e.g. WithRazorpayClient).
func NewClient(key, secret string, log *log.Log, opts ...Option) *Client {
	c := &Client{rz: razorpay.NewClient(key, secret), key: key, secret: secret, log: log}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewClientWithRazorpay builds a client with an existing Razorpay SDK client (e.g. for tests or custom config).
func NewClientWithRazorpay(rz *razorpay.Client, opts ...Option) *Client {
	c := &Client{rz: rz}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Razorpay returns the underlying Razorpay client for advanced use.
func (c *Client) Razorpay() *razorpay.Client {
	return c.rz
}

// CreatePlan creates a plan.
func (c *Client) CreatePlan(req *PlanRequest, extraHeaders map[string]string, extraQueryParams map[string]string) (*Plan, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: create plan", log.Any("request", req))
	data, err := helpers.StructToMap(req)
	if err != nil {
		c.log.Error("payment: plan request", log.Any("error", err))
		return nil, fmt.Errorf("payment: plan request: %w", err)
	}

	res, err := c.rz.Plan.Create(data, extraHeaders)
	if err != nil {
		c.log.Error("payment: create plan", log.Any("error", err))
		return nil, fmt.Errorf("payment: create plan: %w", err)
	}
	c.log.Debug("payment: create plan response", log.Any("response", res))

	out, err := helpers.MapToStruct[*Plan](res)
	if err != nil {
		c.log.Error("payment: parse plan", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse plan: %w", err)
	}
	return out, nil
}

// FetchPlan fetches a plan by ID.
func (c *Client) FetchPlan(planID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Plan, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch plan", log.String("plan_id", planID))
	res, err := c.rz.Plan.Fetch(planID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch plan", log.Any("error", err))
		return nil, fmt.Errorf("payment: get plan: %w", err)
	}
	c.log.Debug("payment: get plan response", log.Any("response", res))

	out, err := helpers.MapTo[*Plan](res)
	if err != nil {
		c.log.Error("payment: parse plan", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse plan: %w", err)
	}
	return out, nil
}

// FetchAllPlans fetches all plans.
func (c *Client) FetchAllPlans(extraHeaders map[string]string, extraQueryParams map[string]any) ([]*Plan, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch all plans")
	res, err := c.rz.Plan.All(extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch all plans", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch all plans: %w", err)
	}
	c.log.Debug("payment: fetch all plans response", log.Any("response", res))
	items, ok := res["items"].([]interface{})
	if !ok {
		c.log.Error("payment: plans response has no items slice")
		return nil, fmt.Errorf("payment: parse all plans: response has no items slice")
	}
	planMaps := make([]map[string]any, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			c.log.Error("payment: plan item is not a map")
			return nil, fmt.Errorf("payment: parse all plans: invalid plan item")
		}
		planMaps = append(planMaps, m)
	}

	out, err := helpers.MapTo[[]*Plan](planMaps)
	if err != nil {
		c.log.Error("payment: parse all plans", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse all plans: %w", err)
	}
	return out, nil
}

// CreateSubscription creates a subscription.
func (c *Client) CreateSubscription(req *SubscriptionRequest, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: create subscription", log.Any("request", req))
	data, err := helpers.StructToMap(req)
	if err != nil {
		c.log.Error("payment: subscription request", log.Any("error", err))
		return nil, fmt.Errorf("payment: subscription request: %w", err)
	}
	c.log.Debug("payment: create subscription data", log.Any("data", data))
	res, err := c.rz.Subscription.Create(data, extraHeaders)
	if err != nil {
		c.log.Error("payment: create subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: create subscription: %w", err)
	}
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// GetSubscription fetches a subscription by ID.
func (c *Client) FetchSubscription(subID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch subscription", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.Fetch(subID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: get subscription: %w", err)
	}
	c.log.Debug("payment: get subscription response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// CancelSubscription cancels a subscription.
func (c *Client) CancelSubscription(subID string, extraHeaders map[string]string, extraQueryParams map[string]any) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: cancel subscription", log.String("subscription_id", subID))
	_, err := c.rz.Subscription.Cancel(subID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: cancel subscription", log.Any("error", err))
		return fmt.Errorf("payment: cancel subscription: %w", err)
	}
	return nil
}

// RefundPayment creates a refund for the given payment (amount in smallest currency unit, e.g. paisa).
func (c *Client) RefundPayment(paymentID string, amount int64, extraHeaders map[string]string, extraQueryParams map[string]any) (*Refund, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: refund payment", log.String("payment_id", paymentID), log.Int64("amount", amount))
	// SDK expects amount as int and merges it into data
	res, err := c.rz.Payment.Refund(paymentID, int(amount), extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: refund payment", log.Any("error", err))
		return nil, fmt.Errorf("payment: refund: %w", err)
	}
	c.log.Debug("payment: refund payment response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Refund](res)
	if err != nil {
		c.log.Error("payment: parse refund", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse refund: %w", err)
	}
	return out, nil
}

// CreateInvoice creates an invoice.
func (c *Client) CreateInvoice(req *InvoiceRequest, extraHeaders map[string]string) (*Invoice, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: create invoice", log.Any("request", req))
	data, err := helpers.StructToMap(req)
	if err != nil {
		c.log.Error("payment: invoice request", log.Any("error", err))
		return nil, fmt.Errorf("payment: invoice request: %w", err)
	}
	res, err := c.rz.Invoice.Create(data, nil)
	if err != nil {
		c.log.Error("payment: create invoice", log.Any("error", err))
		return nil, fmt.Errorf("payment: create invoice: %w", err)
	}
	c.log.Debug("payment: create invoice response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Invoice](res)
	if err != nil {
		c.log.Error("payment: parse invoice", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse invoice: %w", err)
	}
	return out, nil
}

// FetchInvoice fetches an invoice by ID.
func (c *Client) FetchInvoice(invoiceID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Invoice, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch invoice", log.String("invoice_id", invoiceID))
	res, err := c.rz.Invoice.Fetch(invoiceID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch invoice", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch invoice: %w", err)
	}
	c.log.Debug("payment: fetch invoice response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Invoice](res)
	if err != nil {
		c.log.Error("payment: parse invoice", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse invoice: %w", err)
	}
	return out, nil
}

// CancelInvoice cancels an invoice.
func (c *Client) CancelInvoice(invoiceID string, extraHeaders map[string]string, extraQueryParams map[string]any) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: cancel invoice", log.String("invoice_id", invoiceID))
	_, err := c.rz.Invoice.Cancel(invoiceID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: cancel invoice", log.Any("error", err))
		return fmt.Errorf("payment: cancel invoice: %w", err)
	}
	return nil
}

// DeleteInvoice deletes a draft invoice.
func (c *Client) DeleteInvoice(invoiceID string, extraHeaders map[string]string, extraQueryParams map[string]any) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: delete invoice", log.String("invoice_id", invoiceID))
	_, err := c.rz.Invoice.Delete(invoiceID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: delete invoice", log.Any("error", err))
		return fmt.Errorf("payment: delete invoice: %w", err)
	}
	return nil
}

// FetchPayment fetches a payment by ID.
func (c *Client) FetchPayment(paymentID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Payment, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch payment", log.String("payment_id", paymentID))
	res, err := c.rz.Payment.Fetch(paymentID, extraQueryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch payment", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch payment: %w", err)
	}
	c.log.Debug("payment: fetch payment response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Payment](res)
	if err != nil {
		c.log.Error("payment: parse payment", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse payment: %w", err)
	}
	return out, nil
}

// VerifyWebhookSignature verifies the X-Razorpay-Signature header using HMAC-SHA256.
// body must be the raw webhook request body; secret is the webhook secret.
func (c *Client) VerifyWebhookSignature(body []byte, signature string) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: verify webhook signature", log.String("signature", signature))
	err := VerifyWebhookSignature(body, signature, c.secret)
	if err != nil {
		c.log.Error("payment: verify webhook signature", log.Any("error", err))
		return fmt.Errorf("payment: verify webhook signature: %w", err)
	}
	return nil
}

// VerifyWebhookSignature is a package-level helper to verify webhook signatures without a Client.
// Use the raw request body and the X-Razorpay-Signature header value.
func VerifyWebhookSignature(body []byte, signature string, secret string) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	if len(body) == 0 || signature == "" || secret == "" {
		return fmt.Errorf("payment: verify webhook signature: body, signature or secret is empty")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("payment: verify webhook signature: signature does not match")
	}
	return nil
}

// ParseWebhookBody parses the raw webhook body into a WebhookEvent.
// Call VerifyWebhookSignature before trusting the payload.
func ParseWebhookBody(body []byte) (*WebhookEvent, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	var e WebhookEvent
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, fmt.Errorf("payment: parse webhook body: %w", err)
	}
	return &e, nil
}
