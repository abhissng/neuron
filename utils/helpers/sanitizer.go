package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

const (
	defaultMaxDepth = 8
	// defaultMaxBodySize = 20 << 10 // 20 KB
	defaultMaskValue = "****"
)

// Default blocked key names for audit sanitization (used when no WithBlockedKeys is provided).
var defaultBlockedKeys = []string{
	"password", "token", "access_token", "refresh_token",
	"authorization", "cookie", "secret", "otp",
	"api_key", "api_secret", "api_token", "api_password",
	"api_username",
}

// DefaultSanitizer is a sanitizer with default blocked keys for use when no config is needed.
var DefaultSanitizer = NewSanitizer()

// Sanitizer masks sensitive fields in values for safe audit logging.
// Use NewSanitizer with options to configure blocked keys and limits.
type Sanitizer struct {
	blockedKeys map[string]struct{}
	maxDepth    int
	maskValue   string
}

// SanitizeOption configures a Sanitizer.
type SanitizeOption func(*Sanitizer)

// WithBlockedKeys sets the field/key names to mask (case-insensitive).
// If no keys are provided, the default blocked keys are used.
func WithBlockedKeys(keys ...string) SanitizeOption {
	return func(s *Sanitizer) {
		if len(s.blockedKeys) == 0 {
			s.blockedKeys = make(map[string]struct{}, len(keys))
		}
		for _, k := range keys {
			if k != "" {
				s.blockedKeys[strings.ToLower(k)] = struct{}{}
			}
		}
	}
}

// WithMaxDepth sets the maximum recursion depth (default 8). Beyond this, value is "[truncated]".
func WithMaxDepth(depth int) SanitizeOption {
	return func(s *Sanitizer) {
		if depth > 0 {
			s.maxDepth = depth
		}
	}
}

// WithMaskValue sets the string used to replace blocked values (default "***").
func WithMaskValue(v string) SanitizeOption {
	return func(s *Sanitizer) {
		if v == "" {
			v = strings.Repeat("*", len(defaultMaskValue))
		}
		s.maskValue = v
	}
}

// NewSanitizer creates a Sanitizer with the given options.
// With no options, uses default blocked keys (password, token, authorization, etc.).
func NewSanitizer(opts ...SanitizeOption) *Sanitizer {
	s := &Sanitizer{
		blockedKeys: make(map[string]struct{}, len(defaultBlockedKeys)),
		maxDepth:    defaultMaxDepth,
		maskValue:   defaultMaskValue,
	}
	for _, k := range defaultBlockedKeys {
		s.blockedKeys[k] = struct{}{}
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Sanitize returns a copy of v with blocked keys masked for safe audit logging.
// If sanitization panics (e.g. unsupported type), the original value is returned so logging does not fail.
func (s *Sanitizer) Sanitize(v any) any {
	if s == nil || len(s.blockedKeys) == 0 {
		return v
	}
	out := v
	func() {
		defer func() {
			if recover() != nil {
				out = v
			}
		}()
		visited := make(map[uintptr]bool)
		out = s.sanitize(reflect.ValueOf(v), 0, visited)
	}()
	return out
}

func (s *Sanitizer) sanitize(v reflect.Value, depth int, visited map[uintptr]bool) any {
	if !v.IsValid() {
		return nil
	}
	if depth > s.maxDepth {
		return "[truncated]"
	}
	// Unwrap pointers and interfaces so we sanitize the concrete value (e.g. when v is passed as any)
	for {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				return nil
			}
			ptr := v.Pointer()
			if visited[ptr] {
				return "[circular]"
			}
			visited[ptr] = true
			v = v.Elem()
		case reflect.Interface:
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
		default:
			goto done
		}
	}
done:
	switch v.Kind() {
	case reflect.Struct:
		out := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := field.Name
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				name = strings.Split(tag, ",")[0]
			}
			if _, blocked := s.blockedKeys[strings.ToLower(name)]; blocked {
				out[name] = s.maskValue
				continue
			}
			out[name] = s.sanitize(v.Field(i), depth+1, visited)
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		ptr := v.Pointer()
		if visited[ptr] {
			return "[circular]"
		}
		visited[ptr] = true
		out := make(map[string]any)
		for _, key := range v.MapKeys() {
			k := fmt.Sprint(key.Interface())
			if _, blocked := s.blockedKeys[strings.ToLower(k)]; blocked {
				out[k] = s.maskValue
				continue
			}
			out[k] = s.sanitize(v.MapIndex(key), depth+1, visited)
		}
		return out
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return s.sanitizeJSONBytes(v.Bytes(), visited)
		}
		if v.Kind() == reflect.Slice {
			ptr := v.Pointer()
			if visited[ptr] {
				return "[circular]"
			}
			visited[ptr] = true
		}
		out := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = s.sanitize(v.Index(i), depth+1, visited)
		}
		return out
	case reflect.String:
		return s.sanitizeJSONString(v.String(), visited)
	default:
		return v.Interface()
	}
}

