package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/cache"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// TreeNode represents a ticket in the tree with its children.
type TreeNode struct {
	ID       string
	Title    string
	Type     string
	Children []*TreeNode
	Depth    int
}

var treeFlags struct {
	depth int
}

var treeCmd = &cobra.Command{
	Use:               "tree <id> [--depth N]",
	Short:             "Display ticket hierarchy as ASCII tree",
	Long:              `Visualize parent-child relationships recursively using ASCII box-drawing characters.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTicketIDs,
	Example:           "  pm tree GPM-44\n  pm tree GPM-1 --depth 2",
	Run: func(cmd *cobra.Command, args []string) {
		rootID := args[0]
		pmPath := ".pm"

		// Validate depth parameter
		if treeFlags.depth < 0 {
			log.Fatalf("Error: Invalid depth '%d'. Must be a positive integer.", treeFlags.depth)
		}

		// Ensure cache database exists and has current schema
		if err := cache.EnsureCacheReady(pmPath); err != nil {
			log.Fatalf("Error initializing cache: %v", err)
		}

		// Check if cache needs sync and sync if necessary
		shouldSync, err := cache.ShouldSync(pmPath)
		if err != nil {
			log.Printf("Warning: failed to check cache staleness: %v", err)
			log.Println("Continuing with potentially stale cache...")
		} else if shouldSync {
			if err := cache.SyncCache(pmPath); err != nil {
				log.Fatalf("Error syncing cache: %v", err)
			}
		}

		// Open database
		dbPath := filepath.Join(pmPath, ".cache.db")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer db.Close()

		// Verify root ticket exists by querying the cache
		rootTicket, err := getTicketFromCache(db, rootID)
		if err != nil {
			log.Fatalf("Error: Ticket '%s' not found.", rootID)
		}

		// Build tree recursively
		root := &TreeNode{
			ID:    rootTicket.ID,
			Title: rootTicket.Title,
			Type:  rootTicket.Type,
			Depth: 0,
		}

		if treeFlags.depth == 0 {
			// depth=0 means unlimited
			buildTree(db, root, -1)
		} else {
			// --depth N: show N levels (root is level 1, children are level 2, etc.)
			// With root at depth 0, children are at depth 1
			// So we fetch children when depth < N, i.e., maxDepth = N
			// But we check at START of buildTree, so for --depth 1:
			// - maxDepth = 1, root.Depth=0, 0 >= 1? No, fetch
			// - child.Depth=1, 1 >= 1? Yes, stop
			// This would fetch children even with --depth 1!
			// Solution: use --depth 1 → maxDepth = 0 to prevent ANY child fetching
			buildTree(db, root, treeFlags.depth-1)
		}

		// Render the tree
		renderTree(root)
	},
}

// getTicketFromCache retrieves a single ticket from the cache by ID.
func getTicketFromCache(db *sql.DB, ticketID string) (*cache.CachedTicket, error) {
	tickets, err := cache.ListTickets(db, cache.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Case-insensitive search
	searchID := strings.ToUpper(ticketID)
	for i := range tickets {
		if strings.ToUpper(tickets[i].ID) == searchID {
			return &tickets[i], nil
		}
	}

	return nil, fmt.Errorf("ticket not found")
}

// buildTree recursively fetches and builds the tree structure.
// If maxDepth is -1, depth is unlimited. Otherwise, only fetch children when node.Depth < maxDepth.
func buildTree(db *sql.DB, node *TreeNode, maxDepth int) error {
	// Check if we should fetch children
	if maxDepth >= 0 && node.Depth >= maxDepth {
		// Don't fetch children for this node
		return nil
	}

	// Query children of this ticket
	children, err := cache.ListTickets(db, cache.ListOptions{
		ParentFilter: node.ID,
	})
	if err != nil {
		return err
	}

	// Create tree nodes for each child
	for _, child := range children {
		childNode := &TreeNode{
			ID:    child.ID,
			Title: child.Title,
			Type:  child.Type,
			Depth: node.Depth + 1,
		}

		// Recursively build subtree
		if err := buildTree(db, childNode, maxDepth); err != nil {
			return err
		}

		node.Children = append(node.Children, childNode)
	}

	return nil
}

// renderTree renders the tree starting from a root node.
func renderTree(root *TreeNode) {
	fmt.Println(formatNode(root))
	renderChildren(root.Children, "")
}

// renderChildren recursively renders child nodes with proper indentation and box-drawing characters.
func renderChildren(children []*TreeNode, prefix string) {
	for i, child := range children {
		isLast := i == len(children)-1

		// Choose box-drawing character
		var connector string
		var childPrefix string
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		} else {
			connector = "├── "
			childPrefix = prefix + "│   "
		}

		// Print this child
		fmt.Println(prefix + connector + formatNode(child))

		// Recursively print grandchildren
		if len(child.Children) > 0 {
			renderChildren(child.Children, childPrefix)
		}
	}
}

// formatNode formats a single node for display.
// Truncates title to 60 chars if necessary.
func formatNode(node *TreeNode) string {
	title := node.Title
	if len([]rune(title)) > 60 {
		runes := []rune(title)
		title = string(runes[:57]) + "..."
	}

	return fmt.Sprintf("%s: %s", node.ID, title)
}

func init() {
	treeCmd.Flags().IntVarP(&treeFlags.depth, "depth", "d", 0, "Maximum depth to display (0 = unlimited)")
	rootCmd.AddCommand(treeCmd)
}
