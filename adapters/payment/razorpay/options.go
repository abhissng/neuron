package razorpay

import (
	"github.com/razorpay/razorpay-go"
)

// Option configures the payment Client.
type Option func(*Client)

// WithRazorpayClient injects the Razorpay SDK client. Use for tests or custom SDK config (timeout, base URL, etc.).
// If not set, NewClient creates one from key and secret.
func WithRazorpayClient(rz *razorpay.Client) Option {
	return func(c *Client) {
		c.rz = rz
	}
}
