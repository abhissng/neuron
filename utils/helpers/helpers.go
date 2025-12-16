package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/structures"
	"github.com/abhissng/neuron/utils/types"
	"github.com/biter777/countries"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nyaruka/phonenumbers"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"golang.org/x/text/language"
)

// isEmptyPrimitive checks if a primitive type value is empty.
// It returns (isEmpty, wasHandled) where wasHandled indicates if the type was recognized.
func isEmptyPrimitive(v reflect.Value) (bool, bool) {
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == "", true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0, true
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0, true
	case reflect.Bool:
		return !v.Bool(), true
	}
	return false, false
}

// isEmptyCollection checks if a collection type (slice, map, array, func) is empty.
// It returns (isEmpty, wasHandled) where wasHandled indicates if the type was recognized.
func isEmptyCollection(v reflect.Value) (bool, bool) {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0, true
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !IsEmpty(v.Index(i).Interface()) {
				return false, true
			}
		}
		return true, true
	}
	return false, false
}

// isEmptyStruct checks if a struct type is empty by recursively checking all fields.
// It handles time.Time as a special case and returns (isEmpty, wasHandled).
func isEmptyStruct(v reflect.Value) (bool, bool) {
	if v.Kind() != reflect.Struct {
		return false, false
	}

	// Check all struct fields recursively
	for i := 0; i < v.NumField(); i++ {
		// Skip unexported fields to avoid panics if necessary,
		// though IsEmpty generally handles interface conversion safely.
		if !IsEmpty(v.Field(i).Interface()) {
			return false, true
		}
	}
	return true, true
}

// isEmptyKnownType checks for specific named types that have well-defined empty states.
// It handles time.Time and uuid.UUID regardless of their underlying structure (Struct vs Array).
func isEmptyKnownType(v reflect.Value) (bool, bool) {
	if !v.CanInterface() {
		return false, false
	}

	switch val := v.Interface().(type) {
	case time.Time:
		return val.IsZero(), true
	case uuid.UUID:
		return val == uuid.Nil, true
	}
	return false, false
}

// IsEmpty checks if the given interface value represents an empty or zero value.
// It supports custom EmptyCheck interface and handles all Go types recursively.
func IsEmpty[T any](value T) bool {
	// Check if value implements EmptyCheck interface
	if v, ok := any(value).(types.EmptyCheck); ok {
		return v.IsEmpty()
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return true
	}

	// Handle pointer and interface types first
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	}

	// 1. Check Known Types (UUID, Time) - Added this priority check
	if isEmpty, ok := isEmptyKnownType(v); ok {
		return isEmpty
	}

	// 2. Check primitive types
	if isEmpty, ok := isEmptyPrimitive(v); ok {
		return isEmpty
	}

	// 3. Check collection types
	if isEmpty, ok := isEmptyCollection(v); ok {
		return isEmpty
	}

	// 4. Check struct types
	if isEmpty, ok := isEmptyStruct(v); ok {
		return isEmpty
	}

	// Default: Compare with zero value
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

