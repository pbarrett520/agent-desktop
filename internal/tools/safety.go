// Package tools provides tool implementations for the Agent Desktop agent.
// This file contains command safety checks to prevent dangerous operations.
package tools

import (
	"regexp"
	"strings"
)

// blockedPatterns contains regex patterns for commands that should NEVER execute.
// These are catastrophic/dangerous commands that could cause data loss or system damage.
var blockedPatterns = []string{
	// Unix/Linux destructive commands
	`rm\s+-rf\s+[/~*]`,       // rm -rf /, ~, or *
	`rm\s+-fr\s+[/~*]`,       // rm -fr variant
	`mkfs\.`,                 // mkfs.* (filesystem format)
	`dd\s+if=.*\s+of=/dev/`,  // dd writing to devices
	`chmod\s+-R\s+777\s+/`,   // chmod -R 777 /
	`:\(\)\{.*:\|:.*\}`,      // fork bomb pattern

	// Windows CMD destructive commands
	`del\s+/s\s+/q\s+C:\\`,   // del /s /q C:\
	`format\s+C:`,            // format C:
	`reg\s+delete\s+HKLM`,    // registry delete HKLM

	// PowerShell destructive commands
	`Remove-Item\s+.*-Recurse\s+.*-Force\s+[C:\\/$~]`, // Remove-Item -Recurse -Force C:\ or / or ~
	`Remove-Item\s+.*-Force\s+.*-Recurse\s+[C:\\/$~]`, // Remove-Item -Force -Recurse variant
	`rm\s+.*-r\s+.*-fo\s+[C:\\/$~]`,                   // PowerShell rm -r -fo alias
	`Format-Volume\s+`,                                // PowerShell format volume
	`Clear-Disk\s+`,                                   // PowerShell clear disk
	`Initialize-Disk\s+`,                              // PowerShell initialize disk
	`Remove-Partition\s+`,                             // PowerShell remove partition
	`Set-ExecutionPolicy\s+Unrestricted`,              // Dangerous policy change

	// Remote code execution patterns (cross-platform)
	`curl\s+.*\|\s*sh`,                    // curl piped to sh
	`curl\s+.*\|\s*bash`,                  // curl piped to bash
	`wget\s+.*\|\s*sh`,                    // wget piped to sh
	`wget\s+.*\|\s*bash`,                  // wget piped to bash
	`Invoke-Expression.*Invoke-WebRequest`, // PowerShell IEX(IWR ...) pattern
	`iex.*iwr`,                            // PowerShell IEX(IWR) short form
	`Invoke-Expression.*curl`,             // PowerShell IEX curl
	`Invoke-Expression.*wget`,             // PowerShell IEX wget
	`powershell\s+-enc`,                   // powershell encoded commands
	`powershell\s+-e\s`,                   // powershell -e (short for -EncodedCommand)
	`powershell\.exe\s+-enc`,              // powershell.exe encoded
	`pwsh\s+-enc`,                         // pwsh encoded commands
}

// compiledPatterns holds the compiled regex patterns for efficiency.
var compiledPatterns []*regexp.Regexp

func init() {
	compiledPatterns = make([]*regexp.Regexp, len(blockedPatterns))
	for i, pattern := range blockedPatterns {
		compiledPatterns[i] = regexp.MustCompile("(?i)" + pattern)
	}
}

// CheckCommandSafety checks if a command is safe to execute.
// Returns (true, "") if safe, (false, reason) if blocked.
func CheckCommandSafety(command string) (bool, string) {
	// Normalize whitespace for more reliable matching
	normalized := strings.TrimSpace(command)

	for i, re := range compiledPatterns {
		if re.MatchString(normalized) {
			return false, "Command blocked: matches dangerous pattern '" + blockedPatterns[i] + "'"
		}
	}

	return true, ""
}
