package validate

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

func ResourceID(id string) error {
	if id == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}
	for _, r := range id {
		if r < 0x20 {
			return fmt.Errorf("resource ID contains control character at position %d", r)
		}
	}
	cleaned := filepath.Clean(id)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("resource ID contains path traversal: %q", id)
	}
	if strings.ContainsAny(id, "?#") {
		return fmt.Errorf("resource ID contains query characters: %q", id)
	}
	if strings.Contains(id, "%") {
		return fmt.Errorf("resource ID contains percent-encoding: %q (do not pre-encode)", id)
	}
	return nil
}

func SafeString(s string, maxLen int) error {
	if len([]rune(s)) > maxLen {
		return fmt.Errorf("string exceeds maximum length of %d characters", maxLen)
	}
	for i, r := range s {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("string contains control character at position %d", i)
		}
	}
	return nil
}

func ObjectKey(key string) error {
	if key == "" {
		return fmt.Errorf("object key cannot be empty")
	}
	if len([]byte(key)) > 1024 {
		return fmt.Errorf("object key exceeds maximum length of 1024 bytes")
	}
	for i, r := range key {
		if r < 0x20 && r != '\t' {
			return fmt.Errorf("object key contains control character at position %d", i)
		}
	}
	return nil
}

func BucketName(name string) error {
	if name == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '.' {
			return fmt.Errorf("bucket name contains invalid character: %q", string(r))
		}
	}
	return nil
}

func JSONPayload(s string) error {
	for i, r := range s {
		if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("JSON payload contains non-printable character at position %d", i)
		}
	}
	return nil
}
