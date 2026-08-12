package blame

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed error_definition.json
var embeddedBlameData []byte

// localBlameManager is a local instance of the BlameManager struct.
var (
	localBlameManager = &BlameManager{}
)

// getLocalBlameManager returns the localBlameManager instance.
func getLocalBlameManager() *BlameManager {
	return localBlameManager
}

// initLocalBlames initializes the local error blames from a JSON file.
func initLocalBlames() ([]BlameDefinition, error) {

	var localBlames []BlameDefinition
	if err := json.Unmarshal(embeddedBlameData, &localBlames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal local blame definition file: %w", err)
	}

	return localBlames, nil
}

// InitLocalBlameManager initializes the local blame wrapper with the given bundle.
func InitLocalBlameManager(bundle *i18n.Bundle) error {
	blameDefinitions, err := initLocalBlames()
	if err != nil {
		helpers.Println(constant.ERROR, "Error initialising local blame definitions: ", err)
		return err
	}

	// Create a map of error definitions
	blameDefinitionsMap := make(map[types.ErrorCode]Blame)
	for index, def := range blameDefinitions {
		if helpers.IsEmpty(def.ReasonCode) {
			def.ReasonCode = helpers.GenerateReasonCode(ReasonCodeNameSpace, ReasonCodeBase+index)
		}
		blameDefinitionsMap[types.ErrorCode(def.Code)] =
			NewBlame(def.ReasonCode, types.ErrorCode(def.Code), def.Message, def.Description).
				WithComponent(types.ComponentErrorType(def.Component)).
				WithResponseType(types.ResponseErrorType(def.ResponseType)).
				WithBundle(bundle)
	}
	localBlameManager.BlameDefinitions = blameDefinitionsMap

	return nil

}

/*
** These are internal errors function which uses
** local wrapper to determine the  error
 */

// InternalServerError is an internal server error.
func InternalServerError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorInternalServerError, WithCauses(cause))
}

// Bucket Errors
// BucketUploadError is an error when the bucket upload fails.
func BucketUploadError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorBucketUploadFailure, WithCauses(cause))
}

// BucketCredentialError is an error when the bucket credential fails.
func BucketCredentialError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorBucketCredentialFail, WithCauses(cause))
}

// File Errors
// FileNotFoundError is an error when the file is not found.
func FileNotFoundError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorFileUnavailable, WithCauses(cause))
}

// Parameter Errors
// MissingParameterError is an error when a required parameter is missing.
func MissingParameterError(name string) Blame {
	return getLocalBlameManager().FetchBlameForError(ParamMissing, WithField("name", name))
}

// MalformedParameterError is an error when a parameter is malformed.
func MalformedParameterError(name string) Blame {
	return getLocalBlameManager().FetchBlameForError(ParamMalformed, WithField("name", name))
}

// InvalidSourceError is an error when the source is invalid.
func InvalidSourceError(source string) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorInvalidSource, WithField("source", source))
}

// TypeConversionError is an error when the type conversion fails.
func TypeConversionError(name string, value string, targetType string, cause error) Blame {
	data := map[string]interface{}{
		"name":       name,
		"value":      value,
		"targetType": targetType,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorTypeConversion,
		WithFields(data),
		WithCauses(cause),
	)
}

// GinContextKeyFetchError is an error when the Gin context key is missing.
func GinContextKeyFetchError(key string, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorGinContextKeyMissing,
		WithField("key", key),
		WithCauses(cause),
	)
}

// ServiceContextFetchError is an error when the service context is missing.
func ServiceContextFetchError(key string, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorServiceContextMissing,
		WithField("key", key),
		WithCauses(cause),
	)
}

// MarshalingError is an error when marshaling fails.
func MarshalError(encodingType types.CodecType, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorMarshalFailed,
		WithField("type", encodingType.ToUpperCase()),
		WithCauses(cause),
	)
}

// UnMarshalingError is an error when unmarshaling fails.
func UnMarshalError(encodingType types.CodecType, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorUnmarshalFailed,
		WithField("type", encodingType.ToUpperCase()),
		WithCauses(cause),
	)
}

// PublishMessageError is an error when publishing a message fails.
func PublishMessageError(subject, message string, cause error) Blame {
	data := map[string]interface{}{
		"subject": subject,
		"message": message,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorPublishMessageFailed,
		WithFields(data),
		WithCauses(cause),
	)
}

// SubscribeToSubjectError is an error when subscribing to a subject fails.
func SubscribeToSubjectError(subject string, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorSubscribeToSubjectFailed,
		WithField("subject", subject),
		WithCauses(cause),
	)
}

// AlreadySubscribedToSubjectError is an error when the subject is already subscribed.
func AlreadySubscribedToSubjectError(subject string) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorAlreadySubscribedToSubject,
		WithField("subject", subject),
	)
}

// SubjectHandlerError is an error when the subject handler fails.
func SubjectHandlerError(subject string, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorSubjectHandlerFailed,
		WithField("subject", subject),
		WithCauses(cause),
	)
}

// UnsubscribeFailedError is an error when unsubscribing from a subject fails.
func UnsubscribeFailedError(subject string, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorUnsubscribeFailed,
		WithField("subject", subject),
		WithCauses(cause),
	)
}

// PublishRollbackEventFailedError is an error when publishing a rollback event fails.
func PublishRollbackEventFailedError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorPublishRollbackEventFailed,
		WithCauses(cause),
	)
}

// PublishEventToNextSubjectFailedError is an error when publishing an event to the next subject fails.
func PublishEventToNextSubjectFailedError(subject string, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorPublishEventToNextSubjectFailed,
		WithField("subject", subject),
		WithCauses(cause),
	)
}

