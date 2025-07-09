#!/bin/bash
# example-calls.sh - Example API CLI calls
# This script demonstrates various ways to use the generated CLI

# Set API key (replace with your actual key)
export API_CLI_API_KEY="your-api-key-here"

# Colours for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}API CLI Example Calls${NC}"
echo "====================="

# Example 1: List users with default settings
echo -e "\n${GREEN}1. List users (default)${NC}"
echo "Command: api-cli users list"
api-cli users list

# Example 2: List users with pagination
echo -e "\n${GREEN}2. List users with pagination${NC}"
echo "Command: api-cli users list --limit 5 --offset 10"
api-cli users list --limit 5 --offset 10

# Example 3: Get specific user
echo -e "\n${GREEN}3. Get specific user${NC}"
echo "Command: api-cli users get user-123"
api-cli users get user-123

# Example 4: Create a new user
echo -e "\n${GREEN}4. Create new user${NC}"
echo 'Command: api-cli users create --name "Jane Smith" --email jane@example.com'
api-cli users create --name "Jane Smith" --email jane@example.com

# Example 5: Create admin user
echo -e "\n${GREEN}5. Create admin user${NC}"
echo 'Command: api-cli users create --name "Admin User" --email admin@example.com --role admin'
api-cli users create --name "Admin User" --email admin@example.com --role admin

# Example 6: Different output formats
echo -e "\n${GREEN}6. Different output formats${NC}"
echo "JSON format:"
api-cli users list --limit 3 -o json

echo -e "\nYAML format:"
api-cli users list --limit 3 -o yaml

echo -e "\nTable format:"
api-cli users list --limit 3 -o table

# Example 7: Using different base URL
echo -e "\n${GREEN}7. Using staging server${NC}"
echo "Command: api-cli --base-url https://staging-api.example.com/v1 users list"
api-cli --base-url https://staging-api.example.com/v1 users list

# Example 8: Verbose mode for debugging
echo -e "\n${GREEN}8. Verbose mode${NC}"
echo "Command: api-cli -v users get user-123"
api-cli -v users get user-123

# Example 9: Using configuration file
echo -e "\n${GREEN}9. Using config file${NC}"
cat > ~/.api-cli.yaml << EOF
base_url: https://api.example.com/v1
api_key: ${API_CLI_API_KEY}
output: table
verbose: true
EOF
echo "Config file created at ~/.api-cli.yaml"
echo "Command: api-cli users list (will use config file)"
api-cli users list

# Example 10: Error handling
echo -e "\n${GREEN}10. Error handling${NC}"
echo "Command: api-cli users get non-existent-user"
api-cli users get non-existent-user || echo "Error handled gracefully"

# Example 11: Pipe to jq for JSON processing
echo -e "\n${GREEN}11. JSON processing with jq${NC}"
echo "Command: api-cli users list -o json | jq '.users[].name'"
if command -v jq &> /dev/null; then
    api-cli users list -o json | jq '.users[].name'
else
    echo "jq not installed, skipping example"
fi

# Example 12: Script automation
echo -e "\n${GREEN}12. Script automation${NC}"
echo "Creating users from a list..."
cat > /tmp/users.txt << EOF
Alice Johnson,alice@example.com,user
Bob Wilson,bob@example.com,moderator
Charlie Brown,charlie@example.com,admin
EOF

while IFS=',' read -r name email role; do
    echo "Creating user: $name"
    api-cli users create --name "$name" --email "$email" --role "$role"
done < /tmp/users.txt

echo -e "\n${BLUE}Examples complete!${NC}"
