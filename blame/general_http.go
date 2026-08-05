package blame

// RequestBodyInvalid is an error when the request body is invalid.
func RequestBodyInvalid(causes error) Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorRequestBodyInvalid,
		WithCauses(causes))
}

// URLValidationFailed is an error when the URL validation fails.
func URLValidationFailed(url string, cause error) Blame {
	data := map[string]interface{}{
		"url": url,
	}
	return getLocalBlameManager().FetchBlameForError(ErrorURLValidationFailed, WithFields(data), WithCauses(cause))
}

// URLParsingFailed is an error when the URL parsing fails.
func URLParsingFailed(url string, cause error) Blame {
	data := map[string]interface{}{
		"url": url,
	}
	return getLocalBlameManager().FetchBlameForError(ErrorURLParsingFailed, WithFields(data), WithCauses(cause))
}

// URLConstructionFailed is an error when the URL construction fails.
func URLConstructionFailed(url string, queryParams map[string]any, cause error) Blame {
	data := map[string]interface{}{
		"url":         url,
		"queryParams": queryParams,
	}
	return getLocalBlameManager().FetchBlameForError(ErrorURLConstructionFailed, WithFields(data), WithCauses(cause))
}

// CreateRequestBodyFailed is an error when the request body creation fails.
func CreateRequestBodyFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorCreateRequestBodyFailed, WithCauses(cause))
}

// CreateHTTPRequestFailed is an error when the HTTP request creation fails.
func CreateHTTPRequestFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorCreateHTTPRequestFailed, WithCauses(cause))
}

// CreateHTTPClientFailed is an error when the HTTP client creation fails.
func CreateHTTPClientFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorCreateHTTPClientFailed, WithCauses(cause))
}

// DecodeResponseFailed is an error when the response decoding fails.
func DecodeResponseFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorDecodeResponseFailed, WithCauses(cause))
}

// ResponseResultError is an error when the response result has an error.
func ResponseResultError(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorResponseResultError, WithCauses(cause))
}

// MissingCorrelationID is an error when the correlation ID is missing.
func MissingCorrelationID() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingCorrelationID)
}

// MissingRecordsName is an error when the records name is missing.
func MissingRecordsName(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingRecordsName, WithCauses(cause))
}

// MissingFeatureFlags is an error when the feature flags are missing.
func MissingFeatureFlags() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingFeatureFlags)
}