// StepRollbackFailedError is an error when a step rollback fails.
func StepRollbackFailedError(step string, correlationId types.CorrelationID, cause error) Blame {
	data := map[string]interface{}{
		"step":           step,
		"correlation_id": correlationId,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorStepRollbackFailed,
		WithFields(data),
		WithCauses(cause),
	)
}

// UnknownCorrelationIDError is an error when an unknown correlation ID is encountered.
func UnknownCorrelationIDError(correlationID types.CorrelationID, cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorUnknownCorrelationId,
		WithField("correlation_id", correlationID),
		WithCauses(cause),
	)
}

// RequestBodyDataExtractionFailed is an error when request body data extraction fails.
func RequestBodyDataExtractionFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorRequestBodyDataExtractionFailed, WithCauses(cause))
}

// RequestFormDataExtractionFailed is an error when request form data extraction fails.
func RequestFormDataExtractionFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorRequestFormDataExtractionFailed, WithCauses(cause))
}

// BusinessIdPathParamMissing is an error when a business ID path parameter is missing.
func BusinessIdPathParamMissing(causes ...error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorBusinessIdPathParamMissing, WithCauses(causes...))
}

// TimeQueryParamInvalid is an error when a time query parameter is invalid.
func TimeQueryParamInvalid() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorTimeQueryParamInvalid)
}

// UserIdContextMissing is an error when a user ID  is missing from context.
func UserIdContextMissing(userIdField string, cause ...error) Blame {
	data := map[string]interface{}{
		"user_id": userIdField,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorUserIdContextMissing,
		WithFields(data),
		WithCauses(cause...),
	)
}

// UserIdQueryParamMissing is an error when a user ID query parameter is missing.
func UserIdQueryParamMissing(userIdField string, causes ...error) Blame {
	data := map[string]interface{}{
		"user_id": userIdField,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorUserIdQueryParamMissing,
		WithFields(data),
		WithCauses(causes...))
}

// BusinessIdHeaderMissing is an error when a business ID header is missing.
func BusinessIdHeaderMissing(businessIdField string, cause ...error) Blame {
	data := map[string]interface{}{
		"business_id": businessIdField,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorBusinessIdHeaderMissing,
		WithFields(data),
		WithCauses(cause...),
	)
}

// UserIdHeaderMissing is an error when a user ID header is missing.
func UserIdHeaderMissing(userIdField string, causes ...error) Blame {
	data := map[string]interface{}{
		"user_id": userIdField,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorUserIdHeaderMissing,
		WithFields(data),
		WithCauses(causes...))
}

// CorrelationIDHeaderMissing is an error when a correlation ID header is missing.
func CorrelationIDHeaderMissing(correlationIdField string, cause ...error) Blame {
	data := map[string]interface{}{
		"correlation_id": correlationIdField,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorCorrelationIDHeaderMissing,
		WithFields(data),
		WithCauses(cause...),
	)
}

// XSubjectHeaderMissing is an error when an X-Subject header is missing.
func XSubjectHeaderMissing(causes ...error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorXSubjectHeaderMissing,
		WithCauses(causes...))
}

// ServerStartFailed is an error when the server fails to start.
func ServerStartFailed(causes error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorServerStartFailed,
		WithCauses(causes))
}

// BusinessNotFound is an error when the business is not found.
func BusinessNotFound(causes error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorBusinessNotFound,
		WithCauses(causes))
}

// ConfigLoadFailure is an error when the config fails to load.
func ConfigLoadFailure(causes error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorConfigLoadFailure,
		WithCauses(causes))
}

// DatabaseOperationFailed is an error when a database operation fails.
func DatabaseOperationFailed(causes error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorDatabaseOperationFailed,
		WithCauses(causes))
}

// ServiceQueryParamMissing is an error when a service query parameter is missing.
func ServiceQueryParamMissing(serviceField string, causes ...error) Blame {
	data := map[string]interface{}{
		"service": serviceField,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorServiceQueryParamMissing,
		WithFields(data),
		WithCauses(causes...))
}

// MissingServiceName is an error when the service name is missing.
func MissingServiceName(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorServiceNameMissing, WithCauses(cause))
}

// RequestPayloadNil is an error when the request payload is nil.
func RequestPayloadNil(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorRequestPayloadNil, WithCauses(cause))
}

// StateExecutionFailed is an error when a state execution fails.
func StateExecutionFailed(stateName string, cause error) Blame {
	data := map[string]interface{}{
		"state": stateName,
	}
	return getLocalBlameManager().FetchBlameForError(
		ErrorStateExecutionFailed,
		WithFields(data),
		WithCauses(cause),
	)
}

// HeadersNotFound is an error when the headers are not found.
func HeadersNotFound(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorHeadersNotFound, WithCauses(cause))
}

// InactiveService is an error when the service is inactive.
func InactiveService(serviceName string) Blame {
	data := map[string]interface{}{
		"service": serviceName,
	}
	return getLocalBlameManager().FetchBlameForError(ErrorInactiveService, WithFields(data))
}

// ServiceDefinitionNotFound is an error when the service definition is not found.
func ServiceDefinitionNotFound(serviceName string, cause error) Blame {
	data := map[string]interface{}{
		"service": serviceName,
	}
	return getLocalBlameManager().FetchBlameForError(ErrorServiceDefinitionNotFound, WithFields(data), WithCauses(cause))
}

// GeneralKnownError is an error when we want to return any kind of error
func GeneralKnownError(cause error) Blame {
	data := map[string]any{
		"Error": cause.Error(),
	}
	return getLocalBlameManager().FetchBlameForError(ErrGeneralKnownError, WithCauses(cause), WithFields(data))
}

// DatabaseUnavailableError is an error when the database is unavailable.
func DatabaseUnavailableError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrDatabaseUnavailable,
		WithCauses(cause),
	)
}
