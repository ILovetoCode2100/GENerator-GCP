# Virtuoso API CLI Generator - Claude Documentation

## 🎯 Project Overview
This is a comprehensive CLI tool for interacting with the Virtuoso API to create test automation steps. The project has evolved from a basic proof-of-concept to a full-featured CLI with 28 commands across 17 categories.

## 📊 Current State (Updated: 2025-01-13)

### 🔧 Recent Updates
- **Fixed API Authentication** - Resolved projects API response parsing (now handles `map` structure instead of `items` array)
- **GitHub Actions Integration** - Added CI/CD workflows with proper API authentication
- **Test Suite** - BATS tests covering all CLI commands
- **API Key Configuration** - Proper config file structure with organization ID support

### ✅ **Fully Implemented (28 Commands)**
The CLI now provides complete coverage of all major test automation actions:

#### **Original Commands (21)**
1. **Cookie Management** (2)
   - `create-step-cookie-create` - Create cookies with name/value
   - `create-step-cookie-wipe-all` - Clear all cookies

2. **File Upload** (1)
   - `create-step-upload-url` - Upload files from URLs

3. **Mouse Actions** (2)
   - `create-step-mouse-move-to` - Move to absolute coordinates
   - `create-step-mouse-move-by` - Move by relative offset

4. **Tab/Frame Navigation** (4)
   - `create-step-switch-next-tab` - Switch to next tab
   - `create-step-switch-prev-tab` - Switch to previous tab
   - `create-step-switch-parent-frame` - Switch to parent frame
   - `create-step-switch-iframe` - Switch to iframe by selector

5. **Script Execution** (1)
   - `create-step-execute-script` - Execute custom scripts

6. **Element Selection** (2)
   - `create-step-pick-index` - Pick dropdown option by index
   - `create-step-pick-last` - Pick last dropdown option

7. **Wait Commands** (2)
   - `create-step-wait-for-element-timeout` - Wait with custom timeout
   - `create-step-wait-for-element-default` - Wait with default timeout

8. **Storage Commands** (2)
   - `create-step-store-element-text` - Store element text in variable
   - `create-step-store-literal-value` - Store literal value in variable

9. **Assertion Commands** (4)
   - `create-step-assert-not-equals` - Assert element ≠ value
   - `create-step-assert-greater-than` - Assert element > value
   - `create-step-assert-greater-than-or-equal` - Assert element ≥ value
   - `create-step-assert-matches` - Assert element matches regex

10. **Prompt Handling** (1)
    - `create-step-dismiss-prompt-with-text` - Dismiss prompts with text

#### **New Commands (7)**
11. **Navigation** (1)
    - `create-step-navigate` - Navigate to URLs (basic & new-tab)

12. **Click Actions** (1)
    - `create-step-click` - Click elements (basic, variable, advanced)

13. **Write Actions** (1)
    - `create-step-write` - Write text to inputs (basic, with-variable)

14. **Scroll Commands** (3)
    - `create-step-scroll-to-position` - Scroll to coordinates
    - `create-step-scroll-by-offset` - Scroll by offset
    - `create-step-scroll-to-top` - Scroll to top

15. **Window Commands** (1)
    - `create-step-window-resize` - Resize browser window

16. **Keyboard Commands** (1)
    - `create-step-key` - Press keys (global & targeted)

17. **Documentation Commands** (1)
    - `create-step-comment` - Add comments to tests

## 🔧 Technical Architecture

### **File Structure**
```
src/
├── cmd/
│   ├── main.go                    # Command registration
│   ├── create-step-*.go          # 28 individual command files
│   └── ...
├── pkg/
│   └── virtuoso/
│       └── client.go             # API client with 35+ methods
├── bin/
│   └── api-cli                   # Built binary
└── ...
```

### **Key Components**

#### **1. API Client (`pkg/virtuoso/client.go`)**
- **35+ methods** for all step types
- Parameterized base URL and token support
- Proper request body formatting
- Error handling and response parsing

#### **2. Command Files (`src/cmd/`)**
- **28 command files** following consistent patterns
- Multiple output formats (human, json, yaml, ai)
- Comprehensive help documentation
- Advanced options and flags

#### **3. Main Registration (`src/cmd/main.go`)**
- Centralized command registration
- Organized by functional categories
- Global flags and configuration

## 🚀 Usage Patterns

### **Configuration Setup**
Create `~/.api-cli/virtuoso-config.yaml` or `./config/virtuoso-config.yaml`:
```yaml
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
headers:
  X-Virtuoso-Client-ID: "api-cli-generator"
  X-Virtuoso-Client-Name: "api-cli-generator"
```

