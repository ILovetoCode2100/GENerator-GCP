package core

import (
	"time"
)

// YAMLTest represents the top-level structure of a test file
type YAMLTest struct {
	Test     string                 `yaml:"test" json:"test"`
	Desc     string                 `yaml:"desc,omitempty" json:"desc,omitempty"`
	Base     string                 `yaml:"base,omitempty" json:"base,omitempty"`
	Nav      string                 `yaml:"nav,omitempty" json:"nav,omitempty"`
	Config   *Config                `yaml:"config,omitempty" json:"config,omitempty"`
	Setup    []Action               `yaml:"setup,omitempty" json:"setup,omitempty"`
	Do       []Action               `yaml:"do" json:"do"`
	Teardown []Action               `yaml:"teardown,omitempty" json:"teardown,omitempty"`
	Data     map[string]interface{} `yaml:"data,omitempty" json:"data,omitempty"`
}

// Config holds test configuration
type Config struct {
	Retry    int           `yaml:"retry,omitempty" json:"retry,omitempty"`
	Timeout  time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	OnError  string        `yaml:"on_error,omitempty" json:"on_error,omitempty"`
	Viewport string        `yaml:"viewport,omitempty" json:"viewport,omitempty"`
}

// Action represents a test action with all possible types
type Action struct {
	// Navigation
	Nav    string `yaml:"nav,omitempty" json:"nav,omitempty"`
	Scroll string `yaml:"scroll,omitempty" json:"scroll,omitempty"`

	// Interactions
	C      interface{} `yaml:"c,omitempty" json:"c,omitempty"` // Click
	T      interface{} `yaml:"t,omitempty" json:"t,omitempty"` // Type
	K      interface{} `yaml:"k,omitempty" json:"k,omitempty"` // Key
	H      string      `yaml:"h,omitempty" json:"h,omitempty"` // Hover
	Select interface{} `yaml:"select,omitempty" json:"select,omitempty"`

	// Assertions
	Ch  interface{} `yaml:"ch,omitempty" json:"ch,omitempty"`   // Check exists
	Nch interface{} `yaml:"nch,omitempty" json:"nch,omitempty"` // Not check
	Eq  interface{} `yaml:"eq,omitempty" json:"eq,omitempty"`   // Equals
	Neq interface{} `yaml:"neq,omitempty" json:"neq,omitempty"` // Not equals
	Gt  interface{} `yaml:"gt,omitempty" json:"gt,omitempty"`   // Greater than
	Lt  interface{} `yaml:"lt,omitempty" json:"lt,omitempty"`   // Less than

	// Data
	Store  interface{} `yaml:"store,omitempty" json:"store,omitempty"`
	Cookie interface{} `yaml:"cookie,omitempty" json:"cookie,omitempty"`

	// Control
	Wait interface{} `yaml:"wait,omitempty" json:"wait,omitempty"`
	If   *IfBlock    `yaml:"if,omitempty" json:"if,omitempty"`
	Loop *LoopBlock  `yaml:"loop,omitempty" json:"loop,omitempty"`
	Run  string      `yaml:"run,omitempty" json:"run,omitempty"`
	JS   string      `yaml:"js,omitempty" json:"js,omitempty"`
	Note string      `yaml:"note,omitempty" json:"note,omitempty"`

	// Advanced
	Dialog string      `yaml:"dialog,omitempty" json:"dialog,omitempty"`
	Window interface{} `yaml:"window,omitempty" json:"window,omitempty"`
	Upload interface{} `yaml:"upload,omitempty" json:"upload,omitempty"`
	Mouse  interface{} `yaml:"mouse,omitempty" json:"mouse,omitempty"`
}

// IfBlock represents conditional execution
type IfBlock struct {
	Cond string   `yaml:"cond" json:"cond"`
	Then []Action `yaml:"then" json:"then"`
	Else []Action `yaml:"else,omitempty" json:"else,omitempty"`
}

// LoopBlock represents loop execution
type LoopBlock struct {
	Over  interface{} `yaml:"over" json:"over"`
	As    string      `yaml:"as,omitempty" json:"as,omitempty"`
	Do    []Action    `yaml:"do" json:"do"`
	Max   int         `yaml:"max,omitempty" json:"max,omitempty"`
	Until string      `yaml:"until,omitempty" json:"until,omitempty"`
}

// ValidationError provides detailed error information
type ValidationError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column,omitempty"`
	Field   string `json:"field"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
	Example string `json:"example,omitempty"`
}

// CompileResult represents the compiled test
type CompileResult struct {
	Steps       []CompiledStep         `json:"steps"`
	Variables   map[string]interface{} `json:"variables"`
	Checkpoints []string               `json:"checkpoints"`
}

// CompiledStep represents a single CLI command
type CompiledStep struct {
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Options     map[string]interface{} `json:"options"`
	Description string                 `json:"description"`
	LineNumber  int                    `json:"line_number"`
}