// FetchErrorStrings extracts error messages from a slice of errors.
// It filters out nil errors and returns only the error message strings.
func FetchErrorStrings(errs []error) []string {
	errStrings := make([]string, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	return errStrings
}

// FetchErrorStack joins multiple errors into a single error string.
// It returns an empty string if no errors are provided or all are nil.
func FetchErrorStack(errs []error) string {
	err := JoinErrors(errs)
	if err == nil {
		return ""
	}
	return err.Error()
}

// JoinErrors combines multiple errors into a single error using errors.Join.
// It returns nil if no errors are provided or all errors are nil.
func JoinErrors(errs []error) error {
	err := errors.Join(errs...)
	if err == nil {
		return nil
	}
	return err
}

// FetchHTTPStatusCode maps response error types to their corresponding HTTP status codes.
// It returns 500 (Internal Server Error) for unknown response types.
func FetchHTTPStatusCode(response types.ResponseErrorType) int {
	switch response {
	case constant.BadRequest:
		return http.StatusBadRequest
	case constant.Unauthorized:
		return http.StatusUnauthorized
	case constant.Forbidden:
		return http.StatusForbidden
	case constant.NotFound:
		return http.StatusNotFound
	case constant.AlreadyExists:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// IsProdEnvironment checks if the current environment is production.
// It returns true for "prod" or "production" environment values.
func IsProdEnvironment() bool {
	switch GetEnvironment() {
	case "prod", "production":
		return true
	default:
		return false
	}
}

// ConvertToTimeDuration converts a value to time.Duration based on the specified unit.
// Supported units: "seconds", "minutes", "milliseconds", "hours".
func ConvertToTimeDuration(value int, unit string) time.Duration {
	switch unit {
	case "seconds":
		return time.Duration(value) * time.Second
	case "minutes":
		return time.Duration(value) * time.Minute
	case "milliseconds":
		return time.Duration(value) * time.Millisecond
	case "hours":
		return time.Duration(value) * time.Hour
	default:
		return 1
	}
}

// FormatRequestAndCorrelationIDs formats the request ID and correlation ID into a human-readable string.
func FormatRequestAndCorrelationIDs(requestId, correlationId string) string {
	return fmt.Sprintf("Request ID: %s, Correlation ID: %s", requestId, correlationId)
}

// GetDefaultPort returns the default port if DefaultAppPort is set in environment variables
func GetDefaultPort() string {
	port := os.Getenv(constant.DefaultAppPort)
	switch strings.TrimSpace(port) {
	case "":
		return "8001"
	default:
		return port
	}
}

// GetMaxConns returns the default value for MaxConns for postgres or sql.
func GetMaxConns(maxConn int) int {
	numCPU := runtime.NumCPU()
	if maxConn < 4 && numCPU < 4 {
		return 4
	}
	if maxConn < numCPU {
		return numCPU
	}
	return maxConn
}

// GetServiceName returns the service name from the app config or config files
func GetServiceName() string {
	return viper.GetString(constant.Service)
}

// GetDefaultLanguageTag returns the default language tag
func GetDefaultLanguageTag() types.LanguageTag {
	return types.LanguageTag(language.English)
}

// ParseLanguageTag parses a string into a language.Tag and returns a LanguageTag
func ParseLanguageTag(tagString string) types.LanguageTag {
	if tagString == "" {
		return GetDefaultLanguageTag()
	}
	// Parse the string into a language.Tag
	parsedTag, err := language.Parse(tagString)
	if err != nil {
		return GetDefaultLanguageTag()
	}
	return types.LanguageTag(parsedTag)
}

// NewBundle creates a new i18n.Bundle
func NewBundle(language types.LanguageTag) *i18n.Bundle {
	if IsEmpty(language) {
		language = GetDefaultLanguageTag()
	}
	return i18n.NewBundle(types.ToLanguageTag(language))
}

// UserHomeDir returns the user's home directory
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// ExtractBearerToken extracts the bearer token from the Authorization header
func ExtractBearerToken(authHeader string) string {
	if IsEmpty(authHeader) {
		return ""
	}

	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// GenerateReasonCode generates a namespaced reason code as a string
func GenerateReasonCode(namespace string, code int) string {
	if IsEmpty(namespace) {
		return fmt.Sprintf("%d", code)
	}
	namespace = strings.ToUpper(namespace)
	return fmt.Sprintf("%s-%d", namespace, code)
}

// RecoverException recovers from panics and logs the stack trace
func RecoverException(panic any) {
	if panic != nil {
		stack := debug.Stack()
		Println(constant.ERROR, "Exception occured", string(stack))
	}
}

// GetGoROOT returns the Go root directory
func GetGoROOT() string {
	return os.Getenv("GOROOT")
}

// GetHealthyMessageFor returns the healthy message for the given dependency
func GetHealthyMessageFor(dependency string) string {
	return dependency + " " + constant.HealthyStatusMessage
}

// IsFoundInSlice checks if the given key is found in the slice
// IsFoundInSlice checks if a key exists in a slice of strings or a slice of string pointers, ignoring case.
func IsFoundInSlice[T []string | []*string](key string, slice T) bool {
	// We use a type switch on the generic slice `T`.
	// The `any` cast is required to enable the type switch.
	switch s := any(slice).(type) {
	case []string:
		// The slice is of type []string.
		for _, v := range s {
			if strings.EqualFold(v, key) {
				return true
			}
		}
	case []*string:
		// The slice is of type []*string.
		for _, v := range s {
			// We must check for nil pointers and then dereference the pointer `*v` for the comparison.
			if v != nil && strings.EqualFold(*v, key) {
				return true
			}
		}
	}
	return false
}

// IsSuccess returns true if the given status is equal to constant.Success or constant.Completed
func IsSuccess(status types.Status) bool {
	return strings.EqualFold(status.String(), constant.Success.String()) || strings.EqualFold(status.String(), constant.Completed.String())
}

// GetEnvironment retrieves the current environment setting from various sources.
// It checks environment variables and viper configuration in order of priority.
func GetEnvironment() string {
	if os.Getenv(constant.Environment) != "" {
		return os.Getenv(constant.Environment)
	}

	if os.Getenv(constant.RunMode) != "" {
		return os.Getenv(constant.RunMode)
	}

	if viper.GetString(constant.Environment) != "" {
		return viper.GetString(constant.Environment)
	}

	return os.Getenv(constant.Environment)
}

// GetEnvironmentSlug normalizes environment names to standard slugs.
// It maps various environment name variations to consistent short forms.
func GetEnvironmentSlug(environment string) string {
	switch strings.ToLower(environment) {
	case "dev", "development":
		return "dev"
	case "test", "testing":
		return "test"
	case "staging":
		return "staging"
	case "prod", "production":
		return "prod"
	case "uat":
		return "uat"
	default:
		return "dev"
	}
}

// GetAvailablePort finds an available port for the given protocol (TCP or UDP).
func GetAvailablePort(protocol types.Protocol, preferredPort string) (string, error) {
	// If preferredPort is "0", find any free port dynamically
	if preferredPort == "0" || preferredPort == "" {
		port, err := findDynamicPort(protocol)
		if err != nil {
			return "0", fmt.Errorf("failed to find an available port: %w", err)
		}
		return strconv.Itoa(port), nil
	}

	// Check if the preferred port is available
	if isPortAvailable(protocol, preferredPort) {
		return preferredPort, nil
	}

	// Try finding the next available port
	preferredPortInt, _ := strconv.Atoi(preferredPort)
	for port := preferredPortInt + 1; port <= 65535; port++ {
		if isPortAvailable(protocol, strconv.Itoa(port)) {
			return strconv.Itoa(port), nil
		}
	}

	return "0", fmt.Errorf("no available ports found")
}

// findDynamicPort finds an available port dynamically for the specified protocol.
// It uses the OS to allocate a free port automatically.
func findDynamicPort(protocol types.Protocol) (int, error) {
	addr := fmt.Sprintf(":%d", 0)

	switch protocol {
	case constant.TCP:
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return 0, err
		}
		defer func() {
			_ = listener.Close()
		}()
		return listener.Addr().(*net.TCPAddr).Port, nil
	case constant.UDP:
		addr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return 0, err
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return 0, err
		}
		defer func() {
			_ = conn.Close()
		}()
		return conn.LocalAddr().(*net.UDPAddr).Port, nil
	default:
		return 0, fmt.Errorf("unsupported protocol: %s", protocol)
	}

}

// isPortAvailable tests if a specific port is available for the given protocol.
// It attempts to bind to the port and returns true if successful.
func isPortAvailable(protocol types.Protocol, port string) bool {
	addr := fmt.Sprintf(":%s", port)
	switch protocol {
	case constant.TCP:
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			defer func() {
				_ = listener.Close()
			}()
			return true
		}
	case constant.UDP:
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err == nil {
			conn, err := net.ListenUDP("udp", udpAddr)
			if err == nil {
				defer func() {
					_ = conn.Close()
				}()
				return true
			}
		}
	default:
		return false
	}

	return false
}

// ValidateURL checks if the provided string is a valid URL.
// It uses url.ParseRequestURI for validation and returns an error if invalid.
func ValidateURL(requestURL string) error {
	_, err := url.ParseRequestURI(requestURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

// ConstructURLWithParams builds a URL by appending query parameters to a base URL.
// It handles various parameter types and properly encodes them.
func ConstructURLWithParams(baseURL string, params map[string]any) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	query := parsedURL.Query()
	for key, value := range params {
		switch v := value.(type) {
		case string:
			query.Set(key, v)
		case bool:
			query.Set(key, fmt.Sprintf("%t", v)) // Converts `true`/`false` to string
		case int, int8, int16, int32, int64:
			query.Set(key, fmt.Sprintf("%d", v)) // Converts integers to string
		case float32, float64:
			query.Set(key, fmt.Sprintf("%f", v)) // Converts floats to string
		default:
			query.Set(key, fmt.Sprintf("%v", v)) // Default case for unknown types
		}
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// CreateLogDirectory creates the log directory and file for the service.
// It ensures proper permissions and returns the log file path.
func CreateLogDirectory() string {
	serviceName := GetServiceName()
	// Define the log file path
	logFilePath := "/var/log/" + serviceName + ".log"

	// Ensure the log directory exists
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0750); err != nil {
		Println(constant.ERROR, "failed to create log directory: ", err)
		return logFilePath
	}

	// Create the log file if it doesn't exist
	file, err := os.OpenFile(filepath.Clean(logFilePath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		Println(constant.ERROR, "failed to create log file: ", err)
		return logFilePath
	}
	defer func() {
		if err := file.Close(); err != nil {
			Println(constant.ERROR, "Error closing file: ", err)
		}
	}()

	// Ensure the log file has the correct permissions
	if err := os.Chmod(logFilePath, 0600); err != nil && !os.IsNotExist(err) {
		Println(constant.ERROR, "failed to set log file permissions: ", err)
		return logFilePath
	}
	return logFilePath
}

// GetIsLogRotationEnabled checks if log rotation is enabled via environment variables.
// It parses the LOG_ROTATION_ENABLED environment variable as a boolean.
func GetIsLogRotationEnabled() bool {
	enableRotation, _ := strconv.ParseBool(os.Getenv(constant.LogRotationEnabled))
	return enableRotation
}

// Println prints a colored log message with timestamp and log level.
// It exits the program with code 1 if the mode is FATAL.
func Println(mode types.LogMode, args ...any) {
	var color string
	switch mode {
	case constant.INFO:
		color = constant.GreenColor
	case constant.WARN:
		color = constant.YellowColor
	case constant.ERROR, constant.FATAL:
		color = constant.RedColor
	case constant.DEBUG:
		color = constant.BlueColor
	default:
		color = constant.ResetColor // Default to no color
	}

	// Get current time and format it (e.g., "2025-03-04 15:30:45")
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	fmt.Println(color + "[" + timestamp + "] [" + mode.String() + "] " + fmt.Sprint(args...) + constant.ResetColor)
	if mode == constant.FATAL {
		os.Exit(1)
	}
}

// Printf prints a formatted colored log message with timestamp and log level.
// It exits the program with code 1 if the mode is FATAL.
func Printf(mode types.LogMode, format string, args ...interface{}) {
	var color string
	switch mode {
	case constant.INFO:
		color = constant.GreenColor
	case constant.WARN:
		color = constant.YellowColor
	case constant.ERROR, constant.FATAL:
		color = constant.RedColor
	case constant.DEBUG:
		color = constant.BlueColor
	default:
		color = constant.ResetColor // Default to no color
	}

	// Get current time and format it (e.g., "2025-03-04 15:30:45")
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	format = "[" + timestamp + "] [" + mode.String() + "] " + format
	fmt.Printf(color+format+constant.ResetColor, args...)
	if mode == constant.FATAL {
		os.Exit(1)
	}
}

// CorrelationIDFromNatsMsg extracts the correlation ID from NATS message headers.
// It returns the correlation ID used for distributed tracing.
func CorrelationIDFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.CorrelationIDHeader)
}

// MessageIDFromNatsMsg extracts the message ID from NATS message headers.
// It returns the unique message identifier for idempotency handling.
func MessageIDFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.MessageIdHeader)
}

