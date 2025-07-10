# Fix Implementation Strategy

## Commands Requiring Updates

### Legacy Commands to Update (35 total):
- create-step-wait-time
- create-step-wait-element
- create-step-window
- create-step-double-click
- create-step-right-click
- create-step-hover
- create-step-mouse-down
- create-step-mouse-up
- create-step-mouse-move
- create-step-mouse-enter
- create-step-key
- create-step-pick
- create-step-pick-value
- create-step-pick-text
- create-step-upload
- create-step-scroll-top
- create-step-scroll-bottom
- create-step-scroll-element
- create-step-scroll-position
- create-step-assert-matches
- create-step-assert-not-equals
- create-step-store
- create-step-store-value
- create-step-execute-js
- create-step-add-cookie
- create-step-delete-cookie
- create-step-clear-cookies
- create-step-dismiss-alert
- create-step-dismiss-confirm
- create-step-dismiss-prompt
- create-step-switch-iframe
- create-step-switch-next-tab
- create-step-switch-prev-tab
- create-step-switch-parent-frame
- create-step-comment

## Implementation Steps
1. Update command arguments to support optional position
2. Add addCheckpointFlag() to command
3. Use resolveStepContext() for checkpoint/position
4. Update help text and examples
5. Test both legacy and modern syntax
