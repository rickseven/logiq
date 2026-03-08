package cmdintel

import (
	"os/exec"
	"strings"
)

// GetChangedFiles returns a list of files modified in the Git repository
func GetChangedFiles() []string {
	out, err := exec.Command("git", "diff", "--name-only").Output()
	if err != nil {
		return nil
	}

	files := strings.Split(string(out), "\n")
	var cleaned []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			cleaned = append(cleaned, f)
		}
	}

	// Also get untracked/staged files
	out, err = exec.Command("git", "status", "--porcelain").Output()
	if err == nil {
		statusFiles := strings.Split(string(out), "\n")
		for _, f := range statusFiles {
			if len(f) > 3 {
				name := strings.TrimSpace(f[3:])
				if name != "" {
					cleaned = append(cleaned, name)
				}
			}
		}
	}

	return cleaned
}
