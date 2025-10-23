package context

import (
	"os"

	"github.com/abhissng/neuron/adapters/aws"
	"github.com/abhissng/neuron/adapters/email"
	"github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/http"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/mongo"
	"github.com/abhissng/neuron/adapters/oci"
	"github.com/abhissng/neuron/adapters/paseto"
	"github.com/abhissng/neuron/adapters/redis"
	"github.com/abhissng/neuron/adapters/session"
	"github.com/abhissng/neuron/adapters/vault"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/database"
	"github.com/abhissng/neuron/utils/cache"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/cryptography"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/service"
	_ "github.com/abhissng/neuron/utils/timeutil"
)

// AppContext holds application-specific context data.
type AppContext struct {
	*blame.BlameManager
	*paseto.PasetoManager
	*log.Log
	*nats.NATSManager
	*service.Services
	*http.HttpClientManager
	*vault.Vault
	*redis.RedisManager
	*aws.AWSManager
	database.Database
	cache.Cache[string, any]
	*cryptography.CryptoManager
	email.EmailClient
	*oci.OCIManager
	*mongo.MongoManager
	*session.SessionManager

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

// WithBlameManager sets the blame manager for the AppContext.
func WithBlameManager(bw *blame.BlameManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.BlameManager = bw
	}
}

// WithInitBlameManager sets the blame manager for the AppContext.
func WithInitBlameManager(opts *blame.BlameManagerOption) AppContextOption {
	BlameManager, err := blame.NewBlameManager(opts)
	if err != nil {
		helpers.Println(constant.ERROR, "Error initialising blame manager : ", err)
		os.Exit(1)
	}
	return func(ctx *AppContext) {
		ctx.BlameManager = BlameManager
	}
}

// GetBlameManager retrieves the BlameManager from the App context.
func (ctx *AppContext) GetBlameManager() *blame.BlameManager {
	return ctx.BlameManager
}

// WithLogger sets the logger wrapper for the AppContext.
func WithLogger(logger *log.Log) AppContextOption {
	return func(ctx *AppContext) {
		ctx.Log = logger
	}
}

// WithRedisManager sets the redis manager for the AppContext.
func WithRedisManager(manager *redis.RedisManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.RedisManager = manager
	}
}

// WithAWSManager sets the aws manager for the AppContext.
func WithAWSManager(manager *aws.AWSManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.AWSManager = manager
	}
}

// WithDatabase sets the database for the AppContext.
func WithDatabase(database database.Database) AppContextOption {
	return func(ctx *AppContext) {
		ctx.Database = database
	}
}

// WithPasetoManager sets the paseto manager for the AppContext.
func WithPasetoManager(opts ...paseto.PasetoOption) AppContextOption {
	return func(ctx *AppContext) {
		ctx.PasetoManager = paseto.NewPasetoManager(opts...)
	}

}

// WithNATSManager sets the nats manager for the AppContext.
func WithNATSManager(url string, options ...nats.Option) AppContextOption {
	nats, err := nats.NewNATSManager(url, options...)
	if err != nil {
		helpers.Println(constant.ERROR, "Error initialising nats manager : ", err)
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

// WithHttpClientManager sets the http client manager for the AppContext.
func WithHttpClientManager(url string, opts ...http.RequestOption) AppContextOption {
	return func(ctx *AppContext) {
		ctx.HttpClientManager = http.NewHttpClientManager(url, opts...)
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

// WithCryptoManager sets the crypto manager for the AppContext.
func WithCryptoManager(manager *cryptography.CryptoManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.CryptoManager = manager
	}
}

// WithEmailClient sets the email client for the AppContext.
func WithEmailClient(client email.EmailClient) AppContextOption {
	return func(ctx *AppContext) {
		ctx.EmailClient = client
	}
}

// WithOciManager sets the oci manager for the AppContext.
func WithOciManager(manager *oci.OCIManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.OCIManager = manager
	}
}

// WithMongoManager sets the mongo manager for the AppContext.
func WithMongoManager(manager *mongo.MongoManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.MongoManager = manager
	}
}

// WithSessionManager sets the session manager for the AppContext.
func WithSessionManager(manager *session.SessionManager) AppContextOption {
	return func(ctx *AppContext) {
		ctx.SessionManager = manager
	}
}
