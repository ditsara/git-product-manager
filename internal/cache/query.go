package cache

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	sqldialect "github.com/stephenafamo/bob/dialect/sqlite/dialect"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

// CachedTicket is a row from the tickets cache table, used by ListTickets.
type CachedTicket struct {
	ID          string
	Title       string
	Type        string
	Status      string
	HasChildren int
}

// ListOptions controls filtering in ListTickets.
type ListOptions struct {
	// ParentFilter limits results to children of this ticket ID.
	// Combined with Subtree=true it returns all descendants; otherwise direct children only.
	ParentFilter string
	// Subtree, when true with ParentFilter, returns all descendants via materialized-path LIKE.
	// When true without ParentFilter, all tickets are returned (no parent filter at all).
	Subtree bool
	// IncludeStates is a whitelist of statuses. Mutually exclusive with ExcludeStates.
	IncludeStates []string
	// ExcludeStates is a blacklist of statuses. Mutually exclusive with IncludeStates.
	ExcludeStates []string
	// MilestoneFilter limits results to tickets belonging to this milestone ID.
	MilestoneFilter string
	// DependsOn limits results to tickets whose depends_on includes this ticket ID.
	DependsOn string
	// Blocks limits results to tickets that block (are depended upon by) this ticket ID.
	Blocks string
	// Related limits results to tickets with any relationship to this ticket ID.
	Related string
}

// hasChildrenSQL is the computed column expression that detects whether a ticket
// has any direct children. Reused verbatim across all four query paths.
const hasChildrenSQL = "CASE WHEN EXISTS(SELECT 1 FROM tickets AS t WHERE UPPER(t.parent) = UPPER(tickets.id)) THEN 1 ELSE 0 END AS has_children"

// ListTickets queries the SQLite cache and returns tickets matching opts,
// ordered by updated_at descending.
func ListTickets(db *sql.DB, opts ListOptions) ([]CachedTicket, error) {
	ctx := context.Background()

	mods := []bob.Mod[*sqldialect.SelectQuery]{
		sm.Columns(
			sqlite.Quote("id"),
			sqlite.Quote("title"),
			sqlite.Quote("type"),
			sqlite.Quote("status"),
			sqlite.Raw(hasChildrenSQL),
		),
		sm.From("tickets"),
		sm.OrderBy(sqlite.Quote("updated_at")).Desc(),
	}

	// Parent/subtree filtering — four cases:
	//   ParentFilter != "" && Subtree  → all descendants via materialized path LIKE
	//   ParentFilter != "" && !Subtree → direct children only
	//   ParentFilter == "" && !Subtree → top-level tickets only (no parent)
	//   ParentFilter == "" && Subtree  → all tickets (no parent filter)
	if opts.ParentFilter != "" && opts.Subtree {
		parentPath, err := lookupPath(ctx, db, opts.ParentFilter)
		if err != nil {
			return nil, err
		}
		mods = append(mods, sm.Where(sqlite.Quote("path").Like(sqlite.Arg(parentPath+"/%"))))
	} else if opts.ParentFilter != "" {
		// Validate parent exists before querying children
		if _, err := lookupPath(ctx, db, opts.ParentFilter); err != nil {
			return nil, err
		}
		mods = append(mods, sm.Where(
			sqlite.F("UPPER", sqlite.Quote("parent"))().EQ(sqlite.F("UPPER", sqlite.Arg(opts.ParentFilter))()),
		))
	} else if !opts.Subtree {
		mods = append(mods, sm.Where(
			sqlite.Or(
				sqlite.Quote("parent").IsNull(),
				sqlite.Quote("parent").EQ(sqlite.Arg("")),
			),
		))
	}
	// else: Subtree=true, ParentFilter="" → return all tickets without a parent filter

	// Status filtering
	if len(opts.IncludeStates) > 0 {
		mods = append(mods, sm.Where(sqlite.Quote("status").In(stringsToArgs(opts.IncludeStates)...)))
	} else if len(opts.ExcludeStates) > 0 {
		mods = append(mods, sm.Where(sqlite.Quote("status").NotIn(stringsToArgs(opts.ExcludeStates)...)))
	}

	// Milestone filtering — milestones stored as comma-separated string
	if opts.MilestoneFilter != "" {
		pattern := opts.MilestoneFilter
		mods = append(mods, sm.Where(
			sqlite.Or(
				sqlite.Quote("milestones").EQ(sqlite.Arg(pattern)),
				sqlite.Quote("milestones").Like(sqlite.Arg(pattern+",%")),
				sqlite.Quote("milestones").Like(sqlite.Arg("%,"+pattern)),
				sqlite.Quote("milestones").Like(sqlite.Arg("%,"+pattern+",%")),
			),
		))
	}

	// Relationship filtering — each uses an EXISTS subquery against the relationships table.
	// UPPER() normalises ticket IDs so filters are case-insensitive.
	if opts.DependsOn != "" {
		mods = append(mods, sm.Where(sqlite.Raw(
			"EXISTS (SELECT 1 FROM relationships WHERE UPPER(from_ticket) = UPPER(tickets.id) AND UPPER(to_ticket) = UPPER(?) AND relationship_type = 'depends-on')",
			opts.DependsOn,
		)))
	}
	if opts.Blocks != "" {
		// "blocks <id>" = tickets that <id> depends on (tickets blocking <id>)
		mods = append(mods, sm.Where(sqlite.Raw(
			"EXISTS (SELECT 1 FROM relationships WHERE UPPER(to_ticket) = UPPER(tickets.id) AND UPPER(from_ticket) = UPPER(?) AND relationship_type = 'depends-on')",
			opts.Blocks,
		)))
	}
	if opts.Related != "" {
		mods = append(mods, sm.Where(sqlite.Raw(
			"EXISTS (SELECT 1 FROM relationships WHERE (UPPER(from_ticket) = UPPER(tickets.id) AND UPPER(to_ticket) = UPPER(?)) OR (UPPER(to_ticket) = UPPER(tickets.id) AND UPPER(from_ticket) = UPPER(?)))",
			opts.Related, opts.Related,
		)))
	}

	querySQL, queryArgs, err := sqlite.Select(mods...).Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	rows, err := db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list query: %w", err)
	}
	defer rows.Close()

	var tickets []CachedTicket
	for rows.Next() {
		var t CachedTicket
		if err := rows.Scan(&t.ID, &t.Title, &t.Type, &t.Status, &t.HasChildren); err != nil {
			return nil, fmt.Errorf("failed to scan ticket row: %w", err)
		}
		tickets = append(tickets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate ticket rows: %w", err)
	}
	return tickets, nil
}

