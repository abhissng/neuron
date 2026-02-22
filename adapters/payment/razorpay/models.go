package razorpay

import "errors"

// Subscription represents a Razorpay subscription entity.
type Subscription struct {
	ID                  string         `json:"id,omitempty"`
	Entity              string         `json:"entity,omitempty"`
	PlanID              string         `json:"plan_id,omitempty"`
	CustomerID          string         `json:"customer_id,omitempty"`
	Status              string         `json:"status,omitempty"`
	CurrentStart        int64          `json:"current_start,omitempty"`
	CurrentEnd          int64          `json:"current_end,omitempty"`
	EndedAt             *int64         `json:"ended_at,omitempty"`
	Quantity            int            `json:"quantity,omitempty"`
	ChargeAt            int64          `json:"charge_at,omitempty"`
	StartAt             int64          `json:"start_at,omitempty"`
	EndAt               int64          `json:"end_at,omitempty"`
	AuthAttempts        int            `json:"auth_attempts,omitempty"`
	TotalCount          int            `json:"total_count,omitempty"`
	PaidCount           int            `json:"paid_count,omitempty"`
	CustomerNotify      bool           `json:"customer_notify,omitempty"`
	CreatedAt           int64          `json:"created_at,omitempty"`
	ExpireBy            int64          `json:"expire_by,omitempty"`
	ShortURL            string         `json:"short_url,omitempty"`
	ScheduleChangeAt    string         `json:"schedule_change_at,omitempty"`
	HasScheduledChanges bool           `json:"has_scheduled_changes,omitempty"`
	ChangeScheduledAt   *int64         `json:"change_scheduled_at,omitempty"`
	RemainingCount      int            `json:"remaining_count,omitempty"`
	OfferID             string         `json:"offer_id,omitempty"`
	Source              string         `json:"source,omitempty"`
	Notes               map[string]any `json:"notes,omitempty"`
}

func NewSubscription() *Subscription {
	return &Subscription{
		Notes: make(map[string]any),
	}
}

func (s *Subscription) AddNote(key string, value any) {
	if s.Notes == nil {
		s.Notes = make(map[string]any)
	}
	if value == nil {
		delete(s.Notes, key)
	} else {
		s.Notes[key] = value
	}
}

// SubscriptionRequest is the payload for creating a subscription.
type SubscriptionRequest struct {
	PlanID         string         `json:"plan_id,omitempty"`
	TotalCount     int            `json:"total_count,omitempty"`
	Quantity       int            `json:"quantity,omitempty"`
	StartAt        int64          `json:"start_at,omitempty"`
	ExpireBy       int64          `json:"expire_by,omitempty"`
	CustomerNotify *bool          `json:"customer_notify,omitempty"`
	AddOns         []*PlanItem    `json:"add_ons,omitempty"`
	Notes          map[string]any `json:"notes,omitempty"`
	OfferID        string         `json:"offer_id,omitempty"`
}

func NewSubscriptionRequest() *SubscriptionRequest {
	return &SubscriptionRequest{
		AddOns: make([]*PlanItem, 0),
		Notes:  make(map[string]any),
	}
}

func (s *SubscriptionRequest) AddNote(key string, value any) {
	if s.Notes == nil {
		s.Notes = make(map[string]any)
	}
	if value == nil {
		delete(s.Notes, key)
	} else {
		s.Notes[key] = value
	}
}

func (s *SubscriptionRequest) AddAddOn(addOn *PlanItem) {
	s.AddOns = append(s.AddOns, addOn)
}

// Plan represents a Razorpay plan entity.
type Plan struct {
	ID        string         `json:"id,omitempty"`
	Entity    string         `json:"entity,omitempty"`
	Interval  int            `json:"interval,omitempty"`
	Period    string         `json:"period,omitempty"`
	Item      *PlanItem      `json:"item,omitempty"`
	CreatedAt int64          `json:"created_at,omitempty"`
	Notes     map[string]any `json:"notes,omitempty"`
}

func NewPlan() *Plan {
	return &Plan{
		Item:  NewPlanItem(),
		Notes: make(map[string]any),
	}
}

