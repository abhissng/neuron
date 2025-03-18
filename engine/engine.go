package engine

import (
	"errors"
	"time"

	"github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/jwt"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/message"
	"github.com/abhissng/neuron/utils/structures/service"
	"github.com/abhissng/neuron/utils/types"
)

const (
	WaitTimeout = 5 * time.Second
	Payload     = "payload"
)

// ServiceResult structure
type ServiceResult struct {
	Response       map[string]any
	ExecutedStates []*service.ServiceState
}

// NewServiceResult creates a new ServiceResult
func NewServiceResult() *ServiceResult {
	return &ServiceResult{
		Response:       make(map[string]any),
		ExecutedStates: make([]*service.ServiceState, 0),
	}
}

// processServiceStates processes each state defined in the service definition.
// It forwards a request payload to each external service using NATS.
// Returns the response from each state and a list of executed states as a ServiceResult.
func ProcessServiceStates[T any](
	ctx *context.ServiceContext,
	serviceName types.Service,
	waitTimeout time.Duration,
	reqPayload *message.Message[T],
) result.Result[ServiceResult] {
	// Create a new ServiceResult.
	serviceResult := NewServiceResult()

	defer helpers.RecoverException(recover())

	if reqPayload == nil {
		return result.NewFailure[ServiceResult](blame.RequestPayloadNil(errors.New("unable to process service states")))
	}

	svcDef, err := ctx.GetServiceDefinition(serviceName.String())
	if err != nil {

		return result.NewFailure[ServiceResult](blame.ServiceDefinitionNotFound(serviceName.String(), err))
	}
	ctx.Log.Info("Service Definition Retrieved", ctx.Slog(log.Any("service", *svcDef))...)

	if !ctx.IsServiceActive(serviceName.String()) {
		return result.NewFailure[ServiceResult](blame.InactiveService(serviceName.String()))
	}

	var token string

	token, err = jwt.GenerateJWT(serviceName.String(), []string{"admin"}, FetchJWTSecret(ctx), 1*time.Minute)
	if err != nil {
		ctx.Log.Error("failed to generate jwt", log.Err(err))
		return result.NewFailure[ServiceResult](blame.CreateTokenFailed())
	}

	if waitTimeout == 0 {
		waitTimeout = WaitTimeout
	}

	// this Payload is a copy of the original request payload.
	// It is used to build the acknowledgment message.
	payloadToSend := reqPayload

	// Process states in order.
	for _, state := range svcDef.States.States {
		// Generate JWT token

		// For each state, merge the request payload with state-specific info.
		payloadToSend.CurrentService = state.Service

		// Make the external call using the NATS wrapper.
		msg, blameInfo := ctx.NATSManager.PublishAndWait(
			state.ExecuteSubject,
			svcDef.QueueGroup,
			payloadToSend,
			waitTimeout,
			nats.AddHeaderMiddleware(constant.CorrelationIDHeader, reqPayload.CorrelationID.String()),
			nats.AddHeaderMiddleware(constant.AuthorizationHeader, token),
			nats.AddHeaderMiddleware(constant.IPHeader, ctx.ClientIP()),
			nats.LogMiddleware(constant.Publish, ctx.Log))

		if blameInfo != nil {
			return result.NewFailureWithValue(serviceResult, blameInfo)
		}

		resp, err := codec.Decode[*message.Message[T]](msg.Data, codec.JSON)
		if err != nil {
			ctx.Log.Error("failed to unmarshal response", log.String("state", state.Service), log.Err(err))
			return result.NewFailureWithValue(serviceResult, blame.UnMarshalError(codec.JSON, err))
		}
		serviceResult.Response[Payload] = resp
		ctx.Log.Info("Response Retrieved", ctx.Slog(log.Any("response", resp))...)

		if !helpers.IsSuccess(resp.Status) {
			blameInfo := resp.Error.NewErrorResponseBlame(ctx.BlameWrapper)
			ctx.Log.Error("State Execution Failed", ctx.Slog(log.Any("state", state.Service), log.Any("error", blameInfo.ErrorFromBlame()))...)
			return result.NewFailureWithValue(serviceResult, blameInfo)
		}
		// Update initial payload
		payloadToSend = resp

		// Mark state as executed.
		serviceResult.ExecutedStates = append(serviceResult.ExecutedStates, state)

	}

	return result.NewSuccess(serviceResult)
}

