#!/bin/bash

# Migration tool for transitioning from old CLI commands to new consolidated ones
# Usage: ./migrate-commands.sh [options] [file1 file2 ...]
#
# Options:
#   -h, --help           Show this help message
#   -d, --dry-run        Show what would be changed without modifying files
#   -r, --report         Generate a migration report only
#   -o, --output FILE    Write report to file (default: migration-report.txt)
#   -a, --auto-update    Automatically update scripts (prompts for confirmation)
#   -b, --backup         Create backup files before updating (*.backup)
#   -v, --verbose        Show detailed processing information

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DRY_RUN=false
REPORT_ONLY=false
AUTO_UPDATE=false
CREATE_BACKUP=false
VERBOSE=false
OUTPUT_FILE="migration-report.txt"
FILES_TO_PROCESS=()

# Command mapping from old to new format
declare -A COMMAND_MAP=(
    # Assert commands
    ["create-step-assert-equals"]="assert equals"
    ["create-step-assert-not-equals"]="assert not-equals"
    ["create-step-assert-exists"]="assert exists"
    ["create-step-assert-not-exists"]="assert not-exists"
    ["create-step-assert-checked"]="assert checked"
    ["create-step-assert-selected"]="assert selected"
    ["create-step-assert-variable"]="assert variable"
    ["create-step-assert-greater-than"]="assert gt"
    ["create-step-assert-greater-than-or-equal"]="assert gte"
    ["create-step-assert-less-than"]="assert lt"
    ["create-step-assert-less-than-or-equal"]="assert lte"
    ["create-step-assert-matches"]="assert matches"

    # Interact commands
    ["create-step-click"]="interact click"
    ["create-step-double-click"]="interact double-click"
    ["create-step-right-click"]="interact right-click"
    ["create-step-hover"]="interact hover"
    ["create-step-write"]="interact write"
    ["create-step-key"]="interact key"

    # Navigate commands
    ["create-step-navigate"]="navigate url"
    ["create-step-scroll-position"]="navigate scroll-to"
    ["create-step-scroll-top"]="navigate scroll-top"
    ["create-step-scroll-bottom"]="navigate scroll-bottom"
    ["create-step-scroll-element"]="navigate scroll-element"

    # Window commands
    ["create-step-window-resize"]="window resize"
    ["create-step-switch-next-tab"]="window switch-tab next"
    ["create-step-switch-prev-tab"]="window switch-tab prev"
    ["create-step-switch-iframe"]="window switch-frame"
    ["create-step-switch-parent-frame"]="window switch-frame parent"

    # Mouse commands
    ["create-step-mouse-move-to"]="mouse move-to"
    ["create-step-mouse-move-by"]="mouse move-by"
    ["create-step-mouse-move"]="mouse move"
    ["create-step-mouse-down"]="mouse down"
    ["create-step-mouse-up"]="mouse up"
    ["create-step-mouse-enter"]="mouse enter"

    # Data commands
    ["create-step-store-element-text"]="data store-text"
    ["create-step-store-literal-value"]="data store-value"
    ["create-step-cookie-create"]="data cookie-create"
    ["create-step-delete-cookie"]="data cookie-delete"
    ["create-step-cookie-wipe-all"]="data cookie-clear"

    # Dialog commands
    ["create-step-dismiss-alert"]="dialog dismiss-alert"
    ["create-step-dismiss-confirm"]="dialog dismiss-confirm"
    ["create-step-dismiss-prompt"]="dialog dismiss-prompt"
    ["create-step-dismiss-prompt-with-text"]="dialog dismiss-prompt"

    # Wait commands
    ["create-step-wait-element"]="wait element"
    ["create-step-wait-for-element-default"]="wait element"
    ["create-step-wait-for-element-timeout"]="wait element"
    ["create-step-wait-time"]="wait time"

    # File commands
    ["create-step-upload"]="file upload"
    ["create-step-upload-url"]="file upload"

    # Select commands
    ["create-step-pick"]="select option"
    ["create-step-pick-index"]="select index"
    ["create-step-pick-last"]="select last"

    # Misc commands
    ["create-step-comment"]="misc comment"
    ["create-step-execute-script"]="misc execute-script"
)

