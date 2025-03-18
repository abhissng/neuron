package http

import (
	"net/http"

	"github.com/abhissng/neuron/utils/types"
)

// HTTP method constants
const (
	MethodGet     = http.MethodGet
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodDelete  = http.MethodDelete
	MethodPatch   = http.MethodPatch
	MethodOptions = http.MethodOptions
	MethodHead    = http.MethodHead
)

// ContentType constants
const (
	ContentTypeJSON              types.ContentType = "application/json"
	ContentTypeXML               types.ContentType = "application/xml"
	ContentTypeFormURLEncoded    types.ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipartFormData types.ContentType = "multipart/form-data"
	ContentTypeTextPlain         types.ContentType = "text/plain"
	ContentTypeOctetStream       types.ContentType = "application/octet-stream"
	ContentTypeHTML              types.ContentType = "text/html"
	ContentTypeJavaScript        types.ContentType = "application/javascript"
	ContentTypeCSV               types.ContentType = "text/csv"
	ContentTypePDF               types.ContentType = "application/pdf"
	ContentTypeZIP               types.ContentType = "application/zip"
	ContentTypeGZIP              types.ContentType = "application/gzip"
	ContentTypeYAML              types.ContentType = "application/x-yaml"
	ContentTypeMsgPack           types.ContentType = "application/msgpack"
	ContentTypePlainText         types.ContentType = "text/plain"
)
