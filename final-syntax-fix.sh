#!/bin/bash

# Final fix for syntax errors - remove extra closing braces

echo "Applying final syntax fixes..."

# Find all files with syntax errors and fix them
find /Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd -name "*.go" -type f | while read file; do
    # Create a temporary file
    tmpfile="${file}.tmp"
    
    # Process the file to remove extra closing braces before return statements
    awk '
    BEGIN { 
        brace_count = 0
        lines_buffer = ""
        buffer_count = 0
    }
    {
        # Count braces
        gsub(/[^{}]/, "", $0)
        for (i = 1; i <= length($0); i++) {
            if (substr($0, i, 1) == "{") brace_count++
            else if (substr($0, i, 1) == "}") brace_count--
        }
        
        # Reset and read original line
        $0 = lines[NR]
        
        # Buffer the line
        lines_buffer = lines_buffer "\n" $0
        buffer_count++
        
        # If we see "return nil" and have buffered lines, process them
        if ($0 ~ /^\s*return nil\s*$/ && buffer_count > 3) {
            # Check if we have double closing braces before return
            split(lines_buffer, buffer_lines, "\n")
            
            # Count how many lines have just closing braces before return
            closing_brace_count = 0
            for (i = length(buffer_lines) - 2; i >= length(buffer_lines) - 4 && i > 0; i--) {
                if (buffer_lines[i] ~ /^\s*}\s*$/) {
                    closing_brace_count++
                }
            }
            
            # If we have more than one closing brace, skip one
            if (closing_brace_count > 1) {
                skip_next_brace = 1
            }
            
            # Print buffered lines, potentially skipping a brace
            for (i = 2; i <= length(buffer_lines); i++) {
                if (skip_next_brace && buffer_lines[i] ~ /^\s*}\s*$/ && i < length(buffer_lines) - 1) {
                    skip_next_brace = 0
                    continue
                }
                if (buffer_lines[i] != "") print buffer_lines[i]
            }
            
            # Clear buffer
            lines_buffer = ""
            buffer_count = 0
        }
    }
    BEGIN {
        # Read the entire file first
        while ((getline line) > 0) {
            lines[++NR] = line
        }
        NR = 0
        
        # Process the file
        for (i = 1; i <= length(lines); i++) {
            NR = i
            $0 = lines[i]
            
            # For the last few lines, just print them
            if (i > length(lines) - 5 || lines_buffer == "") {
                print $0
            }
        }
    }
    ' "$file" > "$tmpfile" 2>/dev/null
    
    # Only replace if the temp file is valid and non-empty
    if [ -s "$tmpfile" ]; then
        # Do a simple check - remove extra closing braces before return nil
        perl -i -pe 'BEGIN{undef $/;} s/\}\s*\}\s*\n\s*return nil/\}\n\n\treturn nil/smg' "$tmpfile"
        
        # Another pass to clean up any remaining issues
        sed -i '' -E 's/^(\s*})\s*}(\s*)$/\1\2/' "$tmpfile"
        
        mv "$tmpfile" "$file"
        echo "Fixed: $(basename "$file")"
    else
        rm -f "$tmpfile"
    fi
done

echo "Final syntax fixes applied!"