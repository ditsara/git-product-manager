#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Git Product Manager - Comprehensive Sandbox${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"

# Build the binary
echo ""
echo -e "${YELLOW}Building pm binary...${NC}"
./scripts/build.sh > /dev/null 2>&1

# Create a clean test environment
echo ""
echo -e "${YELLOW}Setting up sandbox environment...${NC}"
rm -rf sandbox
mkdir -p sandbox
cd sandbox

# Initialize pm
echo -e "${YELLOW}Initializing .pm directory with SANDBOX prefix...${NC}"
../bin/pm init . --prefix SANDBOX > /dev/null

echo -e "${GREEN}✓ Sandbox initialized${NC}"

# Counters to track ticket IDs
EPIC_ID=""
STORY_IDS=()
TASK_IDS=()
BUG_IDS=()

# Helper function to create ticket and return ID
create_ticket() {
    local title=$1
    local type=${2:-task}
    
    local output=$(../bin/pm new --type "$type" "$title" 2>&1)
    local id=$(echo "$output" | grep -oP 'SANDBOX-\d+' | head -1)
    echo "$id"
}

set_ticket_fields() {
    local id=$1
    local status=$2
    local priority=$3
    local points=$4
    local labels=$5
    local assignee=$6
    
    if [ -n "$status" ]; then
        ../bin/pm move "$id" "$status" > /dev/null 2>&1
    fi
    
    if [ -n "$priority" ]; then
        ../bin/pm edit "$id" --field priority="$priority" > /dev/null 2>&1
    fi
    
    if [ -n "$points" ]; then
        ../bin/pm edit "$id" --field points="$points" > /dev/null 2>&1
    fi
    
    if [ -n "$labels" ]; then
        ../bin/pm edit "$id" --field labels="$labels" > /dev/null 2>&1
    fi
    
    if [ -n "$assignee" ]; then
        ../bin/pm assign "$id" "$assignee" > /dev/null 2>&1
    fi
}

set_parent() {
    local id=$1
    local parent=$2
    
    ../bin/pm edit "$id" --field parent="$parent" > /dev/null 2>&1
}

set_dependency() {
    local from=$1
    local to=$2
    ../bin/pm link "$from" "$to" --type depends-on > /dev/null 2>&1
}

set_blocking() {
    local from=$1
    local to=$2
    ../bin/pm link "$from" "$to" --type blocks > /dev/null 2>&1
}

set_related() {
    local from=$1
    local to=$2
    ../bin/pm link "$from" "$to" --type related > /dev/null 2>&1
}

create_milestone() {
    local id=$1
    local title=$2
    local due=$3
    local desc=$4
    ../bin/pm milestone create "$title" --id "$id" --due "$due" --description "$desc" > /dev/null
    echo "$id"
}