func (p *Plan) AddNote(key string, value any) {
	if p.Notes == nil {
		p.Notes = make(map[string]any)
	}
	if value == nil {
		delete(p.Notes, key)
	} else {
		p.Notes[key] = value
	}
}
func (p *Plan) Validate() error {
	if p.Period == "" {
		return errors.New("period is required")
	}

	if p.Interval == 0 {
		return errors.New("interval is required")
	}
	if p.Item != nil {
		if err := p.Item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// PlanItem holds plan item details.
type PlanItem struct {
	ID           string `json:"id,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	UnitAmount   int64  `json:"unit_amount,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Type         string `json:"type,omitempty"`
	Unit         any    `json:"unit,omitempty"`
	TaxInclusive bool   `json:"tax_inclusive,omitempty"`
	HsnCode      any    `json:"hsn_code,omitempty"`
	SacCode      any    `json:"sac_code,omitempty"`
	TaxRate      any    `json:"tax_rate,omitempty"`
	TaxID        any    `json:"tax_id,omitempty"`
	TaxGroupID   any    `json:"tax_group_id,omitempty"`
	CreatedAt    int64  `json:"created_at,omitempty"`
	UpdatedAt    int64  `json:"updated_at,omitempty"`
}

func NewPlanItem() *PlanItem {
	return &PlanItem{}
}

func (p *PlanItem) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.Amount == 0 {
		return errors.New("amount is required")
	}
	if p.Currency == "" {
		return errors.New("currency is required")
	}
	return nil
}

// PlanRequest is the payload for creating a plan.
// Use Notes to store your own identifiers (e.g. billing_plan_id / billing.plan UUID) for correlation with Razorpay plans.
type PlanRequest struct {
	Period   string         `json:"period"`
	Interval int            `json:"interval"`
	Item     PlanItem       `json:"item"`
	Notes    map[string]any `json:"notes,omitempty"`
}

func NewPlanRequest() *PlanRequest {
	return &PlanRequest{
		Notes: make(map[string]any),
	}
}

func (p *PlanRequest) AddNote(key string, value any) {
	if p.Notes == nil {
		p.Notes = make(map[string]any)
	}
	if value == nil {
		delete(p.Notes, key)
	} else {
		p.Notes[key] = value
	}
}

// Refund represents a Razorpay refund entity.
type Refund struct {
	ID        string         `json:"id"`
	Entity    string         `json:"entity"`
	PaymentID string         `json:"payment_id"`
	Amount    int64          `json:"amount"`
	Currency  string         `json:"currency"`
	Status    string         `json:"status"`
	Speed     string         `json:"speed"`
	CreatedAt int64          `json:"created_at"`
	Notes     map[string]any `json:"notes,omitempty"`
}

func NewRefund() *Refund {
	return &Refund{
		Notes: make(map[string]any),
	}
}

// Invoice represents a Razorpay invoice entity.
type Invoice struct {
	ID              string            `json:"id"`
	Entity          string            `json:"entity"`
	CustomerID      string            `json:"customer_id"`
	CustomerDetails *InvoiceCustomer  `json:"customer_details,omitempty"`
	OrderID         string            `json:"order_id,omitempty"`
	LineItems       []InvoiceLineItem `json:"line_items,omitempty"`
	Amount          int64             `json:"amount"`
	AmountPaid      int64             `json:"amount_paid"`
	AmountDue       int64             `json:"amount_due"`
	Currency        string            `json:"currency"`
	Status          string            `json:"status"`
	ShortURL        string            `json:"short_url"`
	Description     string            `json:"description,omitempty"`
	CreatedAt       int64             `json:"created_at"`
	Notes           map[string]any    `json:"notes,omitempty"`
}

func NewInvoice() *Invoice {
	return &Invoice{
		LineItems: []InvoiceLineItem{},
		Notes:     make(map[string]any),
	}
}

// InvoiceCustomer holds invoice customer details.
type InvoiceCustomer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Contact     string `json:"contact,omitempty"`
	GSTIN       string `json:"gstin,omitempty"`
	BillingAddr string `json:"billing_address,omitempty"`
}

func NewInvoiceCustomer() *InvoiceCustomer {
	return &InvoiceCustomer{}
}

// InvoiceLineItem represents a line item on an invoice.
type InvoiceLineItem struct {
	ID          string `json:"id"`
	ItemID      string `json:"item_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Quantity    int    `json:"quantity"`
	TaxAmount   int64  `json:"tax_amount,omitempty"`
}

func NewInvoiceLineItem() *InvoiceLineItem {
	return &InvoiceLineItem{}
}

// InvoiceRequest is the payload for creating an invoice.
type InvoiceRequest struct {
	Type            string            `json:"type"`
	CustomerID      string            `json:"customer_id,omitempty"`
	CustomerDetails *InvoiceCustomer  `json:"customer_details,omitempty"`
	LineItems       []InvoiceLineItem `json:"line_items,omitempty"`
	Description     string            `json:"description,omitempty"`
	Currency        string            `json:"currency,omitempty"`
	Notes           map[string]any    `json:"notes,omitempty"`
}

func NewInvoiceRequest() *InvoiceRequest {
	return &InvoiceRequest{
		LineItems: []InvoiceLineItem{},
		Notes:     make(map[string]any),
	}
}

// Payment represents a Razorpay payment entity.
type Payment struct {
	ID        string         `json:"id"`
	Entity    string         `json:"entity"`
	OrderID   string         `json:"order_id"`
	Method    string         `json:"method"`
	Amount    int64          `json:"amount"`
	Currency  string         `json:"currency"`
	Status    string         `json:"status"`
	Captured  bool           `json:"captured"`
	Email     string         `json:"email,omitempty"`
	Contact   string         `json:"contact,omitempty"`
	CreatedAt int64          `json:"created_at"`
	Notes     map[string]any `json:"notes,omitempty"`
}

func NewPayment() *Payment {
	return &Payment{
		Notes: make(map[string]any),
	}
}

// WebhookEvent is the standard Razorpay webhook payload.
type WebhookEvent struct {
	Entity    string         `json:"entity"`
	AccountID string         `json:"account_id"`
	Event     string         `json:"event"`
	Contains  []string       `json:"contains"`
	Payload   map[string]any `json:"payload"`
	CreatedAt int64          `json:"created_at"`
}

func NewWebhookEvent() *WebhookEvent {
	return &WebhookEvent{
		Contains: []string{},
		Payload:  make(map[string]any),
	}
}
