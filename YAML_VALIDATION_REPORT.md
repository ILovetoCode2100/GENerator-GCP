# Virtuoso API CLI YAML Validation Report

## Executive Summary

**Date:** July 24, 2025  
**Total Files Tested:** 52  
**Valid Files:** 1 (1.9%)  
**Invalid Files:** 51 (98.1%)  

The YAML validation testing reveals a significant mismatch between the expected YAML format and the actual test files in the repository. Only one file (`newsletter.yaml`) passes validation out of 52 tested files.

## Test Scope

### Directories Tested
- `examples/` - 4 files
- `test-all-commands/` - 7 files  
- `test-yaml-files/` - 8 files
- `test-yaml-suite/` - 21 files
- `pkg/api-cli/yaml-layer/examples/` - 6 files
- Root directory - 6 files

## Results by Directory

| Directory | Total Files | Valid | Invalid | Success Rate |
|-----------|-------------|-------|---------|--------------|
| examples | 4 | 0 | 4 | 0% |
| test-all-commands | 7 | 0 | 7 | 0% |
| test-yaml-files | 8 | 0 | 8 | 0% |
| test-yaml-suite | 21 | 0 | 21 | 0% |
| yaml-layer | 6 | 0 | 6 | 0% |
| root | 6 | 1 | 5 | 16.7% |

## Error Category Distribution

| Error Type | Count | Percentage |
|------------|-------|------------|
| Missing 'test:' field | 34 | 66.7% |
| YAML parse errors | 15 | 29.4% |
| Compilation errors | 1 | 2.0% |
| Invalid action format | 1 | 2.0% |

## Key Findings

### 1. Format Mismatch

The validator expects a compact YAML format:
```yaml
test: <test name>
nav: <optional starting URL>
data:
  key: value
do:
  - <action>: <target>
```

However, most test files use a different structure:
```yaml
name: <test name>
infrastructure:
  organization_id: "2242"
  project:
    name: "Project Name"
steps:
  - type: <command>
    target: <selector>
```

### 2. Common Issues

#### Missing 'test:' Field (66.7% of errors)
- Files use `name:` instead of `test:`
- This is the most common validation failure

#### YAML Parse Errors (29.4% of errors)
- Incorrect time duration format (e.g., `30000` instead of `"30s"`)
- Complex YAML features not supported (anchors, aliases)
- Indentation issues
- Missing quotes around special characters

#### Examples of Parse Errors:
- `line 348: cannot unmarshal !!int 30000 into time.Duration`
- `line 105: could not find expected ':'`
- `line 25: did not find expected node content`

### 3. Valid File Analysis

Only one file passed validation:
```yaml
# newsletter.yaml
test: test newsletter subscription
nav: /
do:
  - wait: body
  - c: "Start"
  - wait: 1000
  - ch: "Success"
```

This file demonstrates:
- Compact action syntax (`c:` for click, `ch:` for check)
- Simple structure without complex nesting
- Direct action definitions in the `do:` section

### 4. Structural Differences

| Current Files | Expected Format |
|--------------|-----------------|
| `name:` | `test:` |
| `steps:` | `do:` |
| `type: navigate` | `nav:` |
| `target: <selector>` | Direct selector |
| Complex nested structure | Flat action list |

## Specific File Issues

### Test-All-Commands Directory
All 7 files fail due to missing `test:` field. These files use the `name:` field instead.

### Test-YAML-Files Directory
- 3 files have time duration parse errors
- 5 files have missing `test:` field

### Test-YAML-Suite Directory
- Mix of missing `test:` field errors
- Several files attempt to use complex YAML features (anchors/aliases)
- Some files have type mismatch errors

### YAML-Layer Examples
All 6 files have YAML parse errors, suggesting they use advanced YAML features or have syntax issues.

## Recommendations

### 1. Format Standardization
- **Option A**: Update all test files to match the expected compact format
- **Option B**: Modify the validator to accept the current extended format
- **Option C**: Support both formats with a version/format indicator

### 2. Documentation
- Create clear documentation on the expected YAML format
- Provide migration guide from current format to expected format
- Include validated examples in the documentation

### 3. Validation Improvements
- Add more descriptive error messages
- Support common time duration formats (milliseconds as integers)
- Consider supporting YAML anchors and aliases

### 4. Testing Strategy
- Create a set of canonical test files that pass validation
- Add unit tests for the YAML validator itself
- Implement pre-commit hooks to validate YAML files

### 5. Migration Path
If updating files to match expected format:
1. Convert `name:` to `test:`
2. Convert `steps:` to `do:`
3. Simplify action syntax to compact form
4. Remove unsupported YAML features
5. Fix time duration formats

## Conclusion

The current state indicates a fundamental disconnect between the YAML validator's expectations and the actual test file formats in the repository. With only 1.9% of files passing validation, immediate action is needed to either:

1. Update the validator to accept the current file format
2. Migrate all test files to the expected format
3. Implement a dual-format support system

The high percentage of "missing test field" errors (66.7%) suggests this is primarily a naming convention issue that could be resolved relatively easily through systematic updates.