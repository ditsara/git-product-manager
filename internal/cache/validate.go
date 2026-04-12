package cache

import "fmt"

// validateRelationshipSymmetry checks that relationship pairs in the ticket set
// are properly mirrored:
//   - depends-on ↔ blocks: if A depends-on B, then B must block A (and vice versa)
//   - related ↔ related:   if A related B, then B must relate back to A
//
// Returns a list of human-readable warning strings, each with an indented
// `pm link` fix command on the second line.
func validateRelationshipSymmetry(relationships []relationshipData) []string {
	type edge struct{ from, to, relType string }

	edgeSet := make(map[edge]bool, len(relationships))
	for _, r := range relationships {
		edgeSet[edge{r.fromTicket, r.toTicket, r.relType}] = true
	}

	var warnings []string
	seenRelated := make(map[string]bool)

	for _, r := range relationships {
		switch r.relType {
		case "depends-on":
			if !edgeSet[edge{r.toTicket, r.fromTicket, "blocks"}] {
				warnings = append(warnings, fmt.Sprintf(
					"%s depends_on %s, but %s does not block %s; to fix:\n   pm link %s %s --type blocks",
					r.fromTicket, r.toTicket, r.toTicket, r.fromTicket,
					r.toTicket, r.fromTicket,
				))
			}

		case "blocks":
			if !edgeSet[edge{r.toTicket, r.fromTicket, "depends-on"}] {
				warnings = append(warnings, fmt.Sprintf(
					"%s blocks %s, but %s does not depends_on %s; to fix:\n   pm link %s %s --type depends-on",
					r.fromTicket, r.toTicket, r.toTicket, r.fromTicket,
					r.toTicket, r.fromTicket,
				))
			}

		case "related":
			// Build a canonical key so each missing pair is reported only once.
			key := r.fromTicket + ":" + r.toTicket
			reverseKey := r.toTicket + ":" + r.fromTicket
			if seenRelated[key] || seenRelated[reverseKey] {
				continue
			}
			seenRelated[key] = true
			if !edgeSet[edge{r.toTicket, r.fromTicket, "related"}] {
				warnings = append(warnings, fmt.Sprintf(
					"%s related %s, but %s does not relate back to %s; to fix:\n   pm link %s %s --type related",
					r.fromTicket, r.toTicket, r.toTicket, r.fromTicket,
					r.toTicket, r.fromTicket,
				))
			}
		}
	}

	return warnings
}