// AuthorizationHeaderFromNatsMsg extracts the authorization header from NATS message.
// It returns the authorization token for authentication purposes.
func AuthorizationHeaderFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.AuthorizationHeader)
}

// IPHeaderFromNatsMsg extracts the IP address header from NATS message.
// It returns the client IP address for logging and security purposes.
func IPHeaderFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.IPHeader)
}

// ErrorHeadeFromNatsMsg extracts the error header from NATS message.
// It returns error information passed through message headers.
func ErrorHeadeFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.ErrorHeader)
}

// GetIssuerFromConfig retrieves the JWT issuer from configuration.
// It returns the issuer identifier used for token validation.
func GetIssuerFromConfig() string {
	return viper.GetString(constant.IssuerKey)
}

// ConstructDiscoveryURL builds a service discovery URL by combining base URL and service name.
// It returns a properly formatted URL for service discovery.
func ConstructDiscoveryURL(URL string, serviceName string) string {
	u, err := url.Parse(URL + serviceName)
	if err != nil {
		return ""
	}
	return u.String()
}

// GetIsOpenSearchEnabled checks if OpenSearch logging is enabled via environment variable.
func GetIsOpenSearchEnabled() bool {
	if os.Getenv(constant.OpenSearchEnabled) != "" {
		return os.Getenv(constant.OpenSearchEnabled) == "true"
	}

	if viper.GetString(constant.OpenSearchEnabled) != "" {
		return viper.GetString(constant.OpenSearchEnabled) == "true"
	}

	return false
}

