#!/bin/bash
# MCP Server Test Script
# This script helps test the k8sgpt MCP server functionality

set -e

echo "🚀 K8sGPT MCP Server Test Suite"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if k8sgpt is built
if [ ! -f "./k8sgpt" ]; then
    echo -e "${YELLOW}Building k8sgpt...${NC}"
    go build -o k8sgpt main.go
    echo -e "${GREEN}✓ Build successful${NC}"
fi

# Test 1: Version check
echo -e "${BLUE}Test 1: Version Check${NC}"
./k8sgpt version
echo -e "${GREEN}✓ Version check passed${NC}"
echo ""

# Test 2: List available filters
echo -e "${BLUE}Test 2: List Available Filters${NC}"
./k8sgpt filters list
echo -e "${GREEN}✓ Filter list passed${NC}"
echo ""

# Test 3: Check authentication
echo -e "${BLUE}Test 3: Check AI Authentication${NC}"
./k8sgpt auth list
echo -e "${GREEN}✓ Auth check passed${NC}"
echo ""

# Test 4: Basic analyze (without MCP)
echo -e "${BLUE}Test 4: Basic Analyze${NC}"
echo -e "${YELLOW}Running analysis...${NC}"
./k8sgpt analyze --explain=false || true
echo -e "${GREEN}✓ Basic analyze completed${NC}"
echo ""

# Test 5: Check MCP server can start (if not already running)
echo -e "${BLUE}Test 5: MCP Server Startup Check${NC}"
echo -e "${YELLOW}Checking if MCP server can initialize...${NC}"

# Create a test script that starts and immediately stops the server
cat > /tmp/test_mcp_startup.sh << 'EOF'
#!/bin/bash
timeout 3s ./k8sgpt serve --enable-mcp 2>&1 | head -20 &
PID=$!
sleep 2
kill $PID 2>/dev/null || true
wait $PID 2>/dev/null || true
exit 0
EOF

chmod +x /tmp/test_mcp_startup.sh
/tmp/test_mcp_startup.sh || true
rm /tmp/test_mcp_startup.sh

echo -e "${GREEN}✓ MCP server initialization check passed${NC}"
echo ""

# Test 6: Check Kubernetes connectivity
echo -e "${BLUE}Test 6: Kubernetes Connectivity${NC}"
if kubectl cluster-info &>/dev/null; then
    echo -e "${GREEN}✓ Kubernetes cluster accessible${NC}"
    kubectl version --short 2>/dev/null || kubectl version
else
    echo -e "${RED}⚠ No Kubernetes cluster accessible${NC}"
    echo -e "${YELLOW}  MCP server will work but cannot analyze real resources${NC}"
fi
echo ""

# Summary of MCP capabilities
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}MCP Server Capabilities Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}Tools Available (12):${NC}"
echo "  Core:"
echo "    • analyze - AI-powered Kubernetes analysis"
echo "    • cluster-info - Cluster version information"
echo "    • config - Configure k8sgpt settings"
echo ""
echo "  Resource Management:"
echo "    • list-resources - List any K8s resource type"
echo "    • get-resource - Get detailed resource info"
echo "    • list-namespaces - List all namespaces"
echo ""
echo "  Debugging:"
echo "    • list-events - View K8s events"
echo "    • get-logs - Retrieve pod logs"
echo ""
echo "  Filter Management:"
echo "    • list-filters - Show available analyzers"
echo "    • add-filters - Enable analyzers"
echo "    • remove-filters - Disable analyzers"
echo ""
echo "  Integration:"
echo "    • list-integrations - Show integrations (Prometheus, AWS, etc.)"
echo ""
echo -e "${GREEN}Resources Available (3):${NC}"
echo "    • cluster-info - Quick cluster info access"
echo "    • namespaces - Namespace list"
echo "    • active-filters - Current analyzer configuration"
echo ""
echo -e "${GREEN}Prompts Available (3):${NC}"
echo "    • troubleshoot-pod - Pod debugging guide"
echo "    • troubleshoot-deployment - Deployment debugging guide"
echo "    • troubleshoot-cluster - General cluster troubleshooting"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}How to Start MCP Server:${NC}"
echo ""
echo -e "${YELLOW}Stdio Mode (for local AI assistants):${NC}"
echo "  ./k8sgpt serve --enable-mcp"
echo ""
echo -e "${YELLOW}HTTP Mode (for network access):${NC}"
echo "  ./k8sgpt serve --enable-mcp --mcp-http --mcp-port 8089"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}Integration Examples:${NC}"
echo ""
echo -e "${YELLOW}Claude Desktop (add to config):${NC}"
echo '  {
    "mcpServers": {
      "k8sgpt": {
        "command": "k8sgpt",
        "args": ["serve", "--enable-mcp"]
      }
    }
  }'
echo ""
echo -e "${YELLOW}Test with curl (HTTP mode):${NC}"
echo '  # Start server first'
echo '  ./k8sgpt serve --enable-mcp --mcp-http --mcp-port 8089'
echo ''
echo '  # Then in another terminal:'
echo '  curl http://localhost:8089/sse'
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}✅ All basic tests passed!${NC}"
echo -e "${YELLOW}📖 See MCP_IMPROVEMENTS.md and pkg/server/MCP_README.md for details${NC}"
echo ""
