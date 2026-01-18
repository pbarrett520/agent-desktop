package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadFile reads the contents of a file.
// If maxLines is provided, it truncates the output to that many lines.
func ReadFile(path string, maxLines *int) ToolResult {
	// Expand path relative to session CWD
	expandedPath := ExpandPath(path, GetSession().CWD)

	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{Success: false, Error: fmt.Sprintf("File not found: %s", expandedPath)}
		}
		return ToolResult{Success: false, Error: err.Error()}
	}

	if info.IsDir() {
		return ToolResult{Success: false, Error: fmt.Sprintf("Not a file: %s", expandedPath)}
	}

	content, err := os.ReadFile(expandedPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	output := string(content)

	if maxLines != nil && *maxLines > 0 {
		lines := strings.Split(output, "\n")
		if len(lines) > *maxLines {
			lines = lines[:*maxLines]
			output = strings.Join(lines, "\n")
			output += fmt.Sprintf("\n... (truncated, showing first %d lines)", *maxLines)
		}
	}

	return ToolResult{Success: true, Output: output}
}

// WriteFile writes content to a file.
// If append is true, it appends to the file instead of overwriting.
// Creates parent directories if they don't exist.
func WriteFile(path string, content string, append bool) ToolResult {
	// Expand path relative to session CWD
	expandedPath := ExpandPath(path, GetSession().CWD)

	// Create parent directories if needed
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create directory: %s", err)}
	}

	var flag int
	if append {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}

	file, err := os.OpenFile(expandedPath, flag, 0644)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	action := "Wrote"
	if append {
		action = "Appended to"
	}

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("%s %s (%d bytes)", action, expandedPath, len(content)),
	}
}

// ListDirectory lists the contents of a directory.
// If showHidden is true, it includes files starting with a dot.
func ListDirectory(path string, showHidden bool) ToolResult {
	// Expand path relative to session CWD
	expandedPath := path
	if path == "" {
		expandedPath = GetSession().CWD
	} else {
		expandedPath = ExpandPath(path, GetSession().CWD)
	}

	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{Success: false, Error: fmt.Sprintf("Directory not found: %s", expandedPath)}
		}
		return ToolResult{Success: false, Error: err.Error()}
	}

	if !info.IsDir() {
		return ToolResult{Success: false, Error: fmt.Sprintf("Not a directory: %s", expandedPath)}
	}

	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Sort entries by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var lines []string
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless requested
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			lines = append(lines, fmt.Sprintf("📁 %s/", name))
		} else {
			info, err := entry.Info()
			if err != nil {
				lines = append(lines, fmt.Sprintf("📄 %s", name))
			} else {
				lines = append(lines, fmt.Sprintf("📄 %s (%s)", name, formatSize(info.Size())))
			}
		}
	}

	output := fmt.Sprintf("Directory: %s\n\n%s", expandedPath, strings.Join(lines, "\n"))
	return ToolResult{Success: true, Output: output}
}

// DeleteFile deletes a file.
// Requires confirm=true to proceed.
func DeleteFile(path string, confirm bool) ToolResult {
	if !confirm {
		return ToolResult{
			Success: false,
			Error:   "Deletion not confirmed. Set confirm=true to delete the file.",
		}
	}

	// Expand path relative to session CWD
	expandedPath := ExpandPath(path, GetSession().CWD)

	info, err := os.Stat(expandedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{Success: false, Error: fmt.Sprintf("File not found: %s", expandedPath)}
		}
		return ToolResult{Success: false, Error: err.Error()}
	}

	if info.IsDir() {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Cannot delete directory with delete_file. Use run_command for directories: %s", expandedPath),
		}
	}

	if err := os.Remove(expandedPath); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Output: fmt.Sprintf("Deleted: %s", expandedPath)}
}

// CopyFile copies a file to a new location.
func CopyFile(source string, destination string) ToolResult {
	// Expand paths relative to session CWD
	srcPath := ExpandPath(source, GetSession().CWD)
	dstPath := ExpandPath(destination, GetSession().CWD)

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{Success: false, Error: fmt.Sprintf("Source file not found: %s", srcPath)}
		}
		return ToolResult{Success: false, Error: err.Error()}
	}

	if srcInfo.IsDir() {
		return ToolResult{Success: false, Error: fmt.Sprintf("Source is not a file: %s", srcPath)}
	}

	// Create parent directories if needed
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create directory: %s", err)}
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Preserve file mode
	os.Chmod(dstPath, srcInfo.Mode())

	return ToolResult{Success: true, Output: fmt.Sprintf("Copied: %s -> %s", srcPath, dstPath)}
}

// MoveFile moves or renames a file.
func MoveFile(source string, destination string) ToolResult {
	// Expand paths relative to session CWD
	srcPath := ExpandPath(source, GetSession().CWD)
	dstPath := ExpandPath(destination, GetSession().CWD)

	if _, err := os.Stat(srcPath); err != nil {
		if os.IsNotExist(err) {
			return ToolResult{Success: false, Error: fmt.Sprintf("Source file not found: %s", srcPath)}
		}
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Create parent directories if needed
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("Failed to create directory: %s", err)}
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	return ToolResult{Success: true, Output: fmt.Sprintf("Moved: %s -> %s", srcPath, dstPath)}
}

// formatSize formats a file size in human-readable form.
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
