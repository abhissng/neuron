package helpers

import (
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
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/structures"
	"github.com/abhissng/neuron/utils/types"
	"github.com/biter777/countries"
	"github.com/nats-io/nats.go"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nyaruka/phonenumbers"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"golang.org/x/text/language"
)

// isEmptyPrimitive handles primitive type checks
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

// isEmptyCollection handles collection type checks
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

// isEmptyStruct handles struct type checks
func isEmptyStruct(v reflect.Value) (bool, bool) {
	if v.Kind() != reflect.Struct {
		return false, false
	}

	// Handle time.Time separately
	if v.Type() == reflect.TypeOf(time.Time{}) {
		return v.Interface().(time.Time).IsZero(), true
	}

	// Check all struct fields recursively
	for i := 0; i < v.NumField(); i++ {
		if !IsEmpty(v.Field(i).Interface()) {
			return false, true
		}
	}
	return true, true
}

// IsEmpty checks if the given interface value represents an empty or zero value.
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

	// Check primitive types
	if isEmpty, ok := isEmptyPrimitive(v); ok {
		return isEmpty
	}

	// Check collection types
	if isEmpty, ok := isEmptyCollection(v); ok {
		return isEmpty
	}

	// Check struct types
	if isEmpty, ok := isEmptyStruct(v); ok {
		return isEmpty
	}

	// Default: Compare with zero value
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

// FetchErrorStrings returns a slice of strings containing the error messages
func FetchErrorStrings(errs []error) []string {
	errStrings := make([]string, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	return errStrings
}

// FetchErrorStack returns a string containing the error messages separated by semicolons
func FetchErrorStack(errs []error) string {
	var s strings.Builder
	for _, err := range errs {
		if err != nil {
			s.WriteString(err.Error())
			s.WriteString("; ")
		}

	}
	return s.String()
}

// FetchHTTPStatusCode returns the HTTP status code associated with the response type
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

// IsProdEnvironment returns true if Environment is set to "prod" or "production"
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
func RecoverException(panic interface{}) {
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

// findDynamicPort finds a free port dynamically for the given protocol.
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

// isPortAvailable checks if a TCP or UDP port is available.
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

// **Helper Function: Validate URL**
func ValidateURL(requestURL string) error {
	_, err := url.ParseRequestURI(requestURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

// **Helper Function: Construct URL with Query Params**
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

// **Helper Function: Create Log Directory**
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

// GetIsLogRotationEnabled returns true if log rotation is enabled
func GetIsLogRotationEnabled() bool {
	enableRotation, _ := strconv.ParseBool(os.Getenv(constant.LogRotationEnabled))
	return enableRotation
}

// Println prints a message with the specified log mode and color
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

// Printf prints a formatted message with the specified log mode and color
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

// CorrelationIDFromNatsMsg returns the correlation ID from a nats.Msg.
func CorrelationIDFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.CorrelationIDHeader)
}

// MessageIDFromNatsMsg returns the message ID from a nats.Msg.
func MessageIDFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.MessageIdHeader)
}

// AuthorizationHeaderFromNatsMsg returns the authorization header from a nats.Msg.
func AuthorizationHeaderFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.AuthorizationHeader)
}

// IPHeaderFromNatsMsg returns the ip header from a nats.Msg.
func IPHeaderFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.IPHeader)
}

// ErrorHeadeFromNatsMsg returns the error header from a nats.Msg.
func ErrorHeadeFromNatsMsg(msg *nats.Msg) string {
	return msg.Header.Get(constant.ErrorHeader)
}

// GetIssuerFromConfig returns the issuer from the config
func GetIssuerFromConfig() string {
	return viper.GetString(constant.IssuerKey)
}

// ConstructDiscoveryURL returns the discovery url
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

func Valid() *bool {
	valid := true
	return &valid
}

// MustGetEnv retrieves the environment variable named by the key.
// If the variable is not set or is empty after trimming whitespace,
// the function logs a fatal error and exits the program.
func MustGetEnv(key string) string {
	value := os.Getenv(key)

	if value == "" || strings.TrimSpace(value) == "" {
		// In a real application, you would use a proper logging system (like Zap)
		// and maybe the logging's Fatal method here.
		Printf(constant.FATAL, "FATAL ERROR: Required environment variable '%s' is not set or is empty.\n", key)
	}

	return value
}

// IsURL checks if the given string is a URL
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

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

// NormalizePrecision normalizes the precision of a float64 value.
func NormalizePrecision(val float64, digits int) float64 {
	scale := math.Pow10(digits)
	return math.Round(val*scale) / scale
}

// DetectInputType checks if the string is a file path or raw JSON.
// Returns "file", "json", or "unknown".
// DetectInputType checks if the string is a file path or raw JSON.
// Returns "file", "json", or "".
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

// ParsePhoneNumber parses a raw phone number string into a standardized Info struct.
//
// defaultRegion: A two-letter (ISO 3166-1) country code (e.g., "US", "IN", "GB").
// This is used to guess the country code if the number is not in international
// format (e.g., to understand that "(415) 555-1212" is a "US" number).
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

// GetRegionForCountryCode gets the primary region (e.g., "IN") for a country code (e.g., 91).
// Note: Some codes map to multiple regions (e.g., +1 maps to US, CA, etc.).
func GetRegionForCountryCode(countryCode int) string {
	// This returns the *main* region for that code (e.g., 1 -> "US")
	return phonenumbers.GetRegionCodeForCountryCode(countryCode)
}

// GetCountryInfoForCode gets the primary region and full name for a country code.
// Note: Some codes map to multiple regions (e.g., +1 maps to US, CA, etc.).
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