// GetOpenSearchAddresses returns a list of OpenSearch node addresses.
func GetOpenSearchAddresses() []string {
	// Example: "http://localhost:9200,http://localhost:9201"
	addresses := os.Getenv(constant.OpenSearchAddresses)
	if addresses == "" {
		addresses = viper.GetString(constant.OpenSearchAddresses)
	}
	if addresses == "" {
		return []string{"http://localhost:9200"} // Default for local dev
	}
	return strings.Split(addresses, ",")
}

// GetOpenSearchIndexName returns the index name for logs.
func GetOpenSearchIndexName() string {
	return formatIndexName(GetServiceName())
}

// GetOpenSearchUsername returns the username for authentication.
func GetOpenSearchUsername() string {
	if os.Getenv(constant.OpenSearchUsername) != "" {
		return os.Getenv(constant.OpenSearchUsername)
	}

	if viper.GetString(constant.OpenSearchUsername) != "" {
		return viper.GetString(constant.OpenSearchUsername)
	}

	return ""
}

// GetOpenSearchPassword returns the password for authentication.
func GetOpenSearchPassword() string {
	if os.Getenv(constant.OpenSearchPassword) != "" {
		return os.Getenv(constant.OpenSearchPassword)
	}

	if viper.GetString(constant.OpenSearchPassword) != "" {
		return viper.GetString(constant.OpenSearchPassword)
	}

	return ""
}

