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

// KeyID returns the Razorpay API key ID (for client-side checkout).
func (c *Client) KeyID() string {
	return c.key
}

// CreateOrder creates an order.
func (c *Client) CreateOrder(req *OrderRequest, extraHeaders map[string]string) (*Order, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: create order", log.Any("request", req))
	data, err := helpers.StructToMap(req)
	if err != nil {
		c.log.Error("payment: order request", log.Any("error", err))
		return nil, fmt.Errorf("payment: order request: %w", err)
	}
	res, err := c.rz.Order.Create(data, extraHeaders)
	if err != nil {
		c.log.Error("payment: create order", log.Any("error", err))
		return nil, fmt.Errorf("payment: create order: %w", err)
	}
	c.log.Debug("payment: create order response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Order](res)
	if err != nil {
		c.log.Error("payment: parse order", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse order: %w", err)
	}
	return out, nil
}

// FetchOrder fetches an order by ID.
func (c *Client) FetchOrder(orderID string, queryParams map[string]any, extraHeaders map[string]string) (*Order, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch order", log.String("order_id", orderID))
	res, err := c.rz.Order.Fetch(orderID, queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch order", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch order: %w", err)
	}
	c.log.Debug("payment: fetch order response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Order](res)
	if err != nil {
		c.log.Error("payment: parse order", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse order: %w", err)
	}
	return out, nil
}

// CapturePayment captures an authorized payment.
func (c *Client) CapturePayment(paymentID string, amount int64, currency string, extraHeaders map[string]string) (*Payment, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: capture payment", log.String("payment_id", paymentID), log.Int64("amount", amount))
	data := map[string]any{"amount": amount, "currency": currency}
	res, err := c.rz.Payment.Capture(paymentID, int(amount), data, extraHeaders)
	if err != nil {
		c.log.Error("payment: capture payment", log.Any("error", err))
		return nil, fmt.Errorf("payment: capture payment: %w", err)
	}
	c.log.Debug("payment: capture payment response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Payment](res)
	if err != nil {
		c.log.Error("payment: parse payment", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse payment: %w", err)
	}
	return out, nil
}

// FetchRefund fetches a refund by ID.
func (c *Client) FetchRefund(refundID string, queryParams map[string]any, extraHeaders map[string]string) (*Refund, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch refund", log.String("refund_id", refundID))
	res, err := c.rz.Refund.Fetch(refundID, queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch refund", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch refund: %w", err)
	}
	c.log.Debug("payment: fetch refund response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Refund](res)
	if err != nil {
		c.log.Error("payment: parse refund", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse refund: %w", err)
	}
	return out, nil
}

// VerifyPaymentSignature verifies the payment signature from Razorpay checkout (order_id|payment_id signed with key_secret).
func (c *Client) VerifyPaymentSignature(orderID, paymentID, signature string) bool {
	defer func() {
		helpers.RecoverException(recover())
	}()
	if orderID == "" || paymentID == "" || signature == "" || c.secret == "" {
		return false
	}
	payload := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// CreatePlan creates a plan.
func (c *Client) CreatePlan(req *PlanRequest, extraHeaders map[string]string) (*Plan, error) {
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
func (c *Client) FetchPlan(planID string, queryParams map[string]any, extraHeaders map[string]string) (*Plan, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch plan", log.String("plan_id", planID))
	res, err := c.rz.Plan.Fetch(planID, queryParams, extraHeaders)
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
func (c *Client) FetchAllPlans(queryParams map[string]any, extraHeaders map[string]string) ([]*Plan, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch all plans")
	res, err := c.rz.Plan.All(queryParams, extraHeaders)
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
	c.log.Debug("payment: create subscription response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// CreateSubscriptionLink creates a subscription link (same API as Create; omit customer_id for link flow). Returns subscription with short_url.
func (c *Client) CreateSubscriptionLink(req *SubscriptionRequest, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: create subscription link", log.Any("request", req))
	data, err := helpers.StructToMap(req)
	if err != nil {
		c.log.Error("payment: subscription link request", log.Any("error", err))
		return nil, fmt.Errorf("payment: subscription link request: %w", err)
	}
	res, err := c.rz.Subscription.Create(data, extraHeaders)
	if err != nil {
		c.log.Error("payment: create subscription link", log.Any("error", err))
		return nil, fmt.Errorf("payment: create subscription link: %w", err)
	}
	c.log.Debug("payment: create subscription link response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription link", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription link: %w", err)
	}
	return out, nil
}

// GetSubscription fetches a subscription by ID.
func (c *Client) FetchSubscription(subID string, queryParams map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch subscription", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.Fetch(subID, queryParams, extraHeaders)
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
func (c *Client) CancelSubscription(subID string, data map[string]any, extraHeaders map[string]string) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: cancel subscription", log.String("subscription_id", subID))
	_, err := c.rz.Subscription.Cancel(subID, data, extraHeaders)
	if err != nil {
		c.log.Error("payment: cancel subscription", log.Any("error", err))
		return fmt.Errorf("payment: cancel subscription: %w", err)
	}
	return nil
}

// FetchAllSubscriptions fetches all subscriptions (supports count, skip, plan_id, from, to via queryParams).
func (c *Client) FetchAllSubscriptions(queryParams map[string]any, extraHeaders map[string]string) ([]*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch all subscriptions")
	res, err := c.rz.Subscription.All(queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch all subscriptions", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch all subscriptions: %w", err)
	}
	c.log.Debug("payment: fetch all subscriptions response", log.Any("response", res))
	items, ok := res["items"].([]interface{})
	if !ok {
		c.log.Error("payment: subscriptions response has no items slice")
		return nil, fmt.Errorf("payment: parse all subscriptions: response has no items slice")
	}
	subMaps := make([]map[string]any, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			c.log.Error("payment: subscription item is not a map")
			return nil, fmt.Errorf("payment: parse all subscriptions: invalid subscription item")
		}
		subMaps = append(subMaps, m)
	}
	out, err := helpers.MapTo[[]*Subscription](subMaps)
	if err != nil {
		c.log.Error("payment: parse all subscriptions", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse all subscriptions: %w", err)
	}
	return out, nil
}

// UpdateSubscription updates a subscription (e.g. quantity, schedule_change_at, plan_id via data).
func (c *Client) UpdateSubscription(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: update subscription", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.Update(subID, data, extraHeaders)
	if err != nil {
		c.log.Error("payment: update subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: update subscription: %w", err)
	}
	c.log.Debug("payment: update subscription response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// FetchPendingUpdate fetches details of a pending scheduled update for a subscription.
func (c *Client) FetchPendingUpdate(subID string, queryParams map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch pending update", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.PendingUpdate(subID, queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch pending update", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch pending update: %w", err)
	}
	c.log.Debug("payment: fetch pending update response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// CancelUpdate cancels a scheduled update for a subscription.
func (c *Client) CancelUpdate(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: cancel update", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.CancelScheduledChanges(subID, data, extraHeaders)
	if err != nil {
		c.log.Error("payment: cancel update", log.Any("error", err))
		return nil, fmt.Errorf("payment: cancel update: %w", err)
	}
	c.log.Debug("payment: cancel update response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// PauseSubscription pauses a subscription (e.g. data: map[string]any{"pause_at": "now"}).
func (c *Client) PauseSubscription(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: pause subscription", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.Pause(subID, data, extraHeaders)
	if err != nil {
		c.log.Error("payment: pause subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: pause subscription: %w", err)
	}
	c.log.Debug("payment: pause subscription response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// ResumeSubscription resumes a paused subscription (e.g. data: map[string]any{"resume_at": "now"}).
func (c *Client) ResumeSubscription(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: resume subscription", log.String("subscription_id", subID))
	res, err := c.rz.Subscription.Resume(subID, data, extraHeaders)
	if err != nil {
		c.log.Error("payment: resume subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: resume subscription: %w", err)
	}
	c.log.Debug("payment: resume subscription response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// FetchSubscriptionInvoices fetches all invoices for a subscription.
func (c *Client) FetchSubscriptionInvoices(subID string, queryParams map[string]any, extraHeaders map[string]string) ([]*Invoice, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch subscription invoices", log.String("subscription_id", subID))
	params := make(map[string]any)
	for k, v := range queryParams {
		params[k] = v
	}
	params["subscription_id"] = subID
	res, err := c.rz.Invoice.All(params, extraHeaders)
	if err != nil {
		c.log.Error("payment: fetch subscription invoices", log.Any("error", err))
		return nil, fmt.Errorf("payment: fetch subscription invoices: %w", err)
	}
	c.log.Debug("payment: fetch subscription invoices response", log.Any("response", res))
	items, ok := res["items"].([]interface{})
	if !ok {
		c.log.Error("payment: invoices response has no items slice")
		return nil, fmt.Errorf("payment: parse subscription invoices: response has no items slice")
	}
	invMaps := make([]map[string]any, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			c.log.Error("payment: invoice item is not a map")
			return nil, fmt.Errorf("payment: parse subscription invoices: invalid invoice item")
		}
		invMaps = append(invMaps, m)
	}
	out, err := helpers.MapTo[[]*Invoice](invMaps)
	if err != nil {
		c.log.Error("payment: parse subscription invoices", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription invoices: %w", err)
	}
	return out, nil
}

// DeleteOfferFromSubscription removes an offer linked to a subscription.
func (c *Client) DeleteOfferFromSubscription(subID string, offerID string, queryParams map[string]any, extraHeaders map[string]string) (*Subscription, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: delete offer from subscription", log.String("subscription_id", subID), log.String("offer_id", offerID))
	res, err := c.rz.Subscription.DeleteOffer(subID, offerID, queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: delete offer from subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: delete offer from subscription: %w", err)
	}
	c.log.Debug("payment: delete offer from subscription response", log.Any("response", res))
	out, err := helpers.MapToStruct[*Subscription](res)
	if err != nil {
		c.log.Error("payment: parse subscription", log.Any("error", err))
		return nil, fmt.Errorf("payment: parse subscription: %w", err)
	}
	return out, nil
}

// RefundPayment creates a refund for the given payment (amount in smallest currency unit, e.g. paisa).
func (c *Client) RefundPayment(paymentID string, amount int64, queryParams map[string]any, extraHeaders map[string]string) (*Refund, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: refund payment", log.String("payment_id", paymentID), log.Int64("amount", amount))
	// SDK expects amount as int and merges it into data
	res, err := c.rz.Payment.Refund(paymentID, int(amount), queryParams, extraHeaders)
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
func (c *Client) FetchInvoice(invoiceID string, queryParams map[string]any, extraHeaders map[string]string) (*Invoice, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch invoice", log.String("invoice_id", invoiceID))
	res, err := c.rz.Invoice.Fetch(invoiceID, queryParams, extraHeaders)
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
func (c *Client) CancelInvoice(invoiceID string, queryParams map[string]any, extraHeaders map[string]string) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: cancel invoice", log.String("invoice_id", invoiceID))
	_, err := c.rz.Invoice.Cancel(invoiceID, queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: cancel invoice", log.Any("error", err))
		return fmt.Errorf("payment: cancel invoice: %w", err)
	}
	return nil
}

// DeleteInvoice deletes a draft invoice.
func (c *Client) DeleteInvoice(invoiceID string, queryParams map[string]any, extraHeaders map[string]string) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: delete invoice", log.String("invoice_id", invoiceID))
	_, err := c.rz.Invoice.Delete(invoiceID, queryParams, extraHeaders)
	if err != nil {
		c.log.Error("payment: delete invoice", log.Any("error", err))
		return fmt.Errorf("payment: delete invoice: %w", err)
	}
	return nil
}

// FetchPayment fetches a payment by ID.
func (c *Client) FetchPayment(paymentID string, queryParams map[string]any, extraHeaders map[string]string) (*Payment, error) {
	defer func() {
		helpers.RecoverException(recover())
	}()
	c.log.Info("payment: fetch payment", log.String("payment_id", paymentID))
	res, err := c.rz.Payment.Fetch(paymentID, queryParams, extraHeaders)
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
