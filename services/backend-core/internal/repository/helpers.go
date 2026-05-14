package repository

<<<<<<< HEAD
import "encoding/json"
=======
import (
	"encoding/json"
	"fmt"
	"regexp"
)
>>>>>>> phase1/security-stability

func marshalJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func unmarshalJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}
<<<<<<< HEAD
=======

// allowedColumnPattern matches valid SQL column names: lowercase letters, digits, underscores.
// This is a defense-in-depth measure for dynamic SQL construction.
var allowedColumnPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// validateSQLColumns ensures all column names match the allowed pattern.
// Use this as a guard before building dynamic SQL with column names.
func validateSQLColumns(columns ...string) error {
	for _, col := range columns {
		if !allowedColumnPattern.MatchString(col) {
			return fmt.Errorf("invalid column name %q", col)
		}
	}
	return nil
}

// joinUpdates joins SET clause fragments with ", ".
// SECURITY: all fragments must use hardcoded column names only -- never user input.
func joinUpdates(updates []string) string {
	result := ""
	for i, u := range updates {
		if i > 0 {
			result += ", "
		}
		result += u
	}
	return result
}

// joinStrings joins strings with a separator.
// SECURITY: all elements must use hardcoded values only -- never user input.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
>>>>>>> phase1/security-stability
