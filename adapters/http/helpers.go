package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/types"
)

// **Decodes based on Content-Type using streaming**
func decodeByContentType[T any](contentType types.ContentType, reader io.Reader) (T, error) {
	switch contentType {
	case ContentTypeJSON:
		return tryDecode[T](decodeJSON, reader)
	case ContentTypeXML:
		return tryDecode[T](decodeXML, reader)
	case ContentTypeYAML:
		return tryDecode[T](decodeYAML, reader)
	case ContentTypeMsgPack:
		return tryDecode[T](decodeMsgPack, reader)
	case ContentTypePlainText:
		return tryDecode[T](decodeText, reader)
	}

	var zero T
	return zero, errors.New("unsupported content type: " + contentType.String())
}

// **Fallback decoder: Tries multiple formats sequentially**
func fallbackDecode[T any](reader io.Reader) (T, error) {
	if decoded, err := tryDecode[T](decodeJSON, reader); err == nil {
		return decoded, nil
	}
	if decoded, err := tryDecode[T](decodeXML, reader); err == nil {
		return decoded, nil
	}
	if decoded, err := tryDecode[T](decodeYAML, reader); err == nil {
		return decoded, nil
	}
	if decoded, err := tryDecode[T](decodeMsgPack, reader); err == nil {
		return decoded, nil
	}

	var zero T
	return zero, errors.New("failed to decode response using fallback methods")
}
func GetContentTypeFromResponse(resp *http.Response) types.ContentType {
	// Get response Content-Type
	contentType := resp.Header.Get("Content-Type")
	contentType = strings.Split(contentType, ";")[0] // Remove charset info
	return types.ContentType(contentType)
}
func GetDecoder(contentType types.ContentType) (DecoderFunc, error) {
	// Get appropriate decoder
	decoder, exists := decoders[contentType]
	if !exists {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return decoder, nil
}

// **Generic Decode Wrapper**
func tryDecode[T any](decodeFunc func(io.Reader, any) error, reader io.Reader) (T, error) {
	var result T
	if err := decodeFunc(reader, &result); err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

// createRequestBody creates the request body based on the content type
func (config *HttpClientManager) createRequestBody(data any) ([]byte, types.ContentType, error) {
	switch config.ContentType {
	case ContentTypeJSON:
		return config.handleJSONContent(data)
	case ContentTypeXML:
		return config.handleXMLContent(data)
	case ContentTypeFormURLEncoded:
		return config.handleFormURLEncodedContent()
	case ContentTypeMultipartFormData:
		return config.handleMultipartFormData()
	case ContentTypeTextPlain, ContentTypeHTML, ContentTypeJavaScript, ContentTypeCSV:
		return config.handleTextContent(data)
	default:
		return nil, "", errors.New("unsupported content type")
	}
}

// handleJSONContent encodes data as JSON
func (config *HttpClientManager) handleJSONContent(data any) ([]byte, types.ContentType, error) {
	jsonData, err := codec.Encode(data, codec.JSON)
	return jsonData, ContentTypeJSON, err
}

// handleXMLContent encodes data as XML
func (config *HttpClientManager) handleXMLContent(data any) ([]byte, types.ContentType, error) {
	xmlData, err := codec.Encode(data, codec.XML)
	return xmlData, ContentTypeXML, err
}

// handleFormURLEncodedContent creates form URL encoded content
func (config *HttpClientManager) handleFormURLEncodedContent() ([]byte, types.ContentType, error) {
	form := url.Values{}
	for key, value := range config.FormValues {
		form.Add(key, value)
	}
	return []byte(form.Encode()), ContentTypeFormURLEncoded, nil
}

// handleMultipartFormData creates multipart form data content
func (config *HttpClientManager) handleMultipartFormData() ([]byte, types.ContentType, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	// Add form fields
	for key, value := range config.FormValues {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", fmt.Errorf("error writing form field: %w", err)
		}
	}

	// Add files
	if err := config.addFilesToMultipartForm(writer); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("error closing writer: %w", err)
	}

	return buf.Bytes(), types.ContentType(writer.FormDataContentType()), nil
}

// addFilesToMultipartForm adds files to the multipart form writer
func (config *HttpClientManager) addFilesToMultipartForm(writer *multipart.Writer) error {
	for field, path := range config.Files {
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}

		err = func() error {
			defer file.Close()

			part, err := writer.CreateFormFile(field, file.Name())
			if err != nil {
				return fmt.Errorf("error creating form file: %w", err)
			}

			if _, err = io.Copy(part, file); err != nil {
				return fmt.Errorf("error copying file content: %w", err)
			}
			return nil
		}()

		if err != nil {
			return err
		}
	}
	return nil
}

// handleTextContent handles text-based content types
func (config *HttpClientManager) handleTextContent(data any) ([]byte, types.ContentType, error) {
	str, ok := data.(string)
	if !ok {
		return nil, "", errors.New("content must be a string for text-based types")
	}
	return []byte(str), config.ContentType, nil
}