// formatIndex name format the index name for the service
func formatIndexName(serviceName string) string {
	environment := GetEnvironmentSlug(GetEnvironment())
	if strings.HasSuffix(serviceName, "-service") {
		return serviceName + "-" + environment + "-logs"
	}
	return serviceName + "-service-" + environment + "-logs"
}

// Valid returns a pointer to true, commonly used for boolean pointer fields.
// It provides a convenient way to set boolean pointer values.
func Valid() *bool {
	valid := true
	return &valid
}

// MustGetEnv retrieves a required environment variable or exits the program.
// If the variable is not set or empty, it logs a fatal error and exits with code 1.
func MustGetEnv(key string) string {
	value := os.Getenv(key)

	if value == "" || strings.TrimSpace(value) == "" {
		// In a real application, you would use a proper logging system (like Zap)
		// and maybe the logging's Fatal method here.
		Printf(constant.FATAL, "FATAL ERROR: Required environment variable '%s' is not set or is empty.\n", key)
	}

	return value
}

// IsURL checks if the given string starts with http:// or https://.
// It performs a simple prefix check to identify URLs.
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// ToNetIPAddr converts a remote address string to a netip.Addr.
// It handles addresses with or without ports and validates IP format.
func ToNetIPAddr(remoteAddress string) (*netip.Addr, error) {
	var host string

	// Try to split if remoteAddress contains port (e.g. "192.168.0.1:5000")
	if strings.Contains(remoteAddress, ":") {
		h, _, err := net.SplitHostPort(remoteAddress)
		if err != nil {
			// It might still be a plain IP like "::1" or malformed
			host = remoteAddress
		} else {
			host = h
		}
	} else {
		host = remoteAddress
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", remoteAddress)
	}

	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return nil, fmt.Errorf("failed to convert %s to netip.Addr", remoteAddress)
	}

	return &ipAddr, nil
}