// sanitizeJSONBytes parses b as JSON; if successful, sanitizes and returns the marshaled string (no string alloc for input).
// Use json.RawMessage(body) at call sites when the payload is JSON to avoid string(body) allocation.
func (s *Sanitizer) sanitizeJSONBytes(b []byte, visited map[uintptr]bool) string {
	if len(b) == 0 {
		return "[binary]"
	}
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return "[binary]"
	}
	var parsed any
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "[binary]"
	}
	sanitized := s.sanitize(reflect.ValueOf(parsed), 0, visited)
	out, err := json.Marshal(sanitized)
	if err != nil {
		return "[binary]"
	}
	return string(out)
}

// sanitizeJSONString parses s as JSON; if successful, sanitizes the parsed value and returns the marshaled string.
// If s does not look like JSON or unmarshal fails, returns s unchanged so non-JSON strings are left as-is.
func (s *Sanitizer) sanitizeJSONString(str string, visited map[uintptr]bool) string {
	trimmed := strings.TrimSpace(str)
	if len(trimmed) == 0 {
		return str
	}
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return str
	}
	var parsed any
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		return str
	}
	sanitized := s.sanitize(reflect.ValueOf(parsed), 0, visited)
	out, err := json.Marshal(sanitized)
	if err != nil {
		return str
	}
	return string(out)
}

// ReadBodySafe reads up to the sanitizer's maxBodySize from r.Body and restores
// the body so handlers can still read it. Skips multipart/form-data.
func (s *Sanitizer) ReadBodySafe(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		return nil, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// SanitizeWith returns a copy of v with sensitive fields masked using the given sanitizer.
// If s is nil, DefaultSanitizer is used.
func SanitizeWith(s *Sanitizer, v any) any {
	if s == nil {
		s = DefaultSanitizer
	}
	return s.Sanitize(v)
}

// SanitizeAny returns a copy of v with default sensitive fields masked.
// For custom masking, use SanitizeWith(yourSanitizer, v) or NewSanitizer(opts...) and Sanitize.
func SanitizeAny(v any) any {
	return SanitizeWith(nil, v)
}

// ReadBodySafeWith reads and restores r.Body using the given sanitizer's config.
// If s is nil, DefaultSanitizer is used. Skips multipart/form-data.
func ReadBodySafeWith(s *Sanitizer, r *http.Request) ([]byte, error) {
	if s == nil {
		s = DefaultSanitizer
	}
	return s.ReadBodySafe(r)
}

// ReadBodySafe reads and restores the body. Skips multipart/form-data.
// For custom config, use ReadBodySafeWith(yourSanitizer, r) or NewSanitizer(opts...) and s.ReadBodySafe(r).
func ReadBodySafe(r *http.Request) ([]byte, error) {
	return ReadBodySafeWith(nil, r)
}
