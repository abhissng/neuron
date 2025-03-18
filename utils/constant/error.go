package constant

import "github.com/abhissng/neuron/utils/types"

// These are ComponentErrorType constant
const (
	ErrService     types.ComponentErrorType = "service"
	ErrModel       types.ComponentErrorType = "model"
	ErrAdaptors    types.ComponentErrorType = "adaptors"
	ErrMiddlewares types.ComponentErrorType = "middlewares"
	ErrController  types.ComponentErrorType = "controller"
	ErrApplication types.ComponentErrorType = "application"
	ErrLibrary     types.ComponentErrorType = "library"
	ErrUtils       types.ComponentErrorType = "utils"
	ErrSqlc        types.ComponentErrorType = "sqlc"
	ErrEngine      types.ComponentErrorType = "engine"
)

// These are generic HTTP request error constant
const (
	BadRequest     types.ResponseErrorType = "BadRequest"
	Forbidden      types.ResponseErrorType = "Forbidden"
	NotFound       types.ResponseErrorType = "NotFound"
	AlreadyExists  types.ResponseErrorType = "AlreadyExists"
	InternalServer types.ResponseErrorType = "InternalServerError"
	Unauthorized   types.ResponseErrorType = "Unauthorized"
)
