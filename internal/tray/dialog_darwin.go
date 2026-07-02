//go:build darwin

package tray

import (
	"fmt"
	"os/exec"
	"strings"
)

func inputDialog(prompt, defaultText string) (string, bool) {
	script := fmt.Sprintf(`display dialog %q default answer %q with title "flow" buttons {"Cancel","OK"} default button "OK"`, prompt, defaultText)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", false
	}
	s := strings.TrimSpace(string(out))
	idx := strings.Index(s, "text returned:")
	if idx < 0 {
		return "", false
	}
	return strings.TrimSpace(s[idx+len("text returned:"):]), true
}