assign_milestone() {
    local ticket_id=$1
    local milestone_id=$2
    ../bin/pm edit "$ticket_id" --field milestones="$milestone_id" > /dev/null 2>&1
}

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Creating 16-Ticket Scenario: User Authentication System${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"

# Phase 1: Create Epic
echo ""
echo -e "${YELLOW}Phase 1: Creating Epic...${NC}"
EPIC_ID=$(create_ticket "User Authentication System" "epic")
set_ticket_fields "$EPIC_ID" "backlog" "critical" "21" "feature,backend" "alice"

# Phase 2: Create Stories under Epic
echo -e "${YELLOW}Phase 2: Creating Stories...${NC}"
STORY_1=$(create_ticket "OAuth2 Integration" "story")
set_parent "$STORY_1" "$EPIC_ID"
set_ticket_fields "$STORY_1" "in-progress" "high" "8" "feature,backend" "alice"

STORY_2=$(create_ticket "Password Authentication" "story")
set_parent "$STORY_2" "$EPIC_ID"
set_ticket_fields "$STORY_2" "todo" "high" "5" "feature,backend" "bob"

STORY_3=$(create_ticket "Session Management" "story")
set_parent "$STORY_3" "$EPIC_ID"
set_ticket_fields "$STORY_3" "backlog" "medium" "5" "feature,backend" "alice"

STORY_4=$(create_ticket "Login UI" "story")
set_parent "$STORY_4" "$EPIC_ID"
set_ticket_fields "$STORY_4" "backlog" "medium" "3" "feature,frontend" "charlie"

# Phase 3: Create OAuth Tasks
echo -e "${YELLOW}Phase 3: Creating OAuth Tasks...${NC}"
TASK_1=$(create_ticket "Setup Google Provider" "task")
set_parent "$TASK_1" "$STORY_1"
set_ticket_fields "$TASK_1" "done" "high" "3" "feature,backend" "alice"

TASK_2=$(create_ticket "Setup GitHub Provider" "task")
set_parent "$TASK_2" "$STORY_1"
set_dependency "$TASK_2" "$TASK_1"
set_ticket_fields "$TASK_2" "in-progress" "high" "3" "feature,backend" "alice"

TASK_3=$(create_ticket "Add OAuth Middleware" "task")
set_parent "$TASK_3" "$STORY_1"
set_dependency "$TASK_3" "$TASK_2"
set_ticket_fields "$TASK_3" "backlog" "high" "2" "feature,backend" "alice"

# Phase 4: Create Password Tasks
echo -e "${YELLOW}Phase 4: Creating Password Tasks...${NC}"
TASK_4=$(create_ticket "Hash Password Implementation" "task")
set_parent "$TASK_4" "$STORY_2"
set_ticket_fields "$TASK_4" "done" "high" "2" "feature,backend" "bob"

TASK_5=$(create_ticket "Login Endpoint" "task")
set_parent "$TASK_5" "$STORY_2"
set_dependency "$TASK_5" "$TASK_4"
set_ticket_fields "$TASK_5" "todo" "high" "3" "feature,backend" "bob"

TASK_6=$(create_ticket "Logout Endpoint" "task")
set_parent "$TASK_6" "$STORY_2"
set_ticket_fields "$TASK_6" "backlog" "high" "2" "feature,backend" "bob"

# Phase 5: Create Session Tasks
echo -e "${YELLOW}Phase 5: Creating Session Management Tasks...${NC}"
TASK_7=$(create_ticket "JWT Token Generation" "task")
set_parent "$TASK_7" "$STORY_3"
set_dependency "$TASK_7" "$TASK_4"
set_ticket_fields "$TASK_7" "backlog" "medium" "3" "feature,backend" "alice"

TASK_8=$(create_ticket "Token Refresh" "task")
set_parent "$TASK_8" "$STORY_3"
set_dependency "$TASK_8" "$TASK_7"
set_ticket_fields "$TASK_8" "backlog" "medium" "2" "feature,backend" "alice"

# Phase 6: Create UI Task
echo -e "${YELLOW}Phase 6: Creating UI Tasks...${NC}"
TASK_9=$(create_ticket "Login Form Component" "task")
set_parent "$TASK_9" "$STORY_4"
set_ticket_fields "$TASK_9" "backlog" "medium" "3" "feature,frontend" "charlie"

# Phase 7: Create Bugs
echo -e "${YELLOW}Phase 7: Creating Bugs...${NC}"
BUG_1=$(create_ticket "Login Token Expiration Edge Case" "bug")
set_related "$BUG_1" "$TASK_7"
set_ticket_fields "$BUG_1" "done" "medium" "2" "bug,backend" "alice"

BUG_2=$(create_ticket "Session Race Condition" "bug")
set_blocking "$BUG_2" "$TASK_8"
set_ticket_fields "$BUG_2" "in-progress" "high" "3" "bug,backend" "bob"

# Phase 8: Add Comments to key tickets
../bin/pm comment SANDBOX-7 -m "This is the first comment"
../bin/pm comment SANDBOX-7 -m "This is the second comment"

# Phase 9: Create Milestones
echo ""
echo -e "${YELLOW}Phase 9: Creating Milestones...${NC}"

# Milestone A: overdue sprint covering the OAuth work (past due date)
MS_1=$(create_milestone "v0-1-oauth-mvp" "v0.1 OAuth MVP" "2026-01-31" "Minimum viable OAuth2 authentication")

# Assign OAuth tasks + related bug to milestone A (10 pts total, 2 done)
assign_milestone "$TASK_1" "$MS_1"   # Setup Google Provider     (3 pts, done)
assign_milestone "$TASK_2" "$MS_1"   # Setup GitHub Provider     (3 pts, in-progress)
assign_milestone "$TASK_3" "$MS_1"   # Add OAuth Middleware      (2 pts, backlog)
assign_milestone "$BUG_1"  "$MS_1"   # Login Token Expiration    (2 pts, done)

# Milestone B: future release milestone covering all stories (future due date)
MS_2=$(create_milestone "v1-0-release" "v1.0 Release" "2026-12-31" "Full authentication system ready for production")

# Assign stories + password task + session bug to milestone B (26 pts total, 1 done)
assign_milestone "$STORY_1" "$MS_2"  # OAuth2 Integration        (8 pts, in-progress)
assign_milestone "$STORY_2" "$MS_2"  # Password Authentication   (5 pts, todo)
assign_milestone "$STORY_3" "$MS_2"  # Session Management        (5 pts, backlog)
assign_milestone "$STORY_4" "$MS_2"  # Login UI                  (3 pts, backlog)
assign_milestone "$TASK_4"  "$MS_2"  # Hash Password Impl        (2 pts, done)
assign_milestone "$BUG_2"   "$MS_2"  # Session Race Condition    (3 pts, in-progress)

echo -e "${GREEN}✓ Milestones created and populated${NC}"


echo ""
echo -e "${GREEN}✓ Sandbox created successfully!${NC}"

# Summary
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Sandbox Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "📊 Tickets Created: 16"
echo "   • 1 Epic ($EPIC_ID)"
echo "   • 4 Stories ($STORY_1, $STORY_2, $STORY_3, $STORY_4)"
echo "   • 9 Tasks ($TASK_1-$TASK_9)"
echo "   • 2 Bugs ($BUG_1, $BUG_2)"
echo ""
echo "🗓️  Milestones Created: 2"
echo "   • $MS_1 — v0.1 OAuth MVP        (due 2026-01-31, overdue, 10 pts)"
echo "   • $MS_2 — v1.0 Release          (due 2026-12-31, active,  26 pts)"
echo ""
echo "🔗 Relationships:"
echo "   • Parent-child hierarchy (Epic → Stories → Tasks)"
echo "   • 5 dependency chains (pm link --type depends-on, symmetric blocks)"
echo "   • 1 blocking relationship (pm link --type blocks, symmetric depends-on)"
echo "   • 1 related association (pm link --type related)"
echo ""
echo "👥 Assignees: alice, bob, charlie"
echo "🏷️  Labels: feature, backend, frontend, bug"
echo "📈 States: backlog, todo, in-progress, done"
echo "⭐ Priorities: critical, high, medium"
echo ""

# Display example commands
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Explore the Sandbox${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Try these commands in the sandbox/ directory:"
echo ""
echo -e "${YELLOW}View the epic and its children:${NC}"
echo "  ../bin/pm list --parent $EPIC_ID"
echo ""
echo -e "${YELLOW}Show a ticket with comments:${NC}"
echo "  ../bin/pm show $TASK_2"
echo ""
echo -e "${YELLOW}See tickets with dependencies:${NC}"
echo "  ../bin/pm list"
echo ""
echo -e "${YELLOW}Check a ticket with dependencies:${NC}"
echo "  ../bin/pm show SANDBOX-7"
echo ""
echo -e "${YELLOW}List tickets by status:${NC}"
echo "  ../bin/pm list --status in-progress"
echo ""
echo -e "${YELLOW}Filter by assignee:${NC}"
echo "  ../bin/pm list --assignee alice"
echo ""
echo -e "${BLUE}── Relationship Filters ────────────────────────────────────${NC}"
echo ""
echo -e "${YELLOW}Who is waiting on a ticket to finish?${NC}"
echo "  ../bin/pm list --depends-on $TASK_4 --all"
echo ""
echo -e "${YELLOW}What is blocking a ticket?${NC}"
echo "  ../bin/pm list --blocks $TASK_8 --all"
echo ""
echo -e "${YELLOW}Show all tickets related to a ticket:${NC}"
echo "  ../bin/pm list --related $TASK_7 --all"
echo ""
echo -e "${YELLOW}Combine: blocked work that is actively in-progress:${NC}"
echo "  ../bin/pm list --depends-on $TASK_4 --status in-progress --all"
echo ""
echo -e "${BLUE}── Milestones ──────────────────────────────────────────────${NC}"
echo ""
echo -e "${YELLOW}List all milestones:${NC}"
echo "  ../bin/pm milestone list"
echo ""
echo -e "${YELLOW}Show overdue milestones only:${NC}"
echo "  ../bin/pm milestone list --overdue"
echo ""
echo -e "${YELLOW}Show milestones with progress percentages:${NC}"
echo "  ../bin/pm milestone list --with-progress"
echo ""
echo -e "${YELLOW}Show milestone details with progress bars:${NC}"
echo "  ../bin/pm milestone show $MS_1"
echo "  ../bin/pm milestone show $MS_2"
echo ""
echo -e "${YELLOW}List tickets belonging to a milestone:${NC}"
echo "  ../bin/pm list --milestone $MS_2"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Directory: sandbox/"
echo "Next steps:"
echo "  cd sandbox"
echo "  ../bin/pm <command>"
echo ""

