//go:build !darwin && !windows

package tray

// inputDialog is a no-op on unsupported platforms.
// Windows support will be added separately.
func inputDialog(prompt, defaultText string) (string, bool) {
	return "", false
}
