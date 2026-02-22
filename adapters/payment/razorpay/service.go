package razorpay

// Service defines the type-safe Razorpay payment operations.
type Service interface {
	CreatePlan(req *PlanRequest, extraHeaders map[string]string, extraQueryParams map[string]string) (*Plan, error)
	FetchPlan(planID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Plan, error)
	FetchAllPlans(extraHeaders map[string]string, extraQueryParams map[string]any) ([]*Plan, error)
	CreateSubscription(req *SubscriptionRequest, extraHeaders map[string]string) (*Subscription, error)
	FetchSubscription(subID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Subscription, error)
	CancelSubscription(subID string, extraHeaders map[string]string, extraQueryParams map[string]any) error
	RefundPayment(paymentID string, amount int64, extraHeaders map[string]string, extraQueryParams map[string]any) (*Refund, error)
	CreateInvoice(req *InvoiceRequest, extraHeaders map[string]string) (*Invoice, error)
	FetchInvoice(invoiceID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Invoice, error)
	CancelInvoice(invoiceID string, extraHeaders map[string]string, extraQueryParams map[string]any) error
	DeleteInvoice(invoiceID string, extraHeaders map[string]string, extraQueryParams map[string]any) error
	FetchPayment(paymentID string, extraHeaders map[string]string, extraQueryParams map[string]any) (*Payment, error)
	VerifyWebhookSignature(body []byte, signature string) error
}

func NewService(client *Client) Service {
	return client
}
