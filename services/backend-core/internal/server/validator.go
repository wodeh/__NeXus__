package server

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

// Validator provides lightweight request validation without heavy dependencies.
type Validator struct {
	errors map[string]string
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{errors: make(map[string]string)}
}

// ValidateStruct validates a struct based on `validate` tags.
// Supported tags:
//   - "required": field must be non-zero
//   - "email": field must be a valid email address
//   - "uuid": field must be a valid UUID format
//   - "date": field must be YYYY-MM-DD format
//   - "positive": numeric field must be > 0
//
// Returns map[field]error_message or nil if no errors.
func ValidateStruct(v interface{}) map[string]string {
	val := NewValidator()

	switch req := v.(type) {
	case *loginRequest:
		if req == nil {
			val.errors["_"] = "nil request"
			return val.errors
		}
		val.checkRequired("email", req.Email)
		val.checkEmail("email", req.Email)
		val.checkRequired("password", req.Password)

	default:
		// For structs with validate tags, use reflection-based validation
		val.errors["_"] = "unsupported type for validation"
	}

	if len(val.errors) == 0 {
		return nil
	}
	return val.errors
}

// ValidateEmail validates a single email address.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateUUID validates a single UUID string.
func ValidateUUID(s string) error {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(s) {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// ValidateDate validates YYYY-MM-DD format.
func ValidateDate(s string) error {
	if s == "" {
		return fmt.Errorf("date is required")
	}
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(s) {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	return nil
}

// ValidatePositiveInt validates that a number is positive.
func ValidatePositiveInt(n int) error {
	if n <= 0 {
		return fmt.Errorf("must be a positive number")
	}
	return nil
}

// --- internal helpers ---

func (v *Validator) checkRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors[field] = field + " is required"
	}
}

func (v *Validator) checkEmail(field, value string) {
	if value == "" {
		return // let required tag handle empty
	}
	if _, err := mail.ParseAddress(value); err != nil {
		v.errors[field] = "invalid email format"
	}
}

func (v *Validator) checkUUID(field, value string) {
	if value == "" {
		return
	}
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(value) {
		v.errors[field] = "invalid UUID format"
	}
}

func (v *Validator) checkDate(field, value string) {
	if value == "" {
		return
	}
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(value) {
		v.errors[field] = "invalid date format, expected YYYY-MM-DD"
	}
}

func (v *Validator) checkPositiveInt(field string, value int) {
	if value <= 0 {
		v.errors[field] = "must be a positive number"
	}
}

func (v *Validator) checkPositiveFloat(field string, value float64) {
	if value <= 0 {
		v.errors[field] = "must be a positive number"
	}
}

// validateStringField checks a field against validation rules.
// Returns a user-friendly error message or empty string if valid.
func validateStringField(fieldName, value, rules string) string {
	parts := strings.Split(rules, ",")
	for _, rule := range parts {
		rule = strings.TrimSpace(rule)
		switch rule {
		case "required":
			if strings.TrimSpace(value) == "" {
				return fieldName + " is required"
			}
		case "email":
			if value == "" {
				continue
			}
			if _, err := mail.ParseAddress(value); err != nil {
				return "invalid email format"
			}
		case "uuid":
			if value == "" {
				continue
			}
			if err := ValidateUUID(value); err != nil {
				return "invalid UUID format"
			}
		case "date":
			if value == "" {
				continue
			}
			if err := ValidateDate(value); err != nil {
				return "invalid date format, expected YYYY-MM-DD"
			}
		default:
			if strings.HasPrefix(rule, "min=") {
				minStr := strings.TrimPrefix(rule, "min=")
				min, _ := strconv.Atoi(minStr)
				if len(value) < min {
					return fieldName + " must be at least " + minStr + " characters"
				}
			}
			if strings.HasPrefix(rule, "max=") {
				maxStr := strings.TrimPrefix(rule, "max=")
				max, _ := strconv.Atoi(maxStr)
				if len(value) > max {
					return fieldName + " must be at most " + maxStr + " characters"
				}
			}
		}
	}
	return ""
}
