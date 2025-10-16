package env

import (
	"os"
	"strings"
)

// Prefix returns the environment-specific prefix for keys/subjects/channels.
// Mapping:
//
//	DEVELOPMENT -> dev_
//	ALPHA       -> alpha_
//	PRODUCTION  -> prod_
//
// Unknown or empty defaults to dev_.
func Prefix() string {
	val := strings.ToUpper(strings.TrimSpace(os.Getenv("CYTONIC_ENVIRONMENT")))
	switch val {
	case "DEVELOPMENT":
		return "dev_"
	case "ALPHA":
		return "alpha_"
	case "PRODUCTION":
		return "prod_"
	default:
		return "dev_"
	}
}

// EnsurePrefixed prepends the environment prefix if the provided s does not
// already start with one of the known prefixes. This prevents double-prefixing
// in case callers inadvertently pass already-prefixed values.
func EnsurePrefixed(s string) string {
	if s == "" {
		return s
	}
	if hasKnownPrefix(s) {
		return s
	}
	return Prefix() + s
}

func hasKnownPrefix(s string) bool {
	if strings.HasPrefix(s, "dev_") || strings.HasPrefix(s, "alpha_") || strings.HasPrefix(s, "prod_") {
		return true
	}
	return false
}