# Function to display help
show_help() {
    echo "Migration tool for transitioning from old CLI commands to new consolidated ones"
    echo ""
    echo "Usage: $0 [options] [file1 file2 ...]"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  -d, --dry-run        Show what would be changed without modifying files"
    echo "  -r, --report         Generate a migration report only"
    echo "  -o, --output FILE    Write report to file (default: migration-report.txt)"
    echo "  -a, --auto-update    Automatically update scripts (prompts for confirmation)"
    echo "  -b, --backup         Create backup files before updating (*.backup)"
    echo "  -v, --verbose        Show detailed processing information"
    echo ""
    echo "Examples:"
    echo "  # Generate report for all shell scripts in current directory"
    echo "  $0 -r *.sh"
    echo ""
    echo "  # Dry run to see what would change"
    echo "  $0 -d my-script.sh"
    echo ""
    echo "  # Auto-update with backups"
    echo "  $0 -a -b *.sh"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -r|--report)
            REPORT_ONLY=true
            shift
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -a|--auto-update)
            AUTO_UPDATE=true
            shift
            ;;
        -b|--backup)
            CREATE_BACKUP=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
        *)
            FILES_TO_PROCESS+=("$1")
            shift
            ;;
    esac
done

# If no files specified, process all .sh files in current directory
if [ ${#FILES_TO_PROCESS[@]} -eq 0 ]; then
    FILES_TO_PROCESS=(*.sh)
fi

# Initialize counters
TOTAL_FILES=0
FILES_WITH_CHANGES=0
TOTAL_COMMANDS_FOUND=0
declare -A COMMAND_USAGE_COUNT

# Function to analyze a file
analyze_file() {
    local file="$1"
    local found_commands=()
    local line_num=0

    if [ ! -f "$file" ]; then
        [ "$VERBOSE" = true ] && echo -e "${YELLOW}Skipping non-existent file: $file${NC}"
        return
    fi

    [ "$VERBOSE" = true ] && echo -e "${BLUE}Analyzing: $file${NC}"

    while IFS= read -r line; do
        ((line_num++))

        # Look for old command patterns
        for old_cmd in "${!COMMAND_MAP[@]}"; do
            if [[ "$line" =~ (api-cli|\.\/bin/api-cli)[[:space:]]+${old_cmd}[[:space:]] ]]; then
                found_commands+=("$line_num:$old_cmd:$line")
                ((TOTAL_COMMANDS_FOUND++))
                ((COMMAND_USAGE_COUNT[$old_cmd]++))
            fi
        done
    done < "$file"

    if [ ${#found_commands[@]} -gt 0 ]; then
        ((FILES_WITH_CHANGES++))
        {
            echo "File: $file"
            echo "Found ${#found_commands[@]} commands to migrate:"
            echo ""

            for cmd_info in "${found_commands[@]}"; do
                IFS=':' read -r line_num old_cmd line_content <<< "$cmd_info"
                new_cmd="${COMMAND_MAP[$old_cmd]}"
                echo "  Line $line_num: $old_cmd → $new_cmd"

                if [ "$VERBOSE" = true ]; then
                    echo "    Old: ${line_content:0:100}..."
                    echo "    New: ${line_content//$old_cmd/$new_cmd}"
                fi
            done
            echo ""
        } >> "$OUTPUT_FILE.tmp"
    fi

    ((TOTAL_FILES++))

    # Return the found commands for processing
    printf '%s\n' "${found_commands[@]}"
}

# Function to update a file
update_file() {
    local file="$1"
    local changes="$2"

    if [ "$CREATE_BACKUP" = true ]; then
        cp "$file" "$file.backup"
        echo -e "${GREEN}Created backup: $file.backup${NC}"
    fi

    # Create temporary file
    local temp_file
    temp_file=$(mktemp)

    # Process the file
    while IFS= read -r line; do
        local new_line="$line"

        # Apply all command replacements
        for old_cmd in "${!COMMAND_MAP[@]}"; do
            if [[ "$line" =~ (api-cli|\.\/bin/api-cli)[[:space:]]+${old_cmd}[[:space:]] ]]; then
                new_cmd="${COMMAND_MAP[$old_cmd]}"
                new_line="${line//$old_cmd/$new_cmd}"

                [ "$VERBOSE" = true ] && echo -e "${YELLOW}Updating: $old_cmd → $new_cmd${NC}"
            fi
        done

        echo "$new_line" >> "$temp_file"
    done < "$file"

    # Replace original file
    mv "$temp_file" "$file"
    echo -e "${GREEN}Updated: $file${NC}"
}

# Main processing
echo -e "${BLUE}=== API CLI Command Migration Tool ===${NC}"
echo ""

# Initialize report file
: > "$OUTPUT_FILE.tmp"

# Analyze all files
echo "Analyzing files..."
declare -A FILE_CHANGES

for file in "${FILES_TO_PROCESS[@]}"; do
    if [ -f "$file" ]; then
        changes=$(analyze_file "$file")
        if [ -n "$changes" ]; then
            FILE_CHANGES["$file"]="$changes"
        fi
    fi
done

# Generate summary report
{
    echo "=== API CLI Command Migration Report ==="
    echo "Generated: $(date)"
    echo ""
    echo "Summary:"
    echo "- Files analyzed: $TOTAL_FILES"
    echo "- Files with old commands: $FILES_WITH_CHANGES"
    echo "- Total commands to migrate: $TOTAL_COMMANDS_FOUND"
    echo ""

    if [ $TOTAL_COMMANDS_FOUND -gt 0 ]; then
        echo "Command Usage Statistics:"
        for cmd in "${!COMMAND_USAGE_COUNT[@]}"; do
            echo "  $cmd: ${COMMAND_USAGE_COUNT[$cmd]} occurrences"
        done | sort -t: -k2 -nr
        echo ""
    fi

    cat "$OUTPUT_FILE.tmp"

    echo ""
    echo "Migration Mapping:"
    echo "-----------------"
    for old_cmd in "${!COMMAND_MAP[@]}"; do
        if [ "${COMMAND_USAGE_COUNT[$old_cmd]:-0}" -gt 0 ]; then
            echo "  $old_cmd → ${COMMAND_MAP[$old_cmd]}"
        fi
    done | sort
} > "$OUTPUT_FILE"

# Clean up temp file
rm -f "$OUTPUT_FILE.tmp"

# Display report
cat "$OUTPUT_FILE"

# Process updates if requested
if [ "$REPORT_ONLY" = true ]; then
    echo ""
    echo -e "${GREEN}Report saved to: $OUTPUT_FILE${NC}"
elif [ "$DRY_RUN" = true ]; then
    echo ""
    echo -e "${YELLOW}This was a dry run. No files were modified.${NC}"
    echo -e "${YELLOW}To apply changes, run without -d/--dry-run flag.${NC}"
elif [ "$AUTO_UPDATE" = true ] && [ $FILES_WITH_CHANGES -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Ready to update $FILES_WITH_CHANGES files.${NC}"

    if [ "$CREATE_BACKUP" = true ]; then
        echo "Backups will be created with .backup extension."
    fi

    read -p "Proceed with updates? (y/N) " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        echo "Updating files..."

        for file in "${!FILE_CHANGES[@]}"; do
            update_file "$file" "${FILE_CHANGES[$file]}"
        done

        echo ""
        echo -e "${GREEN}Migration complete! Updated $FILES_WITH_CHANGES files.${NC}"
        echo -e "${GREEN}Report saved to: $OUTPUT_FILE${NC}"
    else
        echo -e "${YELLOW}Migration cancelled.${NC}"
    fi
else
    echo ""
    echo -e "${YELLOW}To apply changes, use -a/--auto-update flag.${NC}"
    echo -e "${GREEN}Report saved to: $OUTPUT_FILE${NC}"
fi
