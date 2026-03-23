package milestone

import (
	"fmt"
	"regexp"
	"strings"
)

var milestoneIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("milestone ID cannot be empty")
	}
	if !milestoneIDPattern.MatchString(id) {
		return fmt.Errorf("invalid milestone ID %q: must start with a lowercase letter and contain only lowercase letters, digits, and hyphens", id)
	}
	return nil
}

// SlugFromTitle converts a title to a valid milestone ID.
func SlugFromTitle(title string) string {
	slug := strings.ToLower(title)

	// Replace any non-[a-z0-9] character with a hyphen
	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	slug = b.String()

	// Collapse consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "milestone"
	}
	return slug
}
