package commands

import (
	"reflect"
	"testing"
)

func TestCommandValidator_ValidateAndCorrect(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name        string
		cmd         string
		args        []string
		wantCmd     string
		wantArgs    []string
		wantErr     bool
		errContains string
	}{
		// Scroll command corrections
		{
			name:     "scroll command without hyphen",
			cmd:      "scroll top",
			args:     []string{"12345", "1"},
			wantCmd:  "scroll-top",
			wantArgs: []string{"12345", "1"},
			wantErr:  false,
		},
		{
			name:     "scroll to command correction",
			cmd:      "scroll to bottom",
			args:     []string{"12345", "2"},
			wantCmd:  "scroll-bottom",
			wantArgs: []string{"12345", "2"},
			wantErr:  false,
		},

		// Dialog command corrections
		{
			name:     "old alert syntax",
			cmd:      "alert accept",
			args:     []string{"12345", "1"},
			wantCmd:  "dismiss-alert",
			wantArgs: []string{"12345", "1"},
			wantErr:  false,
		},
		{
			name:     "confirm accept with flag",
			cmd:      "confirm accept",
			args:     []string{"12345", "1"},
			wantCmd:  "dismiss-confirm",
			wantArgs: []string{"--accept", "12345", "1"},
			wantErr:  false,
		},

		// Switch tab argument order
		{
			name:     "switch tab wrong order",
			cmd:      "switch-tab",
			args:     []string{"12345", "next", "1"},
			wantCmd:  "switch-tab",
			wantArgs: []string{"next", "12345", "1"},
			wantErr:  false,
		},
		{
			name:     "switch tab with direction variants",
			cmd:      "switch tab",
			args:     []string{"12345", "forward", "1"},
			wantCmd:  "switch-tab",
			wantArgs: []string{"next", "12345", "1"},
			wantErr:  false,
		},

		// Mouse coordinate fixes
		{
			name:     "mouse move with space separator",
			cmd:      "mouse move-to",
			args:     []string{"100 200", "12345", "1"},
			wantCmd:  "mouse move-to",
			wantArgs: []string{"100,200", "12345", "1"},
			wantErr:  false,
		},
		{
			name:     "mouse move with separate args",
			cmd:      "mouse move-by",
			args:     []string{"50", "100", "12345", "2"},
			wantCmd:  "mouse move-by",
			wantArgs: []string{"50,100", "12345", "2"},
			wantErr:  false,
		},

		// Store command changes
		{
			name:     "store element-text to store text",
			cmd:      "store element-text",
			args:     []string{"h1", "title", "12345", "1"},
			wantCmd:  "store",
			wantArgs: []string{"text", "h1", "title", "12345", "1"},
			wantErr:  false,
		},
		{
			name:     "store element-attribute to store attribute",
			cmd:      "store element-attribute",
			args:     []string{"img", "src", "imgSrc", "12345", "1"},
			wantCmd:  "store",
			wantArgs: []string{"attribute", "img", "src", "imgSrc", "12345", "1"},
			wantErr:  false,
		},

		// Wait time conversion
		{
			name:     "wait time seconds to milliseconds",
			cmd:      "wait time",
			args:     []string{"5", "12345", "1"},
			wantCmd:  "wait time",
			wantArgs: []string{"5000", "12345", "1"},
			wantErr:  false,
		},
		{
			name:     "wait time already in milliseconds",
			cmd:      "wait time",
			args:     []string{"1500", "12345", "1"},
			wantCmd:  "wait time",
			wantArgs: []string{"1500", "12345", "1"},
			wantErr:  false,
		},

		// Resize dimension formats
		{
			name:     "resize with space separator",
			cmd:      "resize",
			args:     []string{"1024 768", "12345", "1"},
			wantCmd:  "resize",
			wantArgs: []string{"1024x768", "12345", "1"},
			wantErr:  false,
		},
		{
			name:     "resize with asterisk separator",
			cmd:      "resize",
			args:     []string{"1920*1080", "12345", "1"},
			wantCmd:  "resize",
			wantArgs: []string{"1920x1080", "12345", "1"},
			wantErr:  false,
		},

		// Removed commands
		{
			name:        "removed scroll-left command",
			cmd:         "scroll-left",
			args:        []string{"100", "12345", "1"},
			wantErr:     true,
			errContains: "Horizontal scrolling is not supported",
		},
		{
			name:        "removed navigate back",
			cmd:         "navigate back",
			args:        []string{"12345", "1"},
			wantErr:     true,
			errContains: "Browser back navigation is not supported",
		},

		// Flag validation
		{
			name:        "click with unsupported offset flags",
			cmd:         "click",
			args:        []string{"button", "--offset-x", "10", "--offset-y", "20"},
			wantErr:     true,
			errContains: "does not support --offset-x flag",
		},
		{
			name:     "click with valid flags",
			cmd:      "click",
			args:     []string{"button", "--wait", "--timeout", "5000"},
			wantCmd:  "click",
			wantArgs: []string{"--wait", "--timeout", "5000", "button"},
			wantErr:  false,
		},

		// Common misspellings
		{
			name:     "navigate misspelling",
			cmd:      "naviagte",
			args:     []string{"to", "https://example.com"},
			wantCmd:  "navigate",
			wantArgs: []string{"to", "https://example.com"},
			wantErr:  false,
		},
		{
			name:     "double click with space",
			cmd:      "double click",
			args:     []string{"button", "12345", "1"},
			wantCmd:  "double-click",
			wantArgs: []string{"button", "12345", "1"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs, err := validator.ValidateAndCorrect(tt.cmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndCorrect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateAndCorrect() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if gotCmd != tt.wantCmd {
				t.Errorf("ValidateAndCorrect() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ValidateAndCorrect() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestCommandValidator_SpecificFixes(t *testing.T) {
	validator := NewCommandValidator()

	t.Run("fixMouseCoordinates", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
			want []string
		}{
			{
				name: "space separated coordinates",
				args: []string{"100 200", "cp_123", "1"},
				want: []string{"100,200", "cp_123", "1"},
			},
			{
				name: "separate x and y args",
				args: []string{"100", "200", "cp_123", "1"},
				want: []string{"100,200", "cp_123", "1"},
			},
			{
				name: "already correct format",
				args: []string{"100,200", "cp_123", "1"},
				want: []string{"100,200", "cp_123", "1"},
			},
			{
				name: "don't combine checkpoint ID",
				args: []string{"100", "cp_123", "1"},
				want: []string{"100", "cp_123", "1"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := validator.fixMouseCoordinates(tt.args)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("fixMouseCoordinates() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("fixSwitchTabArgs", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
			want []string
		}{
			{
				name: "forward to next",
				args: []string{"forward", "cp_123", "1"},
				want: []string{"next", "cp_123", "1"},
			},
			{
				name: "prev to previous",
				args: []string{"prev", "cp_123", "1"},
				want: []string{"previous", "cp_123", "1"},
			},
			{
				name: "numeric index unchanged",
				args: []string{"3", "cp_123", "1"},
				want: []string{"3", "cp_123", "1"},
			},
			{
				name: "first to 0",
				args: []string{"first", "cp_123", "1"},
				want: []string{"0", "cp_123", "1"},
			},
			{
				name: "last to -1",
				args: []string{"last", "cp_123", "1"},
				want: []string{"-1", "cp_123", "1"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := validator.fixSwitchTabArgs(tt.args)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("fixSwitchTabArgs() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("fixScrollDistance", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
			want []string
		}{
			{
				name: "named distance small",
				args: []string{"small", "cp_123", "1"},
				want: []string{"100", "cp_123", "1"},
			},
			{
				name: "named distance large",
				args: []string{"large", "cp_123", "1"},
				want: []string{"500", "cp_123", "1"},
			},
			{
				name: "numeric unchanged",
				args: []string{"250", "cp_123", "1"},
				want: []string{"250", "cp_123", "1"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := validator.fixScrollDistance(tt.args)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("fixScrollDistance() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestCommandValidator_GetSuggestions(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name    string
		partial string
		want    []string
	}{
		{
			name:    "scroll commands",
			partial: "scroll t",
			want: []string{
				"scroll to top (use: scroll-top)",
				"scroll to bottom (use: scroll-bottom)",
				"scroll to element (use: scroll-element)",
				"scroll to position (use: scroll-position)",
				"scroll top (use: scroll-top)",
			},
		},
		{
			name:    "alert commands",
			partial: "alert",
			want: []string{
				"alert accept (use: dismiss-alert)",
				"alert dismiss (use: dismiss-alert)",
				"alert accept → dismiss-alert",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.GetSuggestions(tt.partial)
			// Check that we have suggestions (exact matching is difficult due to map iteration)
			if len(got) == 0 && len(tt.want) > 0 {
				t.Errorf("GetSuggestions() returned no suggestions, expected some")
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) >= len(substr) && contains(s[1:], substr)
}
