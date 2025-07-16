#!/bin/bash

# Debug script for Virtuoso MCP Server
echo "🔍 Debugging Virtuoso MCP Server..."
echo "=================================="

# Test if the server can start
echo -e "\n1. Testing server startup..."
timeout 5 node /Users/marklovelady/_dev/_projects/virtuoso-GENerator/virtuoso-mcp-server/dist/index.js 2>&1 | head -10

# Check if CLI exists
echo -e "\n2. Checking Virtuoso CLI..."
if [ -f "/Users/marklovelady/_dev/_projects/virtuoso-GENerator/bin/api-cli" ]; then
    echo "✅ CLI found at: /Users/marklovelady/_dev/_projects/virtuoso-GENerator/bin/api-cli"
    ls -la /Users/marklovelady/_dev/_projects/virtuoso-GENerator/bin/api-cli
else
    echo "❌ CLI not found!"
fi

# Check config
echo -e "\n3. Checking Virtuoso config..."
if [ -f "/Users/marklovelady/.api-cli/virtuoso-config.yaml" ]; then
    echo "✅ Config found"
    echo "First few lines:"
    head -5 /Users/marklovelady/.api-cli/virtuoso-config.yaml
else
    echo "❌ Config not found!"
fi

# Check Claude Desktop logs
echo -e "\n4. Recent Claude Desktop logs (if any errors)..."
LOG_DIR="$HOME/Library/Logs/Claude"
if [ -d "$LOG_DIR" ]; then
    echo "Checking logs in: $LOG_DIR"
    find "$LOG_DIR" -name "*.log" -mtime -1 -exec echo "Log file: {}" \; -exec tail -20 {} \; 2>/dev/null | grep -i "virtuoso\|error" || echo "No recent errors found"
else
    echo "Log directory not found"
fi

# Test MCP protocol
echo -e "\n5. Testing MCP protocol..."
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024.11","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}' | node /Users/marklovelady/_dev/_projects/virtuoso-GENerator/virtuoso-mcp-server/dist/index.js 2>&1 | head -5

echo -e "\n=================================="
echo "Debug complete!"
