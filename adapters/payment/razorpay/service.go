package razorpay

// Service defines the type-safe Razorpay payment operations.
type Service interface {
	CreateOrder(req *OrderRequest, extraHeaders map[string]string) (*Order, error)
	FetchOrder(orderID string, queryParams map[string]any, extraHeaders map[string]string) (*Order, error)
	CapturePayment(paymentID string, amount int64, currency string, extraHeaders map[string]string) (*Payment, error)
	FetchRefund(refundID string, queryParams map[string]any, extraHeaders map[string]string) (*Refund, error)
	VerifyPaymentSignature(orderID, paymentID, signature string) bool
	KeyID() string
	CreatePlan(req *PlanRequest, extraHeaders map[string]string) (*Plan, error)
	FetchPlan(planID string, queryParams map[string]any, extraHeaders map[string]string) (*Plan, error)
	FetchAllPlans(queryParams map[string]any, extraHeaders map[string]string) ([]*Plan, error)
	CreateSubscription(req *SubscriptionRequest, extraHeaders map[string]string) (*Subscription, error)
	CreateSubscriptionLink(req *SubscriptionRequest, extraHeaders map[string]string) (*Subscription, error)
	FetchSubscription(subID string, queryParams map[string]any, extraHeaders map[string]string) (*Subscription, error)
	FetchAllSubscriptions(queryParams map[string]any, extraHeaders map[string]string) ([]*Subscription, error)
	CancelSubscription(subID string, data map[string]any, extraHeaders map[string]string) error
	UpdateSubscription(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error)
	FetchPendingUpdate(subID string, queryParams map[string]any, extraHeaders map[string]string) (*Subscription, error)
	CancelUpdate(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error)
	PauseSubscription(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error)
	ResumeSubscription(subID string, data map[string]any, extraHeaders map[string]string) (*Subscription, error)
	FetchSubscriptionInvoices(subID string, queryParams map[string]any, extraHeaders map[string]string) ([]*Invoice, error)
	DeleteOfferFromSubscription(subID string, offerID string, queryParams map[string]any, extraHeaders map[string]string) (*Subscription, error)
	RefundPayment(paymentID string, amount int64, queryParams map[string]any, extraHeaders map[string]string) (*Refund, error)
	CreateInvoice(req *InvoiceRequest, extraHeaders map[string]string) (*Invoice, error)
	FetchInvoice(invoiceID string, queryParams map[string]any, extraHeaders map[string]string) (*Invoice, error)
	CancelInvoice(invoiceID string, queryParams map[string]any, extraHeaders map[string]string) error
	DeleteInvoice(invoiceID string, queryParams map[string]any, extraHeaders map[string]string) error
	FetchPayment(paymentID string, queryParams map[string]any, extraHeaders map[string]string) (*Payment, error)
	VerifyWebhookSignature(body []byte, signature string) error
}

func NewService(client *Client) Service {
	return client
}
