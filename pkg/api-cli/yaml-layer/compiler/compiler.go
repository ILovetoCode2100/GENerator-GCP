package compiler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/core"
)

// Compiler transforms YAML tests into executable CLI commands
type Compiler struct {
	baseURL   string
	variables map[string]interface{}
	position  int
}

// NewCompiler creates a new compiler instance
func NewCompiler(baseURL string) *Compiler {
	return &Compiler{
		baseURL:   baseURL,
		variables: make(map[string]interface{}),
		position:  1,
	}
}

// Compile transforms a YAMLTest into CLI commands
func (c *Compiler) Compile(test *core.YAMLTest) (*core.CompileResult, error) {
	result := &core.CompileResult{
		Steps:       []core.CompiledStep{},
		Variables:   make(map[string]interface{}),
		Checkpoints: []string{},
	}

	// Initialize variables from data section
	for k, v := range test.Data {
		c.variables[k] = v
		result.Variables[k] = v
	}

	// Set base URL if specified
	if test.Base != "" {
		c.baseURL = test.Base
	}

	// Compile initial navigation if specified
	if test.Nav != "" {
		navStep := c.compileNavigation(test.Nav, 0)
		result.Steps = append(result.Steps, navStep)
	}

	// Compile setup steps
	if len(test.Setup) > 0 {
		setupSteps, err := c.compileActions(test.Setup, "setup")
		if err != nil {
			return nil, fmt.Errorf("setup compilation failed: %w", err)
		}
		result.Steps = append(result.Steps, setupSteps...)
	}

	// Compile main test steps
	mainSteps, err := c.compileActions(test.Do, "do")
	if err != nil {
		return nil, fmt.Errorf("main steps compilation failed: %w", err)
	}
	result.Steps = append(result.Steps, mainSteps...)

	// Compile teardown steps
	if len(test.Teardown) > 0 {
		teardownSteps, err := c.compileActions(test.Teardown, "teardown")
		if err != nil {
			return nil, fmt.Errorf("teardown compilation failed: %w", err)
		}
		result.Steps = append(result.Steps, teardownSteps...)
	}

	return result, nil
}

// compileActions compiles a list of actions into steps
func (c *Compiler) compileActions(actions []core.Action, section string) ([]core.CompiledStep, error) {
	steps := []core.CompiledStep{}

	for i, action := range actions {
		lineNum := i + 1 // Approximate line number
		actionSteps, err := c.compileAction(&action, section, lineNum)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", section, i, err)
		}
		steps = append(steps, actionSteps...)
	}

	return steps, nil
}

