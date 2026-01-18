package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// getWindowsKnownFolder attempts to get the actual path of a Windows known folder.
// This handles OneDrive redirection where Desktop/Documents may be in OneDrive.
// Returns empty string if not found or not on Windows.
func getWindowsKnownFolder(folderName string) string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// Try to read from registry for accurate paths
	// For simplicity, we'll use environment variables and known paths
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	folderLower := strings.ToLower(folderName)

	// Check OneDrive paths first (common in corporate environments)
	oneDrivePaths := []string{
		os.Getenv("OneDrive"),
		os.Getenv("OneDriveConsumer"),
		os.Getenv("OneDriveCommercial"),
	}

	for _, oneDrive := range oneDrivePaths {
		if oneDrive == "" {
			continue
		}
		candidate := filepath.Join(oneDrive, folderName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	// Fall back to standard home directory paths
	standardPath := filepath.Join(home, folderName)
	if info, err := os.Stat(standardPath); err == nil && info.IsDir() {
		return standardPath
	}

	// For Desktop specifically, try the capitalized version
	if folderLower == "desktop" {
		standardPath = filepath.Join(home, "Desktop")
		if info, err := os.Stat(standardPath); err == nil && info.IsDir() {
			return standardPath
		}
	}

	return ""
}

// ExpandPath expands a path, handling:
// - ~ (home directory)
// - Relative paths (relative to cwd)
// - Windows known folders like Desktop, Documents (handles OneDrive redirection)
func ExpandPath(path string, cwd string) string {
	if path == "" {
		return cwd
	}

	// Handle home directory expansion
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		if path == "~" {
			return home
		}
		// Handle ~/path
		if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
			return filepath.Join(home, path[2:])
		}
	}

	// Normalize path separators for cross-platform handling
	normalized := filepath.FromSlash(path)

	// Handle absolute paths - return as-is
	if filepath.IsAbs(normalized) {
		return normalized
	}

	// Handle ./ prefix
	if strings.HasPrefix(normalized, "."+string(filepath.Separator)) || strings.HasPrefix(path, "./") {
		rest := path
		if strings.HasPrefix(path, "./") {
			rest = path[2:]
		} else {
			rest = normalized[2:]
		}
		return filepath.Join(cwd, rest)
	}

	// Check for Windows known folders at start of relative path
	if runtime.GOOS == "windows" {
		parts := strings.Split(normalized, string(filepath.Separator))
		if len(parts) == 0 {
			parts = strings.Split(path, "/")
		}
		
		firstPart := strings.ToLower(parts[0])
		knownFolders := []string{"desktop", "documents", "downloads"}

		for _, folder := range knownFolders {
			if firstPart == folder {
				knownPath := getWindowsKnownFolder(parts[0])
				if knownPath != "" {
					if len(parts) > 1 {
						return filepath.Join(knownPath, filepath.Join(parts[1:]...))
					}
					return knownPath
				}
				break
			}
		}
	}

	// Otherwise, treat as relative to cwd
	return filepath.Join(cwd, normalized)
}
