# Virtuoso CLI Docker Wrapper Script (PowerShell)
# This script simplifies running the Virtuoso CLI in a Docker container on Windows

param(
    [switch]$WrapperHelp,
    [switch]$Build,
    [switch]$Shell,
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Arguments
)

# Configuration
$DOCKER_IMAGE = "virtuoso-cli:latest"
$CONTAINER_NAME = "virtuoso-cli-runner"
$CONFIG_DIR = "$env:USERPROFILE\.virtuoso"
$WORKSPACE_DIR = Get-Location

# Function to print colored output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to check if Docker is running
function Test-Docker {
    try {
        docker info | Out-Null
        return $true
    }
    catch {
        Write-Error "Docker is not running. Please start Docker and try again."
        exit 1
    }
}

# Function to check if image exists
function Test-Image {
    try {
        docker image inspect $DOCKER_IMAGE | Out-Null
        return $true
    }
    catch {
        Write-Warning "Docker image '$DOCKER_IMAGE' not found. Building..."
        docker build -t $DOCKER_IMAGE .
    }
}

# Function to clean up any existing container
function Remove-Container {
    try {
        $containers = docker ps -a --format "table {{.Names}}" | Select-String "^${CONTAINER_NAME}$"
        if ($containers) {
            Write-Info "Cleaning up existing container..."
            docker rm -f $CONTAINER_NAME | Out-Null
        }
    }
    catch {
        # Container doesn't exist, ignore
    }
}

# Function to setup config directory
function Initialize-Config {
    if (!(Test-Path $CONFIG_DIR)) {
        Write-Info "Creating config directory at $CONFIG_DIR"
        New-Item -ItemType Directory -Path $CONFIG_DIR -Force | Out-Null
    }
}

# Function to run the CLI command
function Invoke-CLI {
    param([string[]]$Args)
    
    Test-Docker
    Test-Image
    Remove-Container
    Initialize-Config
    
    Write-Info "Running: virtuoso-cli $($Args -join ' ')"
    
    # Convert Windows paths to Unix-style for Docker
    $workspaceMount = ($WORKSPACE_DIR -replace '\\', '/').Replace(':', '')
    $configMount = ($CONFIG_DIR -replace '\\', '/').Replace(':', '')
    
    # Run the container with the command
    docker run --rm `
        --name $CONTAINER_NAME `
        -v "${workspaceMount}:/workspace" `
        -v "${configMount}:/home/apiuser/.virtuoso" `
        -v "$(Split-Path $WORKSPACE_DIR -Parent)/config:/config" `
        -e VIRTUOSO_API_TOKEN=$env:VIRTUOSO_API_TOKEN `
        -e VIRTUOSO_BASE_URL=$env:VIRTUOSO_BASE_URL `
        -e VIRTUOSO_ORG_ID=$env:VIRTUOSO_ORG_ID `
        -e VIRTUOSO_CONFIG_PATH=/config `
        -e VIRTUOSO_OUTPUT_FORMAT=$env:VIRTUOSO_OUTPUT_FORMAT `
        $DOCKER_IMAGE `
        @Args
}

# Function to show help
function Show-Help {
    @"
Virtuoso CLI Docker Wrapper (PowerShell)

Usage: .\virtuoso.ps1 [OPTIONS] [COMMAND]

This script runs the Virtuoso CLI in a Docker container with proper volume mounts
and environment variable forwarding.

Environment Variables:
  VIRTUOSO_API_TOKEN    - API token for authentication
  VIRTUOSO_BASE_URL     - Base URL for Virtuoso API
  VIRTUOSO_ORG_ID       - Organization ID
  VIRTUOSO_OUTPUT_FORMAT - Output format (human, json, yaml, ai)

Examples:
  .\virtuoso.ps1 --help                           # Show CLI help
  .\virtuoso.ps1 --version                        # Show version
  .\virtuoso.ps1 create-project "Test Project"    # Create a new project
  .\virtuoso.ps1 list-projects                    # List all projects
  .\virtuoso.ps1 validate-config                  # Validate configuration

Special Commands:
  .\virtuoso.ps1 -WrapperHelp                     # Show this help
  .\virtuoso.ps1 -Build                           # Rebuild the Docker image
  .\virtuoso.ps1 -Shell                           # Open interactive shell in container

"@
}

# Function to build the image
function Build-Image {
    Write-Info "Building Docker image..."
    docker build -t $DOCKER_IMAGE .
    Write-Info "Image built successfully"
}

# Function to open interactive shell
function Open-Shell {
    Test-Docker
    Test-Image
    Remove-Container
    Initialize-Config
    
    Write-Info "Opening interactive shell in container..."
    
    # Convert Windows paths to Unix-style for Docker
    $workspaceMount = ($WORKSPACE_DIR -replace '\\', '/').Replace(':', '')
    $configMount = ($CONFIG_DIR -replace '\\', '/').Replace(':', '')
    
    docker run -it --rm `
        --name $CONTAINER_NAME `
        -v "${workspaceMount}:/workspace" `
        -v "${configMount}:/home/apiuser/.virtuoso" `
        -v "$(Split-Path $WORKSPACE_DIR -Parent)/config:/config" `
        -e VIRTUOSO_API_TOKEN=$env:VIRTUOSO_API_TOKEN `
        -e VIRTUOSO_BASE_URL=$env:VIRTUOSO_BASE_URL `
        -e VIRTUOSO_ORG_ID=$env:VIRTUOSO_ORG_ID `
        -e VIRTUOSO_CONFIG_PATH=/config `
        -e VIRTUOSO_OUTPUT_FORMAT=$env:VIRTUOSO_OUTPUT_FORMAT `
        --entrypoint /bin/sh `
        $DOCKER_IMAGE
}

# Main script logic
if ($WrapperHelp) {
    Show-Help
}
elseif ($Build) {
    Build-Image
}
elseif ($Shell) {
    Open-Shell
}
else {
    # Pass all arguments to the CLI
    Invoke-CLI $Arguments
}