// lookupPath returns the materialized path for the given ticket ID (case-insensitive).
func lookupPath(ctx context.Context, db *sql.DB, ticketID string) (string, error) {
	querySQL, queryArgs, err := sqlite.Select(
		sm.Columns(sqlite.Quote("path")),
		sm.From("tickets"),
		sm.Where(sqlite.F("UPPER", sqlite.Quote("id"))().EQ(sqlite.F("UPPER", sqlite.Arg(ticketID))())),
	).Build(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to build parent path query: %w", err)
	}

	var path string
	if err := db.QueryRowContext(ctx, querySQL, queryArgs...).Scan(&path); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("ticket %q not found", ticketID)
		}
		return "", fmt.Errorf("failed to look up path for %q: %w", ticketID, err)
	}
	return path, nil
}

// stringsToArgs converts a string slice into a slice of Bob Arg expressions.
func stringsToArgs(ss []string) []bob.Expression {
	exprs := make([]bob.Expression, len(ss))
	for i, s := range ss {
		exprs[i] = sqlite.Arg(s)
	}
	return exprs
}

// SearchResult is a ticket row returned by SearchTickets, including a body snippet.
type SearchResult struct {
	ID      string
	Title   string
	Type    string
	Status  string
	Snippet string // short excerpt from body around the first match; empty for id/title matches
}

// SearchOptions controls filtering for SearchTickets.
type SearchOptions struct {
	// IncludeStates is a whitelist of statuses. Mutually exclusive with ExcludeStates.
	IncludeStates []string
	// ExcludeStates is a blacklist of statuses. Mutually exclusive with IncludeStates.
	ExcludeStates []string
}

// SearchTickets searches the tickets cache for query across id, title, and body.
// Results are ordered by relevance: id match first, then title, then body.
func SearchTickets(db *sql.DB, query string, opts SearchOptions) ([]SearchResult, error) {
	ctx := context.Background()

	pattern := "%" + query + "%"

	mods := []bob.Mod[*sqldialect.SelectQuery]{
		sm.Columns(
			sqlite.Quote("id"),
			sqlite.Quote("title"),
			sqlite.Quote("type"),
			sqlite.Quote("status"),
			sqlite.Quote("body"),
		),
		sm.From("tickets"),
		sm.Where(sqlite.Or(
			sqlite.Quote("id").Like(sqlite.Arg(pattern)),
			sqlite.Quote("title").Like(sqlite.Arg(pattern)),
			sqlite.Quote("body").Like(sqlite.Arg(pattern)),
		)),
		sm.OrderBy(sqlite.Raw(
			"CASE WHEN UPPER(id) LIKE UPPER(?) THEN 0 WHEN UPPER(title) LIKE UPPER(?) THEN 1 ELSE 2 END",
			pattern, pattern,
		)),
		sm.OrderBy(sqlite.Quote("updated_at")).Desc(),
	}

	if len(opts.IncludeStates) > 0 {
		mods = append(mods, sm.Where(sqlite.Quote("status").In(stringsToArgs(opts.IncludeStates)...)))
	} else if len(opts.ExcludeStates) > 0 {
		mods = append(mods, sm.Where(sqlite.Quote("status").NotIn(stringsToArgs(opts.ExcludeStates)...)))
	}

	querySQL, queryArgs, err := sqlite.Select(mods...).Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build search query: %w", err)
	}

	rows, err := db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var body sql.NullString
		if err := rows.Scan(&r.ID, &r.Title, &r.Type, &r.Status, &body); err != nil {
			return nil, fmt.Errorf("failed to scan search row: %w", err)
		}
		if body.Valid {
			r.Snippet = extractSnippet(body.String, query, 60)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate search rows: %w", err)
	}
	return results, nil
}

// extractSnippet finds the first case-insensitive occurrence of query in text and
// returns a short excerpt of contextWidth characters on each side, wrapped in "...".
func extractSnippet(text, query string, contextWidth int) string {
	lower := strings.ToLower(text)
	lowerQ := strings.ToLower(query)
	idx := strings.Index(lower, lowerQ)
	if idx < 0 {
		return ""
	}
	start := idx - contextWidth
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + contextWidth
	if end > len(text) {
		end = len(text)
	}
	excerpt := strings.TrimSpace(text[start:end])
	// Replace newlines with spaces for single-line display
	excerpt = strings.ReplaceAll(excerpt, "\n", " ")
	excerpt = strings.ReplaceAll(excerpt, "\r", "")
	prefix, suffix := "…", "…"
	if start == 0 {
		prefix = ""
	}
	if end == len(text) {
		suffix = ""
	}
	return prefix + excerpt + suffix
}