// rollbackServiceStates handles the rollback procedure if some state fails.
// It supports two scenarios:
//  1. If svcDef.RollbackOrder is defined, it rolls back only those states (in that order)
//  2. Otherwise, it rolls back all executed states in reverse order.
//
// A state with an empty rollback subject is skipped.
func RollbackServiceStates[T any](ctx *context.ServiceContext, serviceName types.Service, serviceResult *ServiceResult) blame.Blame {
	defer helpers.RecoverException(recover())

	svcDef, err := ctx.GetServiceDefinition(serviceName.String())
	if err != nil {
		return blame.ServiceDefinitionNotFound(serviceName.String(), err)
	}

	if !ctx.IsServiceActive(serviceName.String()) {
		return blame.InactiveService(serviceName.String())
	}

	rollbackSequence := buildRollbackSequence(svcDef, serviceResult.ExecutedStates)

	// Execute rollback for each state
	for _, state := range rollbackSequence {
		executeStateRollback[T](ctx, state, serviceResult)
	}

	return nil
}

// buildRollbackSequence builds the sequence of states to rollback
func buildRollbackSequence(svcDef *service.ServiceDefinition, executedStates []*service.ServiceState) []*service.ServiceState {
	var rollbackSequence []*service.ServiceState

	if len(svcDef.States.RollbackOrder) > 0 {
		stateMap := make(map[string]*service.ServiceState, len(executedStates))
		for _, s := range executedStates {
			stateMap[s.Service] = s
		}
		for _, stateName := range svcDef.States.RollbackOrder {
			if s, exists := stateMap[stateName]; exists {
				rollbackSequence = append(rollbackSequence, s)
			}
		}
		return rollbackSequence
	}

	// Default: rollback in reverse order
	for i := len(executedStates) - 1; i >= 0; i-- {
		rollbackSequence = append(rollbackSequence, executedStates[i])
	}
	return rollbackSequence
}

// executeStateRollback executes rollback for a single state
func executeStateRollback[T any](ctx *context.ServiceContext, state *service.ServiceState, serviceResult *ServiceResult) {
	if state.RollbackSubject == "" {
		ctx.Log.Info("Skipping rollback for state (no rollback subject)", log.String("state", state.Service))
		return
	}

	respPayloadResult := GetPayload[T](serviceResult.Response)
	if !respPayloadResult.IsSuccess() {
		_, err := respPayloadResult.Value()
		ctx.Log.Error("failed to get response payload", log.String("state", state.Service), log.Err(err))
		return
	}

	respPayload, _ := respPayloadResult.Value()
	rollbackPayload := message.NewMessage(constant.Rollback, constant.Pending, respPayload.CorrelationID, respPayload.Payload)
	rollbackPayload.CurrentService = state.Service
	rollbackPayload.Action = constant.Rollback
	rollbackPayload.Status = constant.Pending

	reqBytes, err := codec.Encode(rollbackPayload, codec.JSON)
	if err != nil {
		ctx.Log.Error("failed to marshal rollback payload", log.String("state", state.Service), log.Err(err))
		return
	}

	_, blameInfo := ctx.NATSManager.PublishWithMiddleware(
		state.RollbackSubject,
		reqBytes,
		nats.AddHeaderMiddleware(constant.CorrelationIDHeader, respPayload.CorrelationID.String()),
		nats.LogMiddleware(constant.Publish, ctx.Log))
	if blameInfo != nil {
		ctx.Log.Error("rollback request failed", log.String("state", state.Service), log.Any("error", blameInfo.ErrorFromBlame()))
		return
	}
	ctx.Log.Info("Rollback successful for state", log.String("state", state.Service))
}
