# Payment and Email Adapters Usage

Packages covered:

- [`adapters/payment`](../../adapters/payment)
- [`adapters/payment/razorpay`](../../adapters/payment/razorpay)
- [`adapters/email`](../../adapters/email)

## Payment Manager Pattern

`adapters/payment` provides provider registration and service lookup:

- register provider implementation once at startup
- fetch concrete service by provider enum/name

This decouples calling code from specific gateway implementations.

```go
pm := payment.NewManager()

client := razorpay.NewClient(key, secret, logger)
svc := razorpay.NewService(client)
pm.Register(payment.ProviderRazorpay, svc)

rz, ok := payment.GetService[razorpay.Service](pm, payment.ProviderRazorpay)
_ = rz
_ = ok
```

## Razorpay Adapter

`adapters/payment/razorpay` includes:

- typed request/response models
- client methods for orders/subscriptions/invoices/payments
- webhook signature verification

Recommended usage:

```go
client := razorpay.NewClient(key, secret, logger)
svc := razorpay.NewService(client)
```

Then register service in payment manager and expose through app context.

## Email Adapter

`adapters/email` contains:

- email client setup
- message/attachment modeling
- provider/server configurations

Use it through app-level dependency injection rather than constructing ad-hoc clients in handlers.

```go
emailClient, err := email.NewGomailClient(
	email.WithHost("smtp.example.com"),
	email.WithPort(587),
	email.WithLog(logger),
)
if err != nil {
	return err
}

err = emailClient.Send(&email.EmailData{
	To:      []string{"user@example.com"},
	Subject: "Welcome",
	TextBody: "Hello from neuron",
})
_ = err
```

## Internal Helpers (implementation detail)

Model conversion and request helper methods in these packages can evolve independently of exported interfaces.

## Caveats

- Keep signature verification strict for webhook paths.
- Normalize monetary units consistently (smallest currency unit where required).
- Protect provider secrets and avoid logging sensitive payloads.
