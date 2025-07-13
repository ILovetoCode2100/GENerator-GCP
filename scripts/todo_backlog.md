# TODO/FIXME Backlog

## Remaining TODOs After Commit

### High Priority
1. **src/client/client.go**:
   - Line 10: Uncomment import when generated API is available
   - Line 16: Uncomment apiClient field when API is available
   - Line 71-75: Create the generated client when API is available
   - Line 86-89: Uncomment GetAPIClient method
   - Line 107-109: Handle specific error types from generated API

2. **src/cmd/main.go**:
   - Line 142: Mouse commands enhancement completed ✓

### Medium Priority
3. **Documentation**:
   - docs/guides/usage.md:123 - Update usage examples

4. **Test Scripts**:
   - test-new-commands.sh: Various test improvements needed

### Low Priority
5. **Build Scripts**:
   - Various shell scripts may need cleanup and consolidation

## Resolution Strategy
1. Wait for API generator to produce client code
2. Uncomment and integrate generated client
3. Update documentation with real examples
4. Consolidate test scripts
5. Clean up build automation

## Notes
- Most TODOs are related to pending API client generation
- Command implementations are complete and tested
- Focus should be on integration once API is ready