// NormalizePrecision rounds a float64 value to the specified number of decimal places.
// It uses math.Pow10 and math.Round for precise decimal rounding.
func NormalizePrecision(val float64, digits int) float64 {
	scale := math.Pow10(digits)
	return math.Round(val*scale) / scale
}

// DetectInputType analyzes input to determine if it's a file path or JSON string.
// It returns "file", "json", or empty string for unknown types.
func DetectInputType(input string) string {
	input = strings.TrimSpace(input)

	// ---- Case 1: JSON FIRST (cheap structural heuristic) ----
	if len(input) > 1 {
		first := input[0]
		last := input[len(input)-1]
		if (first == '{' && last == '}') || (first == '[' && last == ']') {
			return "json"
		}
	}

	// ---- Case 2: Ensure real file existence ----
	if fi, err := os.Stat(input); err == nil && !fi.IsDir() {
		return "file"
	}

	// ---- Case 3: Heuristic file-path detection ----
	// Only if it looks like a path (contains slash or backslash)
	// AND has a valid extension
	if (strings.Contains(input, "/") || strings.Contains(input, `\`)) &&
		filepath.Ext(input) != "" {
		return "file"
	}

	return ""
}

// ParsePhoneNumber parses and validates a phone number string using libphonenumber.
// defaultRegion is a two-letter ISO country code used for numbers without country codes.
// It returns detailed phone number information including validation status.
func ParsePhoneNumber(rawNumber, defaultRegion string) (*structures.PhoneNumberInfo, error) {
	// 1. Parse the number using the phonenumbers library
	num, err := phonenumbers.Parse(rawNumber, strings.ToUpper(defaultRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to parse number '%s' (region: %s): %w", rawNumber, defaultRegion, err)
	}

	// 2. Get the region code from the parsed number
	regionCode, countryName := GetCountryInfoForCode(int(num.GetCountryCode()))

	// 3. Create the info struct with all details
	info := &structures.PhoneNumberInfo{
		E164Format:     phonenumbers.Format(num, phonenumbers.E164),
		CountryCode:    num.GetCountryCode(),
		RegionCode:     regionCode,
		CountryName:    countryName,
		IsValid:        phonenumbers.IsValidNumber(num),
		NationalNumber: num.GetNationalNumber(),
	}

	return info, nil
}

// GetRegionForCountryCode returns the primary region code for a given country dialing code.
// Note: Some country codes map to multiple regions (e.g., +1 for US, CA, etc.).
func GetRegionForCountryCode(countryCode int) string {
	// This returns the *main* region for that code (e.g., 1 -> "US")
	return phonenumbers.GetRegionCodeForCountryCode(countryCode)
}

// GetCountryInfoForCode returns both region code and country name for a dialing code.
// It provides human-readable country information for phone number display.
func GetCountryInfoForCode(countryCode int) (regionCode string, countryName string) {
	// Get the *primary* region for this dialing code
	regionCode = phonenumbers.GetRegionCodeForCountryCode(countryCode)

	country := countries.ByName(regionCode)
	if country.IsValid() {
		countryName = country.String()
	} else {
		countryName = "N/A"
	}
	return regionCode, countryName
}

// TailCallerEncoder creates a custom zap caller encoder that shows the last n path segments.
// It provides more readable file paths in log output by trimming deep directory structures.
func TailCallerEncoder(n int) zapcore.CallerEncoder {
	if n <= 0 {
		return zapcore.ShortCallerEncoder
	}
	return func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		path := caller.File

		// Scan from the end; stop after hitting 4 separators (/ or \)
		sep := 0
		i := len(path) - 1
		for ; i >= 0; i-- {
			c := path[i]
			if c == '/' || c == '\\' {
				sep++
				if sep == n {
					break
				}
			}
		}
		start := i + 1
		if start < 0 || start > len(path) {
			start = 0
		}
		tail := path[start:]

		// Normalize only if needed (Windows paths)
		if strings.IndexByte(tail, '\\') >= 0 {
			tail = strings.ReplaceAll(tail, "\\", "/")
		}

		// Build "tail:line" with minimal overhead
		var sb strings.Builder
		sb.Grow(len(tail) + 12)
		sb.WriteString(tail)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(caller.Line))

		enc.AppendString(sb.String())
	}
}

// IsLocalhostHost checks if the host is localhost, 127.0.0.1, or ::1.
func IsLocalhostHost(host string) bool {
	h := strings.Split(host, ":")[0] // strip possible port
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

// HostFromOrigin extracts the host from an Origin header, handling ports and URL parsing.
func HostFromOrigin(origin string) string {
	if origin == "" {
		return ""
	}
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Split(u.Host, ":")[0])
}

// SameSiteName converts http.SameSite to a human-readable string.
func SameSiteName(s http.SameSite) string {
	switch s {
	case http.SameSiteDefaultMode:
		return "Default"
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return "Unknown"
	}
}

// replacePlaceholders replaces {{.key}} in src with values from params.
func ReplacePlaceholders(src string, params map[string]any) (string, error) {
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
	}

	tmpl, err := template.New("template").Funcs(funcMap).Parse(src)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SplitAny splits a string into parts using any of the specified separators.
// It returns a slice of non-empty strings after removing leading/trailing whitespace.
func SplitAny(s string, delims ...string) []string {
	if s == "" {
		return nil
	}

	if len(delims) == 0 {
		delims = []string{"||", ",", "|", ";"}
	}

	var result []string

	for len(s) > 0 {
		idx := -1
		dlen := 0

		// Find earliest delimiter match
		for _, d := range delims {
			if i := strings.Index(s, d); i >= 0 && (idx == -1 || i < idx) {
				idx = i
				dlen = len(d)
			}
		}

		if idx == -1 {
			// No more delimiters
			s = strings.TrimSpace(s)
			if s != "" {
				result = append(result, s)
			}
			break
		}

		part := strings.TrimSpace(s[:idx])
		if part != "" {
			result = append(result, part)
		}

		// Move forward past delimiter
		s = s[idx+dlen:]
	}

	return result
}
