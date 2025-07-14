package shared

// StepRequest represents a request to create a step
type StepRequest struct {
	CheckpointID string                 `json:"checkpointId"`
	Type         string                 `json:"type"`
	Position     int                    `json:"position"`
	Selector     string                 `json:"selector,omitempty"`
	Value        string                 `json:"value,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
}

// StepResult represents the result of creating a step
type StepResult struct {
	ID           string                 `json:"id"`
	CheckpointID string                 `json:"checkpointId"`
	Type         string                 `json:"type"`
	Position     int                    `json:"position"`
	Selector     string                 `json:"selector,omitempty"`
	Value        string                 `json:"value,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
	CreatedAt    string                 `json:"createdAt,omitempty"`
	UpdatedAt    string                 `json:"updatedAt,omitempty"`
}

// InteractionOptions represents options for interaction commands
type InteractionOptions struct {
	Variable    string `json:"variable,omitempty"`
	Target      string `json:"target,omitempty"`
	Position    string `json:"position,omitempty"`
	ElementType string `json:"elementType,omitempty"`
	Modifier    string `json:"modifier,omitempty"`
	Button      string `json:"button,omitempty"`
	ClickCount  int    `json:"clickCount,omitempty"`
}

// NavigationOptions represents options for navigation commands
type NavigationOptions struct {
	NewTab       bool `json:"newTab,omitempty"`
	WaitForLoad  bool `json:"waitForLoad,omitempty"`
	ScrollX      int  `json:"scrollX,omitempty"`
	ScrollY      int  `json:"scrollY,omitempty"`
	ScrollToView bool `json:"scrollToView,omitempty"`
	Smooth       bool `json:"smooth,omitempty"`
}

// AssertionOptions represents options for assertion commands
type AssertionOptions struct {
	Variable      string `json:"variable,omitempty"`
	CaseSensitive bool   `json:"caseSensitive,omitempty"`
	Partial       bool   `json:"partial,omitempty"`
	Regex         bool   `json:"regex,omitempty"`
	Timeout       int    `json:"timeout,omitempty"`
}

// WaitOptions represents options for wait commands
type WaitOptions struct {
	Timeout   int    `json:"timeout,omitempty"`
	Interval  int    `json:"interval,omitempty"`
	Visible   bool   `json:"visible,omitempty"`
	Hidden    bool   `json:"hidden,omitempty"`
	Condition string `json:"condition,omitempty"`
}

// MouseOptions represents options for mouse commands
type MouseOptions struct {
	X          int    `json:"x,omitempty"`
	Y          int    `json:"y,omitempty"`
	OffsetX    int    `json:"offsetX,omitempty"`
	OffsetY    int    `json:"offsetY,omitempty"`
	Button     string `json:"button,omitempty"`
	ClickCount int    `json:"clickCount,omitempty"`
	Duration   int    `json:"duration,omitempty"`
}

