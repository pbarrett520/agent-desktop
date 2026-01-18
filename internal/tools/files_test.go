package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "tools-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

// ReadFile tests

func TestReadFile_Exists(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!\nLine 2\nLine 3"
	os.WriteFile(testFile, []byte(content), 0644)

	result := ReadFile(testFile, nil)

	if !result.Success {
		t.Errorf("ReadFile failed: %s", result.Error)
	}
	if result.Output != content {
		t.Errorf("ReadFile output = %q, want %q", result.Output, content)
	}
}

func TestReadFile_NotExists(t *testing.T) {
	result := ReadFile("/nonexistent/file.txt", nil)

	if result.Success {
		t.Error("ReadFile should fail for nonexistent file")
	}
	if result.Error == "" {
		t.Error("ReadFile should have error message for nonexistent file")
	}
}

func TestReadFile_MaxLines(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test file with many lines
	testFile := filepath.Join(tmpDir, "multiline.txt")
	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	os.WriteFile(testFile, []byte(strings.Join(lines, "\n")), 0644)

	maxLines := 2
	result := ReadFile(testFile, &maxLines)

	if !result.Success {
		t.Errorf("ReadFile failed: %s", result.Error)
	}

	outputLines := strings.Split(result.Output, "\n")
	// Should have 2 lines of content plus truncation notice
	if !strings.Contains(result.Output, "Line 1") {
		t.Error("output should contain Line 1")
	}
	if !strings.Contains(result.Output, "Line 2") {
		t.Error("output should contain Line 2")
	}
	if strings.Contains(result.Output, "Line 3") && !strings.Contains(result.Output, "truncated") {
		t.Error("output should not contain Line 3 (unless in truncation message)")
	}
	if len(outputLines) > 3 && !strings.Contains(result.Output, "truncated") {
		t.Error("output should indicate truncation")
	}
}

// WriteFile tests

func TestWriteFile_Creates(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "newfile.txt")
	content := "New file content"

	result := WriteFile(testFile, content, false)

	if !result.Success {
		t.Errorf("WriteFile failed: %s", result.Error)
	}

	// Verify file was created with correct content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

func TestWriteFile_Overwrites(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "existing.txt")
	os.WriteFile(testFile, []byte("original content"), 0644)

	newContent := "new content"
	result := WriteFile(testFile, newContent, false)

	if !result.Success {
		t.Errorf("WriteFile failed: %s", result.Error)
	}

	data, _ := os.ReadFile(testFile)
	if string(data) != newContent {
		t.Errorf("file content = %q, want %q", string(data), newContent)
	}
}

func TestWriteFile_Appends(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "append.txt")
	os.WriteFile(testFile, []byte("first "), 0644)

	result := WriteFile(testFile, "second", true)

	if !result.Success {
		t.Errorf("WriteFile failed: %s", result.Error)
	}

	data, _ := os.ReadFile(testFile)
	if string(data) != "first second" {
		t.Errorf("file content = %q, want %q", string(data), "first second")
	}
}

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "nested", "dirs", "file.txt")
	content := "nested content"

	result := WriteFile(testFile, content, false)

	if !result.Success {
		t.Errorf("WriteFile failed: %s", result.Error)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read nested file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

// ListDirectory tests

func TestListDirectory_ShowsContents(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create some files and dirs
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	result := ListDirectory(tmpDir, false)

	if !result.Success {
		t.Errorf("ListDirectory failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "file1.txt") {
		t.Error("output should contain file1.txt")
	}
	if !strings.Contains(result.Output, "file2.txt") {
		t.Error("output should contain file2.txt")
	}
	if !strings.Contains(result.Output, "subdir") {
		t.Error("output should contain subdir")
	}
}

func TestListDirectory_HidesHidden(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)

	result := ListDirectory(tmpDir, false)

	if !result.Success {
		t.Errorf("ListDirectory failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "visible.txt") {
		t.Error("output should contain visible.txt")
	}
	if strings.Contains(result.Output, ".hidden") {
		t.Error("output should not contain .hidden when showHidden=false")
	}
}

func TestListDirectory_ShowsHidden(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)

	result := ListDirectory(tmpDir, true)

	if !result.Success {
		t.Errorf("ListDirectory failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "visible.txt") {
		t.Error("output should contain visible.txt")
	}
	if !strings.Contains(result.Output, ".hidden") {
		t.Error("output should contain .hidden when showHidden=true")
	}
}

// DeleteFile tests

func TestDeleteFile_RequiresConfirm(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "todelete.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	result := DeleteFile(testFile, false)

	if result.Success {
		t.Error("DeleteFile should fail without confirm=true")
	}
	if result.Error == "" {
		t.Error("DeleteFile should have error message without confirm")
	}

	// File should still exist
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("file should not be deleted without confirm")
	}
}

func TestDeleteFile_DeletesFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testFile := filepath.Join(tmpDir, "todelete.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	result := DeleteFile(testFile, true)

	if !result.Success {
		t.Errorf("DeleteFile failed: %s", result.Error)
	}

	// File should be deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestDeleteFile_RejectsDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	result := DeleteFile(subDir, true)

	if result.Success {
		t.Error("DeleteFile should fail for directories")
	}
}

// CopyFile tests

func TestCopyFile_CopiesFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")
	content := "copy me"
	os.WriteFile(srcFile, []byte(content), 0644)

	result := CopyFile(srcFile, dstFile)

	if !result.Success {
		t.Errorf("CopyFile failed: %s", result.Error)
	}

	// Both files should exist with same content
	srcData, _ := os.ReadFile(srcFile)
	dstData, _ := os.ReadFile(dstFile)
	if string(srcData) != content {
		t.Error("source file was modified")
	}
	if string(dstData) != content {
		t.Errorf("dest content = %q, want %q", string(dstData), content)
	}
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	result := CopyFile("/nonexistent/file.txt", filepath.Join(tmpDir, "dest.txt"))

	if result.Success {
		t.Error("CopyFile should fail for nonexistent source")
	}
}

// MoveFile tests

func TestMoveFile_MovesFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")
	content := "move me"
	os.WriteFile(srcFile, []byte(content), 0644)

	result := MoveFile(srcFile, dstFile)

	if !result.Success {
		t.Errorf("MoveFile failed: %s", result.Error)
	}

	// Source should not exist
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("source file should not exist after move")
	}

	// Dest should have content
	dstData, _ := os.ReadFile(dstFile)
	if string(dstData) != content {
		t.Errorf("dest content = %q, want %q", string(dstData), content)
	}
}

func TestMoveFile_Renames(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	oldName := filepath.Join(tmpDir, "old.txt")
	newName := filepath.Join(tmpDir, "new.txt")
	content := "rename me"
	os.WriteFile(oldName, []byte(content), 0644)

	result := MoveFile(oldName, newName)

	if !result.Success {
		t.Errorf("MoveFile failed: %s", result.Error)
	}

	// Old name should not exist
	if _, err := os.Stat(oldName); !os.IsNotExist(err) {
		t.Error("old file should not exist after rename")
	}

	// New name should have content
	newData, _ := os.ReadFile(newName)
	if string(newData) != content {
		t.Errorf("new file content = %q, want %q", string(newData), content)
	}
}
