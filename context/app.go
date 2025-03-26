package context

import (
	"os"

	"github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/http"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/paseto"
	"github.com/abhissng/neuron/adapters/vault"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/database"
	"github.com/abhissng/neuron/utils/cache"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/service"
	_ "github.com/abhissng/neuron/utils/timeutil"
)

// AppContext holds application-specific context data.
type AppContext struct {
	*blame.BlameWrapper
	*paseto.PasetoWrapper
	*log.Log
	*nats.NATSManager
	*service.Services
	*http.HttpClientWrapper
	*vault.Vault
	database.Database
	cache.Cache[string, any]
	serviceId      string
	isDebugEnabled bool
	// Add other fields as needed (e.g., user ID, authentication information)
}

// AppContextOption is a function that modifies the AppContext.
type AppContextOption func(*AppContext)

// NewAppContext creates a new AppContext with the given options.
func NewAppContext(opts ...AppContextOption) *AppContext {
	appCtx := &AppContext{}
	for _, opt := range opts {
		opt(appCtx)
	}
	return appCtx
}

// WithServiceId sets the service ID for the AppContext.
func WithServiceID(serviceId string) AppContextOption {
	return func(ctx *AppContext) {
		ctx.serviceId = serviceId
	}
}

// WithDebugEnabled sets the DebugEnabled to true for the AppContext.
func WithDebugEnabled() AppContextOption {
	return func(ctx *AppContext) {
		ctx.isDebugEnabled = true
	}
}

// WithBlameWrapper sets the blame wrapper for the AppContext.
func WithBlameWrapper(localeDir string, languageTag string) AppContextOption {
	blameWrapper, err := blame.NewBlameWrapper(localeDir, languageTag)
	if err != nil {
		helpers.Println(constant.ERROR, "Error initialising blame wrapper : ", err)
		os.Exit(1)
	}

	return func(ctx *AppContext) {
		ctx.BlameWrapper = blameWrapper
	}

}

// GetBlameWrapper retrieves the BlameWrapper from the App context.
func (ctx *AppContext) GetBlameWrapper() *blame.BlameWrapper {
	return ctx.BlameWrapper
}

// WithLogger sets the logger wrapper for the AppContext.
func WithLogger(logger *log.Log) AppContextOption {
	return func(ctx *AppContext) {
		ctx.Log = logger
	}
}

// // WithContext sets the context for the AppContext.
// func WithContext(newCtx context.Context) AppContextOption {
// 	return func(ctx *AppContext) {
// 		ctx.Context = newCtx
// 	}
// }

// WithDatabase sets the database for the AppContext.
func WithDatabase(database database.Database) AppContextOption {
	return func(ctx *AppContext) {
		ctx.Database = database
	}
}

// WithPasetoWrapper sets the paseto wrapper for the AppContext.
func WithPasetoWrapper(opts ...paseto.PasetoOption) AppContextOption {
	return func(ctx *AppContext) {
		ctx.PasetoWrapper = paseto.NewPasetoWrapper(opts...)
	}

}

// WithNATSManager sets the nats wrapper for the AppContext.
func WithNATSManager(url string, options ...nats.Option) AppContextOption {
	nats, err := nats.NewNATSManager(url, options...)
	if err != nil {
		helpers.Println(constant.ERROR, "Error initialising nats wrapper : ", err)
		os.Exit(1)
	}

	return func(ctx *AppContext) {
		ctx.NATSManager = nats
	}
}

// AttachServices sets the services for the AppContext.
func (appCtx *AppContext) AttachServices(services *service.Services) {
	appCtx.Services = services
}

// WithHttpClientWrapper sets the http client wrapper for the AppContext.
func WithHttpClientWrapper(url string, opts ...http.RequestOption) AppContextOption {
	return func(ctx *AppContext) {
		ctx.HttpClientWrapper = http.NewHttpClientWrapper(url, opts...)
	}
}

// WithVault sets the vault for the AppContext.
func WithVault(vlt *vault.Vault) AppContextOption {
	return func(ctx *AppContext) {
		ctx.Vault = vlt
	}
}

// WithCacheManager sets the cache manager for the AppContext.
func WithCacheManager(config *cache.CacheConfig) AppContextOption {
	return func(ctx *AppContext) {
		if config != nil {
			ctx.Cache = cache.NewCacheManagerWithConfig(*config).CreateCache("default")
			return
		}
		ctx.Cache = cache.NewCacheManager().CreateCache("default")
	}
}
