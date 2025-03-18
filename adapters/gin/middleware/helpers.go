package middleware

import (
	"errors"

	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-gonic/gin"
)

// GetServiceContext retrieves the ServiceContext from the gin.Context.
// It returns an error if the ServiceContext is not found or is of the wrong type.
func GetServiceContext(c *gin.Context) (*context.ServiceContext, error) {
	sc, exists := c.Get(constant.ServiceContext)
	if !exists {
		return nil, errors.New("ServiceContext not found in gin.Context")
	}

	serviceCtx, ok := sc.(*context.ServiceContext)
	if !ok {
		return nil, errors.New("invalid type for ServiceContext in gin.Context")
	}

	return serviceCtx, nil
}
