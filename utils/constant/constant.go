package constant

import (
	"time"

	"github.com/abhissng/neuron/utils/types"
)

// These are generic constant for the application
const (
	ServiceContext = "ServiceContext"
	RequestID      = "request_id"
	CorrelationID  = "correlation_id"
	BusinessID     = "business_id"
	UserID         = "user_id"
	Logger         = "logger"
	TraceID        = "trace_id"
	MetaData       = "meta_data"

	// These are general constant for config file
	Service              = "Service"
	Roles                = "Roles"
	Token                = "Token"
	RefreshToken         = "RefreshToken"
	DefaultAppPort       = "DefaultAppPort"
	Environment          = "Environment"
	RunMode              = "RunMode"
	ResponseBodyPrint    = "ResponseBodyPrint"
	LogRotationEnabled   = "LogRotationEnabled"
	SupportEmail         = "SupportEmail"
	HealthyStatusMessage = "connection healthy"
	JWTSecret            = "JWTSecret"
	ProjectId            = "ProjectId"
	VaultPath            = "VaultPath"
	IssuerKey            = "IssuerKey"
)

// These are generic constant status and action
const (
	// These are status constants
	Pending   types.Status = "pending"
	Completed types.Status = "completed"
	Failed    types.Status = "failed"
	Success   types.Status = "success"

	// These are action constants
	Process  types.Action = "process"
	Execute  types.Action = "execute"
	Rollback types.Action = "rollback"
)

// These are generic typed constant for the application
const (
	IS_PROD types.StringConstant = "IS_PROD"
)

// GraceFul Shutdown Constants
const (
	ServerDefaultGracefulTime  time.Duration = 10 * time.Second
	ServiceDefaultGracefulTime time.Duration = 5 * time.Second
)
