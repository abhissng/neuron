package payment

import (
	"sync"

	"github.com/abhissng/neuron/adapters/payment/razorpay"
)

// Provider name constants for use with Register and GetService.
const (
	ProviderRazorpay = "razorpay"
	// Add more as needed: ProviderStripe = "stripe"
)

// Service is the interface that all payment provider services must implement.
// Extend this union type when adding new providers.
type Service interface {
	razorpay.Service // Add more as needed
}

// Manager holds multiple payment provider services and is intended to be attached to context.
// Register each provider's service (interface implementation) and retrieve them type-safely via GetService[T] or convenience methods.
type Manager struct {
	mu       sync.RWMutex
	services map[string]Service
}

// NewManager returns a new payment manager with no services registered.
func NewManager() *Manager {
	return &Manager{
		services: make(map[string]Service),
	}
}

// Register adds a service for the given provider name. Overwrites any existing service for that name.
// Pass the provider's service implementation (e.g. *razorpay.Client which implements razorpay.Service).
func (m *Manager) Register(provider string, service Service) {
	if service == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.services == nil {
		m.services = make(map[string]Service)
	}
	m.services[provider] = service
}

// GetService returns the service for the given provider as type T, or (zero, false) if not registered or type assertion fails.
// T must be a valid Service type (razorpay.Service, stripe.Service, etc.).
//
// Usage:
//
//	svc, ok := payment.GetService[razorpay.Service](m, payment.ProviderRazorpay)
//	stripeSvc, ok := payment.GetService[stripe.Service](m, payment.ProviderStripe)
func GetService[T Service](m *Manager, provider string) (T, bool) {
	if m == nil {
		var zero T
		return zero, false
	}
	m.mu.RLock()
	s, ok := m.services[provider]
	m.mu.RUnlock()
	if !ok || s == nil {
		var zero T
		return zero, false
	}
	t, ok := s.(T)
	return t, ok
}

// RazorpayService returns the Razorpay service if registered under ProviderRazorpay.
func (m *Manager) RazorpayService() (razorpay.Service, bool) {
	return GetService[razorpay.Service](m, ProviderRazorpay)
}

// // StripeService returns the Stripe service if registered under ProviderStripe.
// func (m *Manager) StripeService() (stripe.Service, bool) {
// 	return GetService[stripe.Service](m, ProviderStripe)
// }
