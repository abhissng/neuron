package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/abhissng/neuron/adapters/gin/middleware"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/acknowledgment"
	"github.com/abhissng/neuron/utils/types"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ServiceMiddlewareHandler is a function that takes a *context.ServiceContext and returns a result.Result[bool]
type ServiceMiddlewareHandler func(*context.ServiceContext) result.Result[bool]

// WrapServiceMiddlewareHandler wraps a ServiceMiddlewareHandler and returns a gin.HandlerFunc
func WrapServiceMiddlewareHandler(handler ServiceMiddlewareHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var response result.Result[bool]

		// Fetch the ServiceContext from the gin.Context
		ctx, err := middleware.GetServiceContext(c)
		if err != nil {
			// Handle the error if ServiceContext is not found
			err := blame.ServiceContextFetchError(viper.GetString(constant.SupportEmail), err)
			res := err.FetchErrorResponse(blame.WithTranslation())
			c.AbortWithStatusJSON(500, acknowledgment.NewAPIResponse(false, "", res))
			c.Request.Body.Close() // #nosec G104
			return
		}
		defer func() {
			switch exception := recover(); exception {
			case nil:
				if !response.IsSuccess() {
					_, err := response.Value()
					httpStatus := helpers.FetchHTTPStatusCode(err.FetchResponseType())
					res := err.FetchErrorResponse(blame.WithTranslation())
					ctx.SlogError(constant.MiddlewareFailed, log.WithField("error-code", err.FetchErrCode()))
					c.AbortWithStatusJSON(httpStatus, acknowledgment.NewAPIResponse(false, types.CorrelationID(ctx.GetGinContextCorrelationID()), res))
					c.Request.Body.Close() // #nosec G104
				} else {
					c.Next()
				}
			default:
				handleException("Middleware", ctx, exception)
				c.Request.Body.Close() // #nosec G104
			}
		}()

		response = handler(ctx)
	}
}

// RequestHandler is a function that takes a *context.ServiceContext and returns a result.Result[T]
type RequestHandler[T any] func(*context.ServiceContext) result.Result[T]

// ExecuteControllerHandler executes a RequestHandler and returns a gin.HandlerFunc
func ExecuteControllerHandler[T any](handler RequestHandler[T]) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Fetch the ServiceContext from the gin.Context
		ctx, err := middleware.GetServiceContext(c)
		if err != nil {
			// Handle the error if ServiceContext is not found
			err := blame.ServiceContextFetchError(viper.GetString(constant.SupportEmail), err)
			res := err.FetchErrorResponse(blame.WithTranslation())
			c.AbortWithStatusJSON(500, acknowledgment.NewAPIResponse[any](false, "", res))
			c.Request.Body.Close() // #nosec G104
			return
		}
		var handlerResult result.Result[T]

		defer func() {
			if err := recover(); err != nil {
				handleException("Controller", ctx, err)
				return
			}
			processResult(handlerResult, ctx)
			_ = c.Request.Body.Close() // #nosec G104
		}()

		handlerResult = handler(ctx)
	}
}

// handleException handles exceptions and logs the error, stack trace, and blame
func handleException(handlerType string, ctx *context.ServiceContext, err any) {
	ctx.SlogError("Exception Occured at "+handlerType, log.WithField("error", err))
	stackTrace := debug.Stack()
	helpers.Println(constant.ERROR, "Stack Trace: ", string(stackTrace))
	serverBlame := blame.InternalServerError(fmt.Errorf("error %+v", err))
	ctx.SlogError("Server Blame ", log.WithField("message", serverBlame.FetchErrorResponse()))
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"Message": "An unexpected error occurred. Please contact the administrator.",
	})
}

// processResult processes the result and returns the response
func processResult[T any](res result.Result[T], ctx *context.ServiceContext) {
	if !res.IsSuccess() {
		redirectURL, Redirect := res.Redirect()
		if Redirect {
			ctx.SlogInfo(constant.HandlerRedirect, log.WithField(constant.RedirectToURL, redirectURL))
			ctx.Redirect(http.StatusFound, redirectURL)
			return
		}

		_, blameInfo := res.Value()
		status := helpers.FetchHTTPStatusCode(blameInfo.FetchResponseType())
		errorResponse := blameInfo.FetchErrorResponse(blame.WithTranslation())
		ctx.SlogError(constant.HandlerFailed, log.WithField("Error Message", errorResponse))
		ctx.JSON(status, acknowledgment.NewAPIResponse[any](false, types.CorrelationID(ctx.GetGinContextCorrelationID()), errorResponse))
		return
	}

	redirectURL, Redirect := res.Redirect()
	if Redirect {
		ctx.SlogInfo(constant.HandlerRedirect, log.WithField(constant.RedirectToURL, redirectURL))
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	}

	data, _ := res.Value()
	ctx.JSON(http.StatusOK, acknowledgment.NewAPIResponse[*T](true, ctx.GetGinContextCorrelationID(), data))
}