// KeyOptions represents options for keyboard commands
type KeyOptions struct {
	Key       string   `json:"key,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	Modifiers []string `json:"modifiers,omitempty"`
	Delay     int      `json:"delay,omitempty"`
	Target    string   `json:"target,omitempty"`
}

// WindowOptions represents options for window commands
type WindowOptions struct {
	Width      int  `json:"width,omitempty"`
	Height     int  `json:"height,omitempty"`
	X          int  `json:"x,omitempty"`
	Y          int  `json:"y,omitempty"`
	Maximize   bool `json:"maximize,omitempty"`
	Minimize   bool `json:"minimize,omitempty"`
	Fullscreen bool `json:"fullscreen,omitempty"`
	TabIndex   int  `json:"tabIndex,omitempty"`
}

// Common step types used by Virtuoso API
const (
	StepTypeNavigate          = "NAVIGATE"
	StepTypeClick             = "CLICK"
	StepTypeDoubleClick       = "DOUBLE_CLICK"
	StepTypeRightClick        = "RIGHT_CLICK"
	StepTypeHover             = "HOVER"
	StepTypeWrite             = "FILL"
	StepTypeKey               = "KEY"
	StepTypeScroll            = "SCROLL"
	StepTypeScrollTop         = "SCROLL_TOP"
	StepTypeScrollBottom      = "SCROLL_BOTTOM"
	StepTypeScrollElement     = "SCROLL_ELEMENT"
	StepTypeScrollPosition    = "SCROLL_POSITION"
	StepTypeWait              = "WAIT"
	StepTypeWaitElement       = "WAIT_FOR_ELEMENT"
	StepTypeAssertExists      = "ASSERT_EXISTS"
	StepTypeAssertNotExists   = "ASSERT_NOT_EXISTS"
	StepTypeAssertEquals      = "ASSERT_TEXT"
	StepTypeAssertNotEquals   = "ASSERT_NOT_TEXT"
	StepTypeAssertContains    = "ASSERT_CONTAINS"
	StepTypeAssertNotContains = "ASSERT_NOT_CONTAINS"
	StepTypeAssertChecked     = "ASSERT_CHECKED"
	StepTypeAssertNotChecked  = "ASSERT_NOT_CHECKED"
	StepTypeAssertSelected    = "ASSERT_SELECTED"
	StepTypeAssertNotSelected = "ASSERT_NOT_SELECTED"
	StepTypeAssertVariable    = "ASSERT_VARIABLE"
	StepTypeAssertGreaterThan = "ASSERT_GREATER_THAN"
	StepTypeAssertLessThan    = "ASSERT_LESS_THAN"
	StepTypeAssertMatches     = "ASSERT_MATCHES"
	StepTypeStore             = "STORE"
	StepTypeStoreText         = "STORE_TEXT"
	StepTypeExecuteScript     = "EXECUTE_SCRIPT"
	StepTypeComment           = "COMMENT"
	StepTypeMouseMove         = "MOUSE_MOVE"
	StepTypeMouseDown         = "MOUSE_DOWN"
	StepTypeMouseUp           = "MOUSE_UP"
	StepTypeWindowResize      = "WINDOW_RESIZE"
	StepTypeSwitchTab         = "SWITCH_TAB"
	StepTypeSwitchFrame       = "SWITCH_FRAME"
	StepTypeSwitchParentFrame = "SWITCH_PARENT_FRAME"
	StepTypeUpload            = "UPLOAD"
	StepTypeCookieCreate      = "COOKIE_CREATE"
	StepTypeCookieDelete      = "COOKIE_DELETE"
	StepTypeCookieWipeAll     = "COOKIE_WIPE_ALL"
	StepTypeAlert             = "ALERT"
	StepTypeConfirm           = "CONFIRM"
	StepTypePrompt            = "PROMPT"
)

// Common element types
const (
	ElementTypeButton   = "BUTTON"
	ElementTypeLink     = "LINK"
	ElementTypeInput    = "INPUT"
	ElementTypeCheckbox = "CHECKBOX"
	ElementTypeRadio    = "RADIO"
	ElementTypeSelect   = "SELECT"
	ElementTypeTextarea = "TEXTAREA"
	ElementTypeDiv      = "DIV"
	ElementTypeSpan     = "SPAN"
	ElementTypeImage    = "IMAGE"
	ElementTypeTable    = "TABLE"
	ElementTypeForm     = "FORM"
)

// Common position values
const (
	PositionCenter      = "CENTER"
	PositionTopLeft     = "TOP_LEFT"
	PositionTopRight    = "TOP_RIGHT"
	PositionBottomLeft  = "BOTTOM_LEFT"
	PositionBottomRight = "BOTTOM_RIGHT"
)

// Common mouse buttons
const (
	MouseButtonLeft   = "left"
	MouseButtonRight  = "right"
	MouseButtonMiddle = "middle"
)

// Common keyboard modifiers
const (
	ModifierCtrl  = "ctrl"
	ModifierShift = "shift"
	ModifierAlt   = "alt"
	ModifierMeta  = "meta"
)

// Common keyboard keys
const (
	KeyEnter      = "Enter"
	KeyTab        = "Tab"
	KeyEscape     = "Escape"
	KeyBackspace  = "Backspace"
	KeyDelete     = "Delete"
	KeyArrowUp    = "ArrowUp"
	KeyArrowDown  = "ArrowDown"
	KeyArrowLeft  = "ArrowLeft"
	KeyArrowRight = "ArrowRight"
	KeyPageUp     = "PageUp"
	KeyPageDown   = "PageDown"
	KeyHome       = "Home"
	KeyEnd        = "End"
	KeySpace      = " "
)
