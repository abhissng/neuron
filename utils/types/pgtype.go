package types

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToPgTypeInt4 converts an int to pgtype.Int4.
func ToPgTypeInt4(value int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: value,
		Valid: true,
	}
}

// PgTypeInt4ToInt32 converts pgtype.Int4 to int32. Returns 0 when Valid is false.
func PgTypeInt4ToInt32(p pgtype.Int4) int32 {
	if !p.Valid {
		return 0
	}
	return p.Int32
}

// ToPgTypeInt2 converts an int value to pgtype.Int2 (PostgreSQL INTEGER).
func ToPgTypeInt2(value int16) pgtype.Int2 {
	return pgtype.Int2{
		Int16: value,
		Valid: true}
}

// PgTypeInt2ToInt16 converts pgtype.Int2 to int16. Returns 0 when Valid is false.
func PgTypeInt2ToInt16(p pgtype.Int2) int16 {
	if !p.Valid {
		return 0
	}
	return p.Int16
}

// ToPgTypeInt8 converts an int value to pgtype.Int2 (PostgreSQL INTEGER).
func ToPgTypeInt8(value int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: value,
		Valid: true}
}

// PgTypeInt8ToInt64 converts pgtype.Int8 to int64. Returns 0 when Valid is false.
func PgTypeInt8ToInt64(p pgtype.Int8) int64 {
	if !p.Valid {
		return 0
	}
	return p.Int64
}

// ToPgTypeText converts the provided string into a pgtype.Text for PostgreSQL.
// It returns a Text with Valid set to false when the input is empty or contains only whitespace; otherwise it sets String to the input and Valid to true.
func ToPgTypeText(value string) pgtype.Text {
	if strings.TrimSpace(value) == "" {
		return pgtype.Text{
			Valid: false,
		}
	}
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

// PgTypeTextToString converts pgtype.Text to string. Returns "" when Valid is false.
func PgTypeTextToString(p pgtype.Text) string {
	if !p.Valid {
		return ""
	}
	return p.String
}

// ToPgTypeBool converts a bool to pgtype.Bool.
func ToPgTypeBool(value bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  value,
		Valid: true,
	}
}

// PgTypeBoolToBool converts pgtype.Bool to bool. Returns false when Valid is false.
func PgTypeBoolToBool(p pgtype.Bool) bool {
	if !p.Valid {
		return false
	}
	return p.Bool
}

// ToPgTypeFloat8 converts a float64 to pgtype.Float8.
func ToPgTypeFloat8(value float64) pgtype.Float8 {
	return pgtype.Float8{
		Float64: value,
		Valid:   true,
	}
}

// PgTypeFloat8ToFloat64 converts pgtype.Float8 to float64. Returns 0 when Valid is false.
func PgTypeFloat8ToFloat64(p pgtype.Float8) float64 {
	if !p.Valid {
		return 0
	}
	return p.Float64
}

// ToPgTypeTimestamptz converts a time.Time to a pgtype.Timestamptz, marking the result invalid if the input is the zero time.
func ToPgTypeTimestamptz(value time.Time) pgtype.Timestamptz {
	if value.IsZero() {
		return pgtype.Timestamptz{
			Valid: false,
		}
	}
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

// PgTypeTimestamptzToTime converts pgtype.Timestamptz to time.Time. Returns zero time when Valid is false.
func PgTypeTimestamptzToTime(p pgtype.Timestamptz) time.Time {
	if !p.Valid {
		return time.Time{}
	}
	return p.Time
}

// ToPgTypeTimestamp converts the provided time.Time into a pgtype.Timestamp and marks the result invalid when the input is the zero time value.
func ToPgTypeTimestamp(value time.Time) pgtype.Timestamp {
	if value.IsZero() || value.Equal(time.Time{}) {
		return pgtype.Timestamp{
			Valid: false,
		}
	}
	return pgtype.Timestamp{
		Time:  value,
		Valid: true,
	}
}

// PgTypeTimestampToTime converts pgtype.Timestamp to time.Time. Returns (zero time, false) when Valid is false.
func PgTypeTimestampToTime(p pgtype.Timestamp) time.Time {
	if !p.Valid {
		return time.Time{}
	}
	return p.Time
}

// ToUUID converts a pgtype.UUID to uuid.UUID.
func ToUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return uuid.UUID(p.Bytes)
}