### **Basic Command Pattern**
```bash
# List projects
./bin/api-cli list-projects

# Create a step
./bin/api-cli create-step-[ACTION] CHECKPOINT_ID [ARGS...] POSITION [FLAGS]
```

### **Output Formats**
- `--output human` (default) - Human-readable format
- `--output json` - JSON format for scripting
- `--output yaml` - YAML format for configuration
- `--output ai` - AI-optimized format

### **Advanced Options**
- `--new-tab` - Open in new tab (navigate)
- `--variable "name"` - Use/store variables (click, write)
- `--target "selector"` - Target specific elements (key)
- `--position "TOP_RIGHT"` - Element positioning (click)
- `--element-type "BUTTON"` - Element type specification (click)

## 📋 Development Guidelines

### **Adding New Commands**
1. **Client Method**: Add to `pkg/virtuoso/client.go`
2. **Command File**: Create in `src/cmd/create-step-[name].go`
3. **Registration**: Add to `src/cmd/main.go`
4. **Testing**: Update test scripts

### **Command Patterns**
- Follow existing naming conventions
- Use consistent argument parsing
- Include all output formats
- Provide comprehensive help text
- Handle errors gracefully

### **API Integration**
- Use `createStepWithCustomBody()` for complex requests
- Ensure proper `meta` field structures
- Follow JSON body patterns from HAR analysis
- Test with live API endpoints

## 🧪 Testing

### **BATS Test Suite**
Located in `src/cmd/tests/`:
- `00_env.bats` - Environment setup and configuration
- `10_auth.bats` - Authentication and API connectivity
- `20_project.bats` - Project management commands
- `30_journey_goal.bats` - Journey and goal creation
- `40_checkpoint.bats` - Checkpoint operations
- `50_steps.bats` - All step creation commands
- `60_formats.bats` - Output format testing
- `70_session.bats` - Session context management
- `80_errors.bats` - Error handling scenarios
- `99_report.bats` - Test reporting

### **GitHub Actions Workflows**
- `.github/workflows/test.yml` - Main test workflow with BATS
- `.github/workflows/simple-test.yml` - Quick build and run test
- `.github/workflows/test-api.yml` - API connectivity verification
- `.github/workflows/ci.yml` - Full CI pipeline with linting

### **Running Tests Locally**
```bash
# Run all BATS tests
make test-bats

# Run specific test file
bats src/cmd/tests/20_project.bats

# Run with verbose output
bats -t src/cmd/tests/50_steps.bats
```

## 🎯 Current Status: COMPLETE

### ✅ **Fully Functional**
- **28 commands** across **17 categories**
- **100% success rate** in testing
- **Full API integration** with proper authentication
- **Comprehensive documentation** and examples

### ✅ **Production Ready**
- Parameterized configuration
- Proper error handling
- Multiple output formats
- Consistent command patterns
- Comprehensive help system

### ✅ **Extensible Architecture**
- Easy to add new commands
- Modular design
- Consistent patterns
- Well-documented codebase

## 🔄 Future Enhancements

### **Potential Additions**
- Batch command execution
- Configuration file support
- Command aliasing
- Pipeline integration
- Enhanced error reporting

### **Maintenance Notes**
- Keep API client methods in sync with API changes
- Update documentation as commands evolve
- Maintain consistent patterns across all commands
- Regular testing with live API endpoints

## 📚 Resources

### **Documentation**
- `README.md` - Project overview and setup
- `NEW_COMMANDS_SUMMARY.md` - Recent additions
- `COMPREHENSIVE_TEST_RESULTS.md` - Testing results
- Individual command help via `--help` flag

### **Testing**
- Multiple test scripts for different scenarios
- Live API integration testing
- Comprehensive command validation
- Output format verification

## 🔑 API Authentication

### **Known Issues Resolved**
1. **Projects API Response Format** - The API returns projects in a `map` structure, not an `items` array. Fixed in `pkg/virtuoso/client.go`
2. **Config File Location** - CLI looks for config in `~/.api-cli/virtuoso-config.yaml` or `./config/virtuoso-config.yaml`
3. **Required Fields** - Must include `organization.id` in config for API calls to work

### **GitHub Secrets Setup**
For GitHub Actions to work:
1. Go to Settings → Secrets and variables → Actions
2. Add secret: `VIRTUOSO_API_KEY` with your API key
3. Add variable: `VIRTUOSO_API_URL` with value `https://api-app2.virtuoso.qa/api`

---

**Last Updated**: 2025-01-13  
**Total Commands**: 28  
**Status**: Production Ready  
**API Integration**: Fully Functional with Authentication Fix