#!/bin/bash

# Common functions for all sub-agents
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

RESULTS_DIR="ultrathink-debug-results"

log() {
    local level=$1
    local agent=$2
    local message=$3
    local color=$NC
    
    case $level in
        "INFO") color=$BLUE ;;
        "SUCCESS") color=$GREEN ;;
        "WARNING") color=$YELLOW ;;
        "ERROR") color=$RED ;;
        "AGENT") color=$PURPLE ;;
        "DEBUG") color=$CYAN ;;
    esac
    
    echo -e "${color}[$(date +%H:%M:%S)] [$agent] $message${NC}"
}