// ToPgTypeUUID converts a uuid.UUID to a pgtype.UUID, treating uuid.Nil as a NULL value.
// If u is uuid.Nil the returned pgtype.UUID has Valid set to false; otherwise Bytes is set to u and Valid is true.
func ToPgTypeUUID(u uuid.UUID) pgtype.UUID {
	if u == uuid.Nil {
		return pgtype.UUID{
			Valid: false,
		}
	}
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// DecimalConfig defines configuration for FloatToPgNumeric
type DecimalConfig struct {
	Places       int  // default: 3
	TrimIntegers bool // omit decimals for whole numbers
	MinPlaces    int  // lower bound on precision
	MaxPlaces    int  // upper bound on precision
}

// DecimalOpt modifies DecimalConfig
type DecimalOpt func(*DecimalConfig)

// Prec sets the number of decimal places
func Prec(n int) DecimalOpt {
	return func(c *DecimalConfig) { c.Places = n }
}

// SmartTrim enables whole-number trimming
func SmartTrim() DecimalOpt {
	return func(c *DecimalConfig) { c.TrimIntegers = true }
}

// Limit sets the minimum and maximum precision bounds
func Limit(min, max int) DecimalOpt {
	return func(c *DecimalConfig) {
		c.MinPlaces = min
		c.MaxPlaces = max
	}
}

// FloatToPgNumeric converts a float64 into pgtype.Numeric using configurable precision rules
func FloatToPgNumeric(val float64, opts ...DecimalOpt) pgtype.Numeric {
	// defaults
	cfg := &DecimalConfig{
		Places:       2,
		TrimIntegers: false,
		MinPlaces:    0,
		MaxPlaces:    10,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	prec := cfg.Places

	// Handle whole numbers if SmartTrim is active
	if cfg.TrimIntegers && val == math.Trunc(val) {
		prec = 0
	}

	// Clamp within allowed bounds
	if prec < cfg.MinPlaces {
		prec = cfg.MinPlaces
	} else if prec > cfg.MaxPlaces {
		prec = cfg.MaxPlaces
	}

	// Format and scan into pgtype.Numeric
	formatted := strconv.FormatFloat(adjustPrecision(val, prec), 'f', prec, 64)
	var num pgtype.Numeric
	_ = num.Scan(formatted)
	return num
}

// FloatToPgWhole ensures integers remain cleanly formatted
func FloatToPgWhole(val float64) pgtype.Numeric {
	return FloatToPgNumeric(val, SmartTrim())
}

// FloatToPgFixed converts with exact decimal precision
func FloatToPgFixed(val float64, places int) pgtype.Numeric {
	return FloatToPgNumeric(val, Prec(places))
}

// FloatToPgClean trims decimals for integers and enforces precision
func FloatToPgClean(val float64, places int) pgtype.Numeric {
	return FloatToPgNumeric(val, SmartTrim(), Prec(places))
}

// PgTypeNumericToFloat converts pgtype.Numeric to float64. Returns (0, false) when Valid is false, or 0 when AssignTo fails.
func PgTypeNumericToFloat(p pgtype.Numeric) float64 {
	if !p.Valid {
		return 0
	}

	if f, err := p.Float64Value(); err == nil {
		return f.Float64
	}
	return 0.0
}

// adjustPrecision rounds to a given decimal count
func adjustPrecision(f float64, decimals int) float64 {
	scale := math.Pow10(decimals)
	return math.Round(f*scale) / scale
}

// ParseInterval converts a human-readable duration string into pgtype.Interval.
// Supports formats like "30s", "45 sec", "2 minutes", etc.
func ParseToPgTypeInterval(input string) pgtype.Interval {
	var iv pgtype.Interval

	txt := strings.TrimSpace(strings.ToLower(input))
	if txt == "" {
		return iv
	}

	// Attempt simple time unit match (seconds/minutes)
	pattern := regexp.MustCompile(`^(\d+)\s*(sec(?:onds?)?|s|min(?:utes?)?|m)$`)
	m := pattern.FindStringSubmatch(txt)
	if len(m) == 3 {
		num, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return iv
		}

		unit := m[2]
		micros := int64(0)
		switch unit {
		case "s", "sec", "second", "seconds":
			micros = num * 1_000_000
		case "m", "min", "minute", "minutes":
			micros = num * 60 * 1_000_000
		}

		if micros > 0 {
			iv.Microseconds = micros
			iv.Valid = true
			return iv
		}
	}

	// fallback — let pgtype handle complex or ISO8601-like formats
	if err := iv.Scan(txt); err != nil {
		return iv
	}
	return iv
}

// PgTypeIntervalToDuration converts pgtype.Interval to time.Duration using Microseconds (Days and Months are not represented in time.Duration).
// Returns 0 when Valid is false.
func PgTypeIntervalToDuration(p pgtype.Interval) time.Duration {
	if !p.Valid {
		return 0
	}
	// time.Duration is in nanoseconds; Interval.Microseconds is in microseconds
	d := time.Duration(p.Microseconds) * time.Microsecond
	d += time.Duration(p.Days) * 24 * time.Hour
	// Approximate: 1 month ≈ 30 days (no exact mapping to time.Duration)
	d += time.Duration(p.Months) * 30 * 24 * time.Hour
	return d
}

// ToPgTypeDate converts t to a pgtype.Date; if t is the zero time the returned Date has Valid set to false.
func ToPgTypeDate(t time.Time) pgtype.Date {
	if t.IsZero() || t.Equal(time.Time{}) {
		return pgtype.Date{
			Valid: false,
		}
	}
	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

// PgTypeDateToTime converts pgtype.Date to time.Time. Returns zero time when Valid is false.
func PgTypeDateToTime(p pgtype.Date) time.Time {
	if !p.Valid {
		return time.Time{}
	}
	return p.Time
}

// ToPgTime converts t to a pgtype.Time representing the number of microseconds
// since midnight and marks it invalid if t is the zero time.
//
// When t is not the zero value, Microseconds is set to the total microseconds
// elapsed since midnight (hours, minutes, seconds, and nanoseconds converted)
// and Valid is true. When t is zero, Valid is false.
func ToPgTime(t time.Time) pgtype.Time {
	if t.IsZero() || t.Equal(time.Time{}) {
		return pgtype.Time{
			Valid: false,
		}
	}
	return pgtype.Time{
		Microseconds: int64(t.Hour()*3600*1_000_000 +
			t.Minute()*60*1_000_000 +
			t.Second()*1_000_000 +
			t.Nanosecond()/1000),
		Valid: true,
	}
}

// PgTypeTimeToTime converts pgtype.Time to time.Time (time-of-day since midnight in UTC).
// Returns zero time when Valid is false.
func PgTypeTimeToTime(p pgtype.Time) time.Time {
	if !p.Valid {
		return time.Time{}
	}
	sec := p.Microseconds / 1_000_000
	nsec := (p.Microseconds % 1_000_000) * 1000
	return time.Date(0, 1, 1, int(sec/3600), int((sec%3600)/60), int(sec%60), int(nsec), time.UTC)
}

// ToPgType converts a Go value into a specific pgtype, constrained by PgType.
func ToPgType[T PgType](value any, opts ...DecimalOpt) (T, error) {
	var zero T

	switch any(zero).(type) {
	case pgtype.Int2:
		v, ok := value.(int16)
		if !ok {
			return zero, fmt.Errorf("expected int16, got %T", value)
		}
		return any(ToPgTypeInt2(v)).(T), nil

	case pgtype.Int4:
		v, ok := value.(int32)
		if !ok {
			return zero, fmt.Errorf("expected int32, got %T", value)
		}
		return any(ToPgTypeInt4(v)).(T), nil

	case pgtype.Int8:
		switch val := value.(type) {
		case int64:
			return any(ToPgTypeInt8(val)).(T), nil
		case int:
			return any(ToPgTypeInt8(int64(val))).(T), nil
		default:
			return zero, fmt.Errorf("expected int/int64, got %T", value)
		}

	case pgtype.Text:
		v, ok := value.(string)
		if !ok {
			return zero, fmt.Errorf("expected string, got %T", value)
		}
		return any(ToPgTypeText(v)).(T), nil

	case pgtype.Bool:
		v, ok := value.(bool)
		if !ok {
			return zero, fmt.Errorf("expected bool, got %T", value)
		}
		return any(ToPgTypeBool(v)).(T), nil

	case pgtype.Float8:
		v, ok := value.(float64)
		if !ok {
			return zero, fmt.Errorf("expected float64, got %T", value)
		}
		return any(ToPgTypeFloat8(v)).(T), nil

	case pgtype.Timestamp:
		v, ok := value.(time.Time)
		if !ok {
			return zero, fmt.Errorf("expected time.Time, got %T", value)
		}
		return any(ToPgTypeTimestamp(v)).(T), nil

	case pgtype.Timestamptz:
		v, ok := value.(time.Time)
		if !ok {
			return zero, fmt.Errorf("expected time.Time, got %T", value)
		}
		return any(ToPgTypeTimestamptz(v)).(T), nil

	case pgtype.UUID:
		switch val := value.(type) {
		case uuid.UUID:
			return any(ToPgTypeUUID(val)).(T), nil
		case string:
			id, err := uuid.Parse(val)
			if err != nil {
				return zero, fmt.Errorf("invalid UUID string: %v", err)
			}
			return any(ToPgTypeUUID(id)).(T), nil
		default:
			return zero, fmt.Errorf("expected uuid.UUID or string, got %T", value)
		}

	case pgtype.Date:
		v, ok := value.(time.Time)
		if !ok {
			return zero, fmt.Errorf("expected time.Time, got %T", value)
		}
		return any(ToPgTypeDate(v)).(T), nil

	case pgtype.Interval:
		v, ok := value.(string)
		if !ok {
			return zero, fmt.Errorf("expected duration string, got %T", value)
		}
		interval := ParseToPgTypeInterval(v)
		if !interval.Valid {
			return zero, fmt.Errorf("invalid interval string %q", v)
		}
		return any(interval).(T), nil
	case pgtype.Numeric:
		v, ok := value.(float64)
		if !ok {
			return zero, fmt.Errorf("expected float64, got %T", value)
		}
		return any(FloatToPgNumeric(v, opts...)).(T), nil

	default:
		return zero, fmt.Errorf("unsupported pgtype: %T", zero)
	}
}

// ToPgTypeNil returns the zero value of the specified pgtype T, which can be used to represent a NULL/invalid value for that PostgreSQL type.
func ToPgTypeNil[T PgType]() T {
	var zero T
	return zero
}

// ToJSONBAny marshals v to JSON and returns the resulting bytes suitable for JSONB storage.
// If v is nil or marshaling fails, it returns nil.
func ToJSONBAny(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