// compileAction compiles a single action into one or more steps
func (c *Compiler) compileAction(action *core.Action, section string, lineNum int) ([]core.CompiledStep, error) {
	steps := []core.CompiledStep{}

	// Navigation
	if action.Nav != "" {
		step := c.compileNavigation(action.Nav, lineNum)
		steps = append(steps, step)
	}

	// Scroll
	if action.Scroll != "" {
		step := c.compileScroll(action.Scroll, lineNum)
		steps = append(steps, step)
	}

	// Click
	if action.C != nil {
		step, err := c.compileClick(action.C, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Type
	if action.T != nil {
		typeSteps, err := c.compileType(action.T, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, typeSteps...)
	}

	// Key
	if action.K != nil {
		step, err := c.compileKey(action.K, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Hover
	if action.H != "" {
		step := c.compileHover(action.H, lineNum)
		steps = append(steps, step)
	}

	// Check exists
	if action.Ch != nil {
		step, err := c.compileCheck(action.Ch, "exists", lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Check not exists
	if action.Nch != nil {
		step, err := c.compileCheck(action.Nch, "not-exists", lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Equals
	if action.Eq != nil {
		step, err := c.compileEquals(action.Eq, false, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Not equals
	if action.Neq != nil {
		step, err := c.compileEquals(action.Neq, true, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Store
	if action.Store != nil {
		step, err := c.compileStore(action.Store, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Wait
	if action.Wait != nil {
		step, err := c.compileWait(action.Wait, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// JavaScript
	if action.JS != "" {
		step := c.compileJS(action.JS, lineNum)
		steps = append(steps, step)
	}

	// Note/Comment
	if action.Note != "" {
		step := c.compileNote(action.Note, lineNum)
		steps = append(steps, step)
	}

	// Dialog
	if action.Dialog != "" {
		step := c.compileDialog(action.Dialog, lineNum)
		steps = append(steps, step)
	}

	// Select
	if action.Select != nil {
		step, err := c.compileSelect(action.Select, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Window
	if action.Window != nil {
		step, err := c.compileWindow(action.Window, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Upload
	if action.Upload != nil {
		step, err := c.compileUpload(action.Upload, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Mouse
	if action.Mouse != nil {
		step, err := c.compileMouse(action.Mouse, lineNum)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	// Control flow
	if action.If != nil {
		// For now, add a comment indicating conditional block
		step := core.CompiledStep{
			Command:     "step-misc",
			Args:        []string{"comment", fmt.Sprintf("IF: %s", action.If.Cond)},
			Description: fmt.Sprintf("Conditional block: %s", action.If.Cond),
			LineNumber:  lineNum,
		}
		steps = append(steps, step)

		// Compile then branch
		thenSteps, err := c.compileActions(action.If.Then, "if.then")
		if err != nil {
			return nil, err
		}
		steps = append(steps, thenSteps...)

		// Compile else branch if present
		if len(action.If.Else) > 0 {
			elseComment := core.CompiledStep{
				Command:     "step-misc",
				Args:        []string{"comment", "ELSE"},
				Description: "Else branch",
				LineNumber:  lineNum,
			}
			steps = append(steps, elseComment)

			elseSteps, err := c.compileActions(action.If.Else, "if.else")
			if err != nil {
				return nil, err
			}
			steps = append(steps, elseSteps...)
		}
	}

	if action.Loop != nil {
		// Add comment for loop
		step := core.CompiledStep{
			Command:     "step-misc",
			Args:        []string{"comment", fmt.Sprintf("LOOP: %v", action.Loop.Over)},
			Description: fmt.Sprintf("Loop over: %v", action.Loop.Over),
			LineNumber:  lineNum,
		}
		steps = append(steps, step)

		// Compile loop body
		loopSteps, err := c.compileActions(action.Loop.Do, "loop.do")
		if err != nil {
			return nil, err
		}
		steps = append(steps, loopSteps...)
	}

	// Check if any actions were compiled
	if len(steps) == 0 {
		// Try to provide helpful error message
		actionJSON, _ := json.Marshal(action)
		return nil, fmt.Errorf("no recognized action in: %s. Valid actions include: nav, c (click), t (type), ch (check), wait, etc.", string(actionJSON))
	}

	// Update position for each step
	for i := range steps {
		if steps[i].Options == nil {
			steps[i].Options = make(map[string]interface{})
		}
		steps[i].Options["position"] = c.position
		c.position++
	}

	return steps, nil
}

// compileNavigation compiles a navigation action
func (c *Compiler) compileNavigation(nav string, lineNum int) core.CompiledStep {
	url := c.expandVariables(nav)

	// Handle relative URLs
	if !strings.HasPrefix(url, "http") && c.baseURL != "" {
		url = c.baseURL + url
	}

	return core.CompiledStep{
		Command:     "step-navigate",
		Args:        []string{"to", url},
		Description: fmt.Sprintf("Navigate to %s", url),
		LineNumber:  lineNum,
	}
}

// compileScroll compiles a scroll action
func (c *Compiler) compileScroll(scroll string, lineNum int) core.CompiledStep {
	scroll = c.expandVariables(scroll)

	// Determine scroll type
	var args []string
	switch scroll {
	case "top":
		args = []string{"top"}
	case "bottom":
		args = []string{"bottom"}
	default:
		if strings.Contains(scroll, ",") {
			// Coordinates: "100,200"
			args = []string{"position", scroll}
		} else if _, err := strconv.Atoi(scroll); err == nil {
			// Vertical position
			args = []string{"position", "0," + scroll}
		} else {
			// Selector
			args = []string{"element", scroll}
		}
	}

	return core.CompiledStep{
		Command:     "step-navigate",
		Args:        append([]string{"scroll"}, args...),
		Description: fmt.Sprintf("Scroll %s", scroll),
		LineNumber:  lineNum,
	}
}

// compileClick compiles a click action
func (c *Compiler) compileClick(click interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := click.(type) {
	case string:
		selector := c.expandVariables(v)
		return core.CompiledStep{
			Command:     "step-interact",
			Args:        []string{"click", selector},
			Description: fmt.Sprintf("Click %s", selector),
			LineNumber:  lineNum,
		}, nil

	case map[interface{}]interface{}:
		// Extract selector and options
		for k, opts := range v {
			selector := c.expandVariables(fmt.Sprintf("%v", k))
			step := core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"click", selector},
				Description: fmt.Sprintf("Click %s", selector),
				LineNumber:  lineNum,
				Options:     make(map[string]interface{}),
			}

			// Process options
			if optsMap, ok := opts.(map[interface{}]interface{}); ok {
				if pos, hasPos := optsMap["pos"]; hasPos {
					step.Options["position_type"] = fmt.Sprintf("%v", pos)
				}
				if varName, hasVar := optsMap["var"]; hasVar {
					c.variables[fmt.Sprintf("%v", varName)] = selector
				}
			}

			return step, nil
		}

	case map[string]interface{}:
		// Extract selector and options (string keys)
		for k, opts := range v {
			selector := c.expandVariables(k)
			step := core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"click", selector},
				Description: fmt.Sprintf("Click %s", selector),
				LineNumber:  lineNum,
				Options:     make(map[string]interface{}),
			}

			// Process options
			if optsMap, ok := opts.(map[string]interface{}); ok {
				if pos, hasPos := optsMap["pos"]; hasPos {
					step.Options["position_type"] = fmt.Sprintf("%v", pos)
				}
				if varName, hasVar := optsMap["var"]; hasVar {
					c.variables[fmt.Sprintf("%v", varName)] = selector
				}
			} else if optsMap, ok := opts.(map[interface{}]interface{}); ok {
				if pos, hasPos := optsMap["pos"]; hasPos {
					step.Options["position_type"] = fmt.Sprintf("%v", pos)
				}
				if varName, hasVar := optsMap["var"]; hasVar {
					c.variables[fmt.Sprintf("%v", varName)] = selector
				}
			}

			return step, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid click format: %T", click)
}

// compileType compiles type actions
func (c *Compiler) compileType(t interface{}, lineNum int) ([]core.CompiledStep, error) {
	steps := []core.CompiledStep{}

	switch v := t.(type) {
	case string:
		// Simple text input - need to determine selector from context
		text := c.expandVariables(v)
		return []core.CompiledStep{{
			Command:     "step-interact",
			Args:        []string{"write", "[focused]", text},
			Description: fmt.Sprintf("Type: %s", text),
			LineNumber:  lineNum,
		}}, nil

	case map[interface{}]interface{}:
		// Map of selector: value
		for k, val := range v {
			selector := c.expandVariables(fmt.Sprintf("%v", k))
			value := c.expandVariables(fmt.Sprintf("%v", val))

			// Clear existing text first
			steps = append(steps, core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"write", selector, ""},
				Description: fmt.Sprintf("Clear %s", selector),
				LineNumber:  lineNum,
			})

			// Type new value
			steps = append(steps, core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"write", selector, value},
				Description: fmt.Sprintf("Type '%s' into %s", value, selector),
				LineNumber:  lineNum,
			})
		}
		return steps, nil

	case map[string]interface{}:
		// Map of selector: value (string keys)
		for k, val := range v {
			selector := c.expandVariables(k)
			value := c.expandVariables(fmt.Sprintf("%v", val))

			// Clear existing text first
			steps = append(steps, core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"write", selector, ""},
				Description: fmt.Sprintf("Clear %s", selector),
				LineNumber:  lineNum,
			})

			// Type new value
			steps = append(steps, core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"write", selector, value},
				Description: fmt.Sprintf("Type '%s' into %s", value, selector),
				LineNumber:  lineNum,
			})
		}
		return steps, nil
	}

	return nil, fmt.Errorf("invalid type format: %T", t)
}

// compileKey compiles keyboard actions
func (c *Compiler) compileKey(key interface{}, lineNum int) (core.CompiledStep, error) {
	keyStr := fmt.Sprintf("%v", key)
	keyStr = c.expandVariables(keyStr)

	// Map common key names
	keyMap := map[string]string{
		"enter":     "Enter",
		"return":    "Enter",
		"esc":       "Escape",
		"escape":    "Escape",
		"tab":       "Tab",
		"space":     "Space",
		"backspace": "Backspace",
		"delete":    "Delete",
		"up":        "ArrowUp",
		"down":      "ArrowDown",
		"left":      "ArrowLeft",
		"right":     "ArrowRight",
	}

	if mapped, ok := keyMap[strings.ToLower(keyStr)]; ok {
		keyStr = mapped
	}

	return core.CompiledStep{
		Command:     "step-interact",
		Args:        []string{"key", keyStr},
		Description: fmt.Sprintf("Press %s", keyStr),
		LineNumber:  lineNum,
	}, nil
}

// compileHover compiles a hover action
func (c *Compiler) compileHover(selector string, lineNum int) core.CompiledStep {
	selector = c.expandVariables(selector)
	return core.CompiledStep{
		Command:     "step-interact",
		Args:        []string{"hover", selector},
		Description: fmt.Sprintf("Hover over %s", selector),
		LineNumber:  lineNum,
	}
}

// compileCheck compiles assertion actions
func (c *Compiler) compileCheck(check interface{}, checkType string, lineNum int) (core.CompiledStep, error) {
	selector := c.expandVariables(fmt.Sprintf("%v", check))

	return core.CompiledStep{
		Command:     "step-assert",
		Args:        []string{checkType, selector},
		Description: fmt.Sprintf("Assert %s: %s", checkType, selector),
		LineNumber:  lineNum,
	}, nil
}

// compileEquals compiles equals/not-equals assertions
func (c *Compiler) compileEquals(eq interface{}, isNot bool, lineNum int) (core.CompiledStep, error) {
	cmdType := "equals"
	if isNot {
		cmdType = "not-equals"
	}

	switch v := eq.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			selector := c.expandVariables(fmt.Sprintf("%v", k))
			expected := c.expandVariables(fmt.Sprintf("%v", val))

			return core.CompiledStep{
				Command:     "step-assert",
				Args:        []string{cmdType, selector, expected},
				Description: fmt.Sprintf("Assert %s %s: %s", selector, cmdType, expected),
				LineNumber:  lineNum,
			}, nil
		}

	case map[string]interface{}:
		for k, val := range v {
			selector := c.expandVariables(k)
			expected := c.expandVariables(fmt.Sprintf("%v", val))

			return core.CompiledStep{
				Command:     "step-assert",
				Args:        []string{cmdType, selector, expected},
				Description: fmt.Sprintf("Assert %s %s: %s", selector, cmdType, expected),
				LineNumber:  lineNum,
			}, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid equals format: %T", eq)
}

// compileStore compiles store actions
func (c *Compiler) compileStore(store interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := store.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			selector := c.expandVariables(fmt.Sprintf("%v", k))
			varName := fmt.Sprintf("%v", val)

			// Store in compiler variables
			c.variables[varName] = ""

			return core.CompiledStep{
				Command:     "step-data",
				Args:        []string{"store", "element-text", selector, varName},
				Description: fmt.Sprintf("Store text from %s as %s", selector, varName),
				LineNumber:  lineNum,
			}, nil
		}

	case map[string]interface{}:
		for k, val := range v {
			selector := c.expandVariables(k)
			varName := fmt.Sprintf("%v", val)

			// Store in compiler variables
			c.variables[varName] = ""

			return core.CompiledStep{
				Command:     "step-data",
				Args:        []string{"store", "element-text", selector, varName},
				Description: fmt.Sprintf("Store text from %s as %s", selector, varName),
				LineNumber:  lineNum,
			}, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid store format: %T", store)
}

// compileWait compiles wait actions
func (c *Compiler) compileWait(wait interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := wait.(type) {
	case int:
		return core.CompiledStep{
			Command:     "step-wait",
			Args:        []string{"time", strconv.Itoa(v)},
			Description: fmt.Sprintf("Wait %dms", v),
			LineNumber:  lineNum,
		}, nil

	case float64:
		return core.CompiledStep{
			Command:     "step-wait",
			Args:        []string{"time", strconv.Itoa(int(v))},
			Description: fmt.Sprintf("Wait %dms", int(v)),
			LineNumber:  lineNum,
		}, nil

	case string:
		selector := c.expandVariables(v)
		return core.CompiledStep{
			Command:     "step-wait",
			Args:        []string{"element", selector},
			Description: fmt.Sprintf("Wait for %s", selector),
			LineNumber:  lineNum,
		}, nil

	case map[interface{}]interface{}:
		// Extended wait syntax
		if forSel, ok := v["for"]; ok {
			selector := c.expandVariables(fmt.Sprintf("%v", forSel))
			step := core.CompiledStep{
				Command:     "step-wait",
				Args:        []string{"element", selector},
				Description: fmt.Sprintf("Wait for %s", selector),
				LineNumber:  lineNum,
				Options:     make(map[string]interface{}),
			}

			if maxWait, ok := v["max"]; ok {
				step.Options["timeout"] = maxWait
			}

			return step, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid wait format: %T", wait)
}

// compileJS compiles JavaScript execution
func (c *Compiler) compileJS(js string, lineNum int) core.CompiledStep {
	js = c.expandVariables(js)
	return core.CompiledStep{
		Command:     "step-misc",
		Args:        []string{"execute", js},
		Description: "Execute JavaScript",
		LineNumber:  lineNum,
	}
}

// compileNote compiles comment/note actions
func (c *Compiler) compileNote(note string, lineNum int) core.CompiledStep {
	note = c.expandVariables(note)
	return core.CompiledStep{
		Command:     "step-misc",
		Args:        []string{"comment", note},
		Description: fmt.Sprintf("Note: %s", note),
		LineNumber:  lineNum,
	}
}

// compileDialog compiles dialog actions
func (c *Compiler) compileDialog(dialog string, lineNum int) core.CompiledStep {
	dialog = strings.ToLower(dialog)

	var args []string
	switch dialog {
	case "accept", "dismiss":
		args = []string{"dismiss-alert"}
	case "confirm":
		args = []string{"dismiss-confirm", "--accept"}
	case "cancel":
		args = []string{"dismiss-confirm", "--reject"}
	default:
		// Assume it's text for prompt
		args = []string{"dismiss-prompt-with-text", dialog}
	}

	return core.CompiledStep{
		Command:     "step-dialog",
		Args:        args,
		Description: fmt.Sprintf("Handle dialog: %s", dialog),
		LineNumber:  lineNum,
	}
}

// compileSelect compiles select actions
func (c *Compiler) compileSelect(sel interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := sel.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			selector := c.expandVariables(fmt.Sprintf("%v", k))
			value := c.expandVariables(fmt.Sprintf("%v", val))

			// Determine select type
			if idx, err := strconv.Atoi(value); err == nil {
				// Select by index
				return core.CompiledStep{
					Command:     "step-interact",
					Args:        []string{"select", "index", selector, strconv.Itoa(idx)},
					Description: fmt.Sprintf("Select index %d in %s", idx, selector),
					LineNumber:  lineNum,
				}, nil
			}

			// Select by text/value
			return core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"select", "option", selector, value},
				Description: fmt.Sprintf("Select '%s' in %s", value, selector),
				LineNumber:  lineNum,
			}, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid select format: %T", sel)
}

// compileWindow compiles window actions
func (c *Compiler) compileWindow(window interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := window.(type) {
	case string:
		v = c.expandVariables(v)

		// Parse window action
		if v == "maximize" {
			return core.CompiledStep{
				Command:     "step-window",
				Args:        []string{"maximize"},
				Description: "Maximize window",
				LineNumber:  lineNum,
			}, nil
		} else if strings.Contains(v, "x") {
			// Resize: "1024x768"
			return core.CompiledStep{
				Command:     "step-window",
				Args:        []string{"resize", v},
				Description: fmt.Sprintf("Resize window to %s", v),
				LineNumber:  lineNum,
			}, nil
		} else if v == "next" || v == "prev" || v == "previous" {
			return core.CompiledStep{
				Command:     "step-window",
				Args:        []string{"switch", "tab", v},
				Description: fmt.Sprintf("Switch to %s tab", v),
				LineNumber:  lineNum,
			}, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid window format: %v", window)
}

// compileUpload compiles file upload actions
func (c *Compiler) compileUpload(upload interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := upload.(type) {
	case map[interface{}]interface{}:
		for k, val := range v {
			selector := c.expandVariables(fmt.Sprintf("%v", k))
			fileURL := c.expandVariables(fmt.Sprintf("%v", val))

			return core.CompiledStep{
				Command:     "step-file",
				Args:        []string{"upload", selector, fileURL},
				Description: fmt.Sprintf("Upload file to %s", selector),
				LineNumber:  lineNum,
			}, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid upload format: %T", upload)
}

// compileMouse compiles mouse actions
func (c *Compiler) compileMouse(mouse interface{}, lineNum int) (core.CompiledStep, error) {
	switch v := mouse.(type) {
	case string:
		v = c.expandVariables(v)

		// Parse mouse action
		if v == "down" {
			return core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"mouse", "down"},
				Description: "Mouse down",
				LineNumber:  lineNum,
			}, nil
		} else if v == "up" {
			return core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"mouse", "up"},
				Description: "Mouse up",
				LineNumber:  lineNum,
			}, nil
		} else if strings.Contains(v, ",") {
			// Coordinates
			return core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"mouse", "move-to", v},
				Description: fmt.Sprintf("Move mouse to %s", v),
				LineNumber:  lineNum,
			}, nil
		}

	case map[interface{}]interface{}:
		// Extended mouse syntax
		if action, ok := v["action"]; ok {
			actionStr := fmt.Sprintf("%v", action)

			if target, hasTarget := v["target"]; hasTarget {
				targetStr := c.expandVariables(fmt.Sprintf("%v", target))
				return core.CompiledStep{
					Command:     "step-interact",
					Args:        []string{"mouse", actionStr, targetStr},
					Description: fmt.Sprintf("Mouse %s %s", actionStr, targetStr),
					LineNumber:  lineNum,
				}, nil
			}

			return core.CompiledStep{
				Command:     "step-interact",
				Args:        []string{"mouse", actionStr},
				Description: fmt.Sprintf("Mouse %s", actionStr),
				LineNumber:  lineNum,
			}, nil
		}
	}

	return core.CompiledStep{}, fmt.Errorf("invalid mouse format: %T", mouse)
}

// expandVariables replaces variable references with their values
func (c *Compiler) expandVariables(s string) string {
	// Replace {{var}} style variables
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)
	s = varPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := varPattern.FindStringSubmatch(match)[1]
		if val, ok := c.variables[varName]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})

	// Replace $var style variables
	dollarPattern := regexp.MustCompile(`\$(\w+)`)
	s = dollarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := dollarPattern.FindStringSubmatch(match)[1]
		if val, ok := c.variables[varName]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})

	// Replace ${ENV:VAR} style environment variables
	envPattern := regexp.MustCompile(`\$\{ENV:(\w+)\}`)
	s = envPattern.ReplaceAllStringFunc(s, func(match string) string {
		// In real implementation, would read from environment
		return match // Keep as-is for now
	})

	return s
}
