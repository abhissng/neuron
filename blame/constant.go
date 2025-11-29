package blame

import (
	"github.com/abhissng/neuron/utils/types"
)

const (
	ReasonCodeNameSpace = "INTLIB"
	ReasonCodeBase      = 100000
)

// Error Identifiers for internal library
const (
	ErrorInternalServerError             types.ErrorCode = "error-internal-server-error"
	ErrorBucketUploadFailure             types.ErrorCode = "error-bucket-upload-failure"
	ErrorBucketCredentialFail            types.ErrorCode = "error-bucket-credential-failure" // #nosec G101
	ErrorFileUnavailable                 types.ErrorCode = "error-file-unavailable"
	ParamMissing                         types.ErrorCode = "param-not-found"
	ParamMalformed                       types.ErrorCode = "param-malformed"
	ErrorInvalidSource                   types.ErrorCode = "error-source-invalid"
	ErrorTypeConversion                  types.ErrorCode = "error-type-conversion"
	ErrorGinContextKeyMissing            types.ErrorCode = "gin-context-key-not-found"
	ErrorServiceContextMissing           types.ErrorCode = "service-context-not-found"
	ErrorMarshalFailed                   types.ErrorCode = "error-marshal-failed"
	ErrorUnmarshalFailed                 types.ErrorCode = "error-unmarshal-failed"
	ErrorPublishMessageFailed            types.ErrorCode = "error-publish-message-failed"
	ErrorAlreadySubscribedToSubject      types.ErrorCode = "error-already-subscribed-to-subject"
	ErrorSubscribeToSubjectFailed        types.ErrorCode = "error-subscribe-to-subject-failed"
	ErrorSubjectHandlerFailed            types.ErrorCode = "error-subject-handler-failed"
	ErrorUnsubscribeFailed               types.ErrorCode = "error-unsubscribe-failed"
	ErrorPublishRollbackEventFailed      types.ErrorCode = "error-publish-rollback-event-failed"
	ErrorPublishEventToNextSubjectFailed types.ErrorCode = "error-publish-event-to-next-subject-failed"
	ErrorStepRollbackFailed              types.ErrorCode = "error-step-rollback-failed"
	ErrorUnknownCorrelationId            types.ErrorCode = "error-unknown-correlation-id"
	ErrorCreateTokenFailed               types.ErrorCode = "error-create-token-failed"
	ErrorCreateTokenIdFailed             types.ErrorCode = "error-create-token-id-failed"
	ErrorMissingAuthCredential           types.ErrorCode = "error-missing-auth-credential" // #nosec G101
	ErrorMalformedAuthToken              types.ErrorCode = "error-malformed-auth-token"    // #nosec G101
	ErrorExpiredAuthToken                types.ErrorCode = "error-expired-auth-token"      // #nosec G101
	ErrorUntrustedTokenIssuer            types.ErrorCode = "error-untrusted-token-issuer"  // #nosec G101
	ErrorAuthPayloadInvalid              types.ErrorCode = "error-auth-payload-invalid"
	ErrorAuthValidationFailed            types.ErrorCode = "error-auth-validation-failed"
	ErrorRequestBodyDataExtractionFailed types.ErrorCode = "error-request-body-data-extraction-failed"
	ErrorRequestFormDataExtractionFailed types.ErrorCode = "error-form-data-extraction-failed"
	ErrorBusinessIdPathParamMissing      types.ErrorCode = "error-business-id-path-param-missing"
	ErrorTimeQueryParamInvalid           types.ErrorCode = "error-time-query-param-invalid"
	ErrorUserIdContextMissing            types.ErrorCode = "error-user-id-context-missing"
	ErrorUserIdQueryParamMissing         types.ErrorCode = "error-user-id-query-param-missing"
	ErrorBusinessIdHeaderMissing         types.ErrorCode = "error-business-id-header-missing"
	ErrorUserIdHeaderMissing             types.ErrorCode = "error-user-id-header-missing"
	ErrorCorrelationIDHeaderMissing      types.ErrorCode = "error-correlation-id-header-missing"
	ErrorAuthSignatureMissing            types.ErrorCode = "error-auth-signature-missing" // #nosec G101
	ErrorAuthSignatureInvalid            types.ErrorCode = "error-auth-signature-invalid" // #nosec G101
	ErrorXSubjectHeaderMissing           types.ErrorCode = "error-x-subject-header-missing"
	ErrorServerStartFailed               types.ErrorCode = "error-server-start-failed"
	ErrorRequestBodyInvalid              types.ErrorCode = "error-request-body-invalid"
	ErrorBusinessNotFound                types.ErrorCode = "error-business-not-found"
	ErrorConfigLoadFailure               types.ErrorCode = "error-config-load-failure"
	ErrorDatabaseOperationFailed         types.ErrorCode = "error-database-operation-failed"
	ErrorServiceQueryParamMissing        types.ErrorCode = "error-service-query-param-missing"
	ErrorServiceNameMissing              types.ErrorCode = "error-service-name-missing"
	ErrorRequestPayloadNil               types.ErrorCode = "error-request-payload-nil"
	ErrorStateExecutionFailed            types.ErrorCode = "error-state-execution-failed"
	ErrorHeadersNotFound                 types.ErrorCode = "error-headers-not-found"
	ErrorInactiveService                 types.ErrorCode = "error-inactive-service"
	ErrorServiceDefinitionNotFound       types.ErrorCode = "error-service-definition-not-found"
	ErrorURLValidationFailed             types.ErrorCode = "error-url-validation-failed"
	ErrorURLParsingFailed                types.ErrorCode = "error-url-parsing-failed"
	ErrorURLConstructionFailed           types.ErrorCode = "error-url-construction-failed"
	ErrorCreateRequestBodyFailed         types.ErrorCode = "error-create-request-body-failed"
	ErrorCreateHTTPRequestFailed         types.ErrorCode = "error-create-http-request-failed"
	ErrorCreateHTTPClientFailed          types.ErrorCode = "error-create-http-client-failed"
	ErrorDecodeResponseFailed            types.ErrorCode = "error-decode-response-failed"
	ErrorResponseResultError             types.ErrorCode = "error-response-result-error"
	ErrorMissingCorrelationID            types.ErrorCode = "error-missing-correlation-id"
	ErrorMissingRecordsName              types.ErrorCode = "error-missing-records-name"
	ErrorMissingXUserRole                types.ErrorCode = "error-missing-x-user-role"
	ErrorMissingXOrgId                   types.ErrorCode = "error-missing-x-org-id"
	ErrorMissingXUserId                  types.ErrorCode = "error-missing-x-user-id"
	ErrorSessionNotFound                 types.ErrorCode = "error-session-not-found"
	ErrorSessionMalformed                types.ErrorCode = "error-session-malformed"
	ErrorSessionValidationFailed         types.ErrorCode = "error-session-validation-failed"
	ErrorSessionInvalid                  types.ErrorCode = "error-session-invalid"
	ErrorSessionUnauthenticated          types.ErrorCode = "error-session-unauthenticated"
	ErrorMissingFeatureFlags             types.ErrorCode = "error-missing-feature-flags"
	ErrorMissingXLocationId              types.ErrorCode = "error-missing-x-location-id"
)
