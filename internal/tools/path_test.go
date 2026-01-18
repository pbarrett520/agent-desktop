package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExpandPath_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not get home dir: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~", home},
		{"~/Documents", filepath.Join(home, "Documents")},
		{"~/foo/bar", filepath.Join(home, "foo", "bar")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input, "/some/cwd")
			if got != tt.want {
				t.Errorf("ExpandPath(%q, cwd) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandPath_AbsolutePath(t *testing.T) {
	var tests []struct {
		input string
	}

	if runtime.GOOS == "windows" {
		tests = []struct {
			input string
		}{
			{"C:\\Users\\test"},
			{"C:\\Program Files\\app"},
			{"D:\\data"},
		}
	} else {
		tests = []struct {
			input string
		}{
			{"/home/user"},
			{"/var/log"},
			{"/etc/config"},
		}
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input, "/some/cwd")
			if got != tt.input {
				t.Errorf("ExpandPath(%q, cwd) = %q, want %q (absolute path unchanged)", tt.input, got, tt.input)
			}
		})
	}
}

func TestExpandPath_RelativePath(t *testing.T) {
	cwd := filepath.Join(os.TempDir(), "testcwd")

	tests := []struct {
		input string
		want  string
	}{
		{"foo", filepath.Join(cwd, "foo")},
		{"foo/bar", filepath.Join(cwd, "foo", "bar")},
		{"./script.py", filepath.Join(cwd, "script.py")},
		{"../parent", filepath.Join(cwd, "..", "parent")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input, cwd)
			// Normalize for comparison (handles ./ prefix)
			wantNorm := tt.want
			gotNorm := got
			if strings.HasPrefix(tt.input, "./") {
				// For ./ paths, just check it joins correctly
				if !strings.HasSuffix(got, filepath.Base(tt.input[2:])) {
					t.Errorf("ExpandPath(%q, %q) = %q, should end with %q", tt.input, cwd, got, filepath.Base(tt.input[2:]))
				}
				return
			}
			if gotNorm != wantNorm {
				t.Errorf("ExpandPath(%q, %q) = %q, want %q", tt.input, cwd, got, tt.want)
			}
		})
	}
}

func TestExpandPath_EmptyReturnsSession(t *testing.T) {
	cwd := "/test/current/dir"
	got := ExpandPath("", cwd)
	if got != cwd {
		t.Errorf("ExpandPath(\"\", %q) = %q, want %q", cwd, got, cwd)
	}
}

func TestExpandPath_WindowsKnownFolders(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	home, _ := os.UserHomeDir()
	cwd := home

	// Test that Desktop/foo expands (actual path depends on Windows config)
	got := ExpandPath("Desktop", cwd)
	
	// Should either be the actual Desktop path or fallback to home/Desktop
	if got == "" {
		t.Error("ExpandPath(\"Desktop\", cwd) returned empty string")
	}
	
	// Should contain Desktop somewhere in the path
	if !strings.Contains(strings.ToLower(got), "desktop") {
		t.Errorf("ExpandPath(\"Desktop\", cwd) = %q, expected to contain 'desktop'", got)
	}
}

func TestExpandPath_NormalizesSlashes(t *testing.T) {
	home, _ := os.UserHomeDir()
	
	// Test with forward slashes on any OS
	got := ExpandPath("~/foo/bar", "/cwd")
	expected := filepath.Join(home, "foo", "bar")
	
	if got != expected {
		t.Errorf("ExpandPath(\"~/foo/bar\", cwd) = %q, want %q", got, expected)
	}
}
