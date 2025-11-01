package types

import (
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

// ToPgTypeInt32 converts an int value to pgtype.Int4 (PostgreSQL INTEGER).
func ToPgTypeInt32(value int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: value,
		Valid: true,
	}
}

// ToPgTypeInt2 converts an int value to pgtype.Int2 (PostgreSQL INTEGER).
func ToPgTypeInt2(value int16) pgtype.Int2 {
	return pgtype.Int2{
		Int16: value,
		Valid: true}
}

// ToPgTypeInt8 converts an int value to pgtype.Int2 (PostgreSQL INTEGER).
func ToPgTypeInt8(value int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: value,
		Valid: true}
}

// ToPgTypeText converts a string to pgtype.Text.
func ToPgTypeText(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

// ToPgTypeBool converts a bool to pgtype.Bool.
func ToPgTypeBool(value bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  value,
		Valid: true,
	}
}

// ToPgTypeFloat8 converts a float64 to pgtype.Float8.
func ToPgTypeFloat8(value float64) pgtype.Float8 {
	return pgtype.Float8{
		Float64: value,
		Valid:   true,
	}
}

// ToPgTypeTimestamptz converts a time.Time to pgtype.Timestamptz.
func ToPgTypeTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

// ToPgTypeTimestamp converts a time.Time to pgtype.Timestamp.
func ToPgTypeTimestamp(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  value,
		Valid: true,
	}
}

// ToUUID converts a pgtype.UUID to uuid.UUID.
func ToUUID(p pgtype.UUID) (uuid.UUID, error) {
	if !p.Valid {
		return uuid.Nil, fmt.Errorf("uuid is null")
	}
	return uuid.UUID(p.Bytes), nil
}

// ToPgTypeUUID converts a uuid.UUID to pgtype.UUID.
func ToPgTypeUUID(u uuid.UUID) pgtype.UUID {
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
		Places:       3,
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
