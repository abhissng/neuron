package blame

// CreateTokenFailed is an error when creating a token fails.
func CreateTokenFailed() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorCreateTokenFailed)
}

// CreateTokenIdFailed is an error when creating a token ID fails.
func CreateTokenIdFailed() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorCreateTokenIdFailed)
}

// MissingAuthCredential is an error when an auth credential is missing.
func MissingAuthCredential(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingAuthCredential, WithCauses(cause))
}

// MalformedAuthToken is an error when an auth token is malformed.
func MalformedAuthToken(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMalformedAuthToken, WithCauses(cause))
}

// ExpiredAuthToken is an error when an auth token expires.
func ExpiredAuthToken() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorExpiredAuthToken)
}

// UntrustedTokenIssuer is an error when an auth token issuer is untrusted.
func UntrustedTokenIssuer() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorUntrustedTokenIssuer)
}

// AuthPayloadInvalid is an error when an auth payload is invalid.
func AuthPayloadInvalid() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorAuthPayloadInvalid)
}

// AuthValidationFailed is an error when an auth validation fails.
func AuthValidationFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorAuthValidationFailed, WithCauses(cause))
}

// AuthSignatureMissing is an error when an auth signature is missing.
func AuthSignatureMissing() Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorAuthSignatureMissing,
	)
}

// AuthSignatureInvalid is an error when an auth signature is invalid.
func AuthSignatureInvalid() Blame {
	return getLocalBlameManager().FetchBlameForError(
		ErrorAuthSignatureInvalid,
	)
}

// MissingXUserRole is an error when the X-User-Role is missing.
func MissingXUserRole() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingXUserRole)
}

// MissingXOrgId is an error when the X-Org-Id is missing.
func MissingXOrgId() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingXOrgId)
}

// MissingXUserId is an error when the X-User-Id is missing.
func MissingXUserId() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingXUserId)
}

// SessionNotFound is an error when the session is not found.
func SessionNotFound() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorSessionNotFound)
}

// SessionMalformed is an error when the session is malformed.
func SessionMalformed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorSessionMalformed, WithCauses(cause))
}

// SessionValidationFailed is an error when the session validation fails.
func SessionValidationFailed(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorSessionValidationFailed, WithCauses(cause))
}

// SessionInvalid is an error when the session is invalid.
func SessionInvalid() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorSessionInvalid)
}

// SessionUnauthenticated is an error when the session is unauthenticated.
func SessionUnauthenticated() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorSessionUnauthenticated)
}

// MissingXLocationId is an error when the X-Location-Id is missing.
func MissingXLocationId() Blame {
	return getLocalBlameManager().FetchBlameForError(ErrorMissingXLocationId)
}

// UnAuthorizedAccess is an error when the user is unauthorized to access a resource.
func UnAuthorizedAccess(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrUnAuthorizedAccess, WithCauses(cause))
}

// UnAuthorizedUser is an error when the user is unauthorized to perform an action.
func UnAuthorizedUser(cause error) Blame {
	return getLocalBlameManager().FetchBlameForError(ErrUnAuthorizedUser, WithCauses(cause))
}
