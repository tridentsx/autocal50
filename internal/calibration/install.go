package calibration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// InstallResult describes what happened when installing a profile.
type InstallResult struct {
	Installed    bool   `json:"installed"`
	Path         string `json:"path"`         // where the profile was copied
	Instructions string `json:"instructions"` // manual steps if auto-install failed
}

// InstallICCProfile attempts to install the profile into the OS color management
// system. Falls back to instructions if auto-install isn't possible.
func InstallICCProfile(profilePath string) (*InstallResult, error) {
	if _, err := os.Stat(profilePath); err != nil {
		return nil, fmt.Errorf("profile not found: %s", profilePath)
	}

	switch runtime.GOOS {
	case "linux":
		return installLinux(profilePath)
	case "windows":
		return installWindows(profilePath)
	case "darwin":
		return installDarwin(profilePath)
	default:
		return &InstallResult{Instructions: manualInstructions(profilePath)}, nil
	}
}

func installLinux(src string) (*InstallResult, error) {
	// Try colord first (GNOME/KDE).
	if path, err := exec.LookPath("colormgr"); err == nil {
		_ = path
		// Copy to system ICC directory.
		dst := filepath.Join(os.Getenv("HOME"), ".local", "share", "icc", filepath.Base(src))
		os.MkdirAll(filepath.Dir(dst), 0755)
		if err := copyFile(src, dst); err == nil {
			// Import via colormgr.
			exec.Command("colormgr", "import-profile", dst).Run()
			return &InstallResult{Installed: true, Path: dst,
				Instructions: fmt.Sprintf(
					"Profile installed to %s\n\n"+
						"To activate:\n"+
						"1. Open Settings → Color\n"+
						"2. Select your projector display\n"+
						"3. Click 'Add Profile' and select the installed profile\n"+
						"4. Set it as the default profile for that display", dst),
			}, nil
		}
	}

	// Fallback: just copy to ~/.local/share/icc/
	dst := filepath.Join(os.Getenv("HOME"), ".local", "share", "icc", filepath.Base(src))
	os.MkdirAll(filepath.Dir(dst), 0755)
	if err := copyFile(src, dst); err != nil {
		return &InstallResult{Instructions: fmt.Sprintf(linuxManual, src)}, nil
	}
	return &InstallResult{Installed: true, Path: dst,
		Instructions: fmt.Sprintf(
			"Profile copied to %s\n\n"+
				"To activate:\n"+
				"• GNOME: Settings → Color → select display → Add Profile\n"+
				"• KDE: System Settings → Color Management → Add Profile\n"+
				"• Or run: colormgr import-profile %s", dst, dst),
	}, nil
}

func installWindows(src string) (*InstallResult, error) {
	// Windows: copy to system color directory and register.
	sysDir := os.Getenv("SYSTEMROOT")
	if sysDir == "" {
		sysDir = `C:\Windows`
	}
	dst := filepath.Join(sysDir, "System32", "spool", "drivers", "color", filepath.Base(src))

	if err := copyFile(src, dst); err != nil {
		// No admin rights — give manual instructions.
		return &InstallResult{
			Path: src,
			Instructions: fmt.Sprintf(
				"Profile saved to: %s\n\n"+
					"To install:\n"+
					"1. Right-click the .icc file → 'Install Profile'\n"+
					"   (this copies it to Windows\\System32\\spool\\drivers\\color)\n\n"+
					"To activate:\n"+
					"2. Open Settings → System → Display → Advanced display\n"+
					"3. Click 'Display adapter properties'\n"+
					"4. Go to 'Color Management' tab → click 'Color Management...'\n"+
					"5. Select your projector, click 'Add...', choose the profile\n"+
					"6. Click 'Set as Default Profile'\n\n"+
					"For madVR users:\n"+
					"• madVR will auto-detect the profile if it's set as the display default\n"+
					"• Or set it manually in madVR settings → devices → your display → properties", src),
		}, nil
	}

	return &InstallResult{Installed: true, Path: dst,
		Instructions: fmt.Sprintf(
			"Profile installed to %s\n\n"+
				"To activate:\n"+
				"1. Settings → System → Display → Advanced display\n"+
				"2. Display adapter properties → Color Management tab\n"+
				"3. Select your projector → Add → choose the profile\n"+
				"4. Set as Default Profile", dst),
	}, nil
}

func installDarwin(src string) (*InstallResult, error) {
	// macOS: copy to ~/Library/ColorSync/Profiles/
	home, _ := os.UserHomeDir()
	dst := filepath.Join(home, "Library", "ColorSync", "Profiles", filepath.Base(src))
	os.MkdirAll(filepath.Dir(dst), 0755)

	if err := copyFile(src, dst); err != nil {
		return &InstallResult{
			Path: src,
			Instructions: fmt.Sprintf(macManual, src),
		}, nil
	}

	return &InstallResult{Installed: true, Path: dst,
		Instructions: fmt.Sprintf(
			"Profile installed to %s\n\n"+
				"To activate:\n"+
				"1. System Settings → Displays\n"+
				"2. Select your projector display\n"+
				"3. Click the 'Color Profile' dropdown\n"+
				"4. Select '%s'\n\n"+
				"The profile takes effect immediately for all apps.", dst, filepath.Base(src)),
	}, nil
}

const linuxManual = `Could not auto-install the profile.

Manual installation:
1. Copy the file to ~/.local/share/icc/:
   cp "%s" ~/.local/share/icc/

2. Activate it:
   • GNOME: Settings → Color → select your projector → Add Profile
   • KDE: System Settings → Color Management → Add Profile
   • CLI: colormgr import-profile <path>

3. For mpv, add to mpv.conf:
   icc-profile=~/.local/share/icc/<filename>.icc`

const macManual = `Could not auto-install the profile.

Manual installation:
1. Copy the file to ~/Library/ColorSync/Profiles/:
   cp "%s" ~/Library/ColorSync/Profiles/

2. System Settings → Displays → select projector → Color Profile dropdown
   → select the new profile

The profile takes effect immediately for all apps.`

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func manualInstructions(path string) string {
	return fmt.Sprintf("ICC profile saved to: %s\nPlease install it manually in your OS color management settings.", path)
}
