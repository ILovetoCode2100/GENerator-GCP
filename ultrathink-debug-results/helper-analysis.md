# Helper Function Analysis

## Key Helper Functions

### resolveStepContext
Purpose: Resolves checkpoint ID and position from session or args

### addCheckpointFlag
Purpose: Adds --checkpoint flag to commands

### outputStepResult
Purpose: Consistent output formatting across formats

## Available Functions in step_helpers.go:
- resolveStepContext(args []string, checkpointFlag int, positionIndex int) (*StepContext, error) {
- saveStepContext(ctx *StepContext) {
- validateOutputFormat(format string) error {
- outputStepResult(output *StepOutput) error {
- outputStepResultJSON(output *StepOutput) error {
- outputStepResultYAML(output *StepOutput) error {
- outputStepResultAI(output *StepOutput) error {
- outputStepResultHuman(output *StepOutput) error {
- addCheckpointFlag(cmd *cobra.Command, checkpointFlag *int) {
- parseIntArg(arg string, fieldName string) (int, error) {
- enableNegativeNumbers(cmd *cobra.Command) {
