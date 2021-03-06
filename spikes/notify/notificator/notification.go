package notificator

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Options struct {
	DefaultIcon string
	AppName     string
}

const (
	UR_NORMAL   = "normal"
	UR_CRITICAL = "critical"
)

type notifier interface {
	push(title string, text string, sound bool, iconPath string) *exec.Cmd
	pushCritical(title string, text string, sound bool, iconPath string) *exec.Cmd
}

type Notificator struct {
	notifier    notifier
	defaultIcon string
}

func (n Notificator) Push(title string, text string, sound bool, iconPath string, urgency string) error {
	icon := n.defaultIcon

	if iconPath != "" {
		icon = iconPath
	}

	if urgency == UR_CRITICAL {
		return n.notifier.pushCritical(title, text, sound, icon).Run()
	}

	return n.notifier.push(title, text, sound, icon).Run()

}

type osxNotificator struct {
	AppName string
}

func (o osxNotificator) push(title string, text string, sound bool, iconPath string) *exec.Cmd {

	// Checks if terminal-notifier exists, and is accessible.

	term_notif := CheckTermNotif()
	os_version_check := CheckMacOSVersion()

	// String sound argument set if sound boolean is true
	soundArg := ""

	// if terminal-notifier exists, use it.
	// else, fall back to osascript. (Mavericks and later.)
	if term_notif == true {
		// if sound true, set argument
		if sound {
			soundArg += "\"default\""
		}
		return exec.Command("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-sound", soundArg)
	} else if os_version_check == true {
		// if sound true, set argument
		if sound {
			soundArg += "sound name \"beep\""
		}
		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.AppName, title, soundArg)
		return exec.Command("osascript", "-e", notification)
	}

	// if sound true, set argument
	if sound {
		return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title)
	}

	// finally falls back to growlnotify.
	return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title)
}

// Causes the notification to stick around until clicked.
func (o osxNotificator) pushCritical(title string, text string, sound bool, iconPath string) *exec.Cmd {

	// same function as above...

	term_notif := CheckTermNotif()
	os_version_check := CheckMacOSVersion()

	if term_notif == true {
		// timeout set to 30 seconds, to show the importance of the notification
		command := fmt.Sprintf("terminal-notifier", "-title", o.AppName, "-message", text, "-subtitle", title, "-timeout", "30")
		if sound {
			return exec.Command(command, "-sound default")
		}
		return exec.Command(command)
	} else if os_version_check == true {
		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.AppName, title)
		return exec.Command("osascript", "-e", notification)
	}
	if sound {
		return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title, "-e", "default")
	}
	return exec.Command("growlnotify", "-n", o.AppName, "--image", iconPath, "-m", title)

}

type linuxNotificator struct{}

func (l linuxNotificator) push(title string, text string, sound bool, iconPath string) *exec.Cmd {
	if sound {
		return exec.Command("notify-send", "-i", iconPath, title, text, "-e", "default")
	}
	return exec.Command("notify-send", "-i", iconPath, title, text)
}

// Causes the notification to stick around until clicked.
func (l linuxNotificator) pushCritical(title string, text string, sound bool, iconPath string) *exec.Cmd {
	if sound {
		return exec.Command("notify-send", "-i", iconPath, title, text, "-u", "critical", "-e", "default")
	}
	return exec.Command("notify-send", "-i", iconPath, title, text, "-u", "critical")
}

type windowsNotificator struct{}

func (w windowsNotificator) push(title string, text string, sound bool, iconPath string) *exec.Cmd {
	if sound {
		return exec.Command("growlnotify", "/i:", iconPath, "/t:", title, text, "-e", "default")
	}
	return exec.Command("growlnotify", "/i:", iconPath, "/t:", title, text)
}

// Causes the notification to stick around until clicked.
func (w windowsNotificator) pushCritical(title string, text string, sound bool, iconPath string) *exec.Cmd {
	if sound {
		return exec.Command("notify-send", "-i", iconPath, title, text, "/s", "true", "/p", "2", "-e", "default")
	}
	return exec.Command("notify-send", "-i", iconPath, title, text, "/s", "true", "/p", "2")
}

func New(o Options) *Notificator {

	var Notifier notifier

	switch runtime.GOOS {

	case "darwin":
		Notifier = osxNotificator{AppName: o.AppName}
	case "linux":
		Notifier = linuxNotificator{}
	case "windows":
		Notifier = windowsNotificator{}

	}

	return &Notificator{notifier: Notifier, defaultIcon: o.DefaultIcon}
}

// Helper function for macOS

func CheckTermNotif() bool {
	// Checks if terminal-notifier exists, and is accessible.

	check_term_notif := exec.Command("which", "terminal-notifier")
	err := check_term_notif.Start()

	if err != nil {
		return false
	} else {
		err = check_term_notif.Wait()
		if err != nil {
			return false
		}
	}
	// no error, so return true. (terminal-notifier exists)
	return true
}

func CheckMacOSVersion() bool {
	// Checks if the version of macOS is 10.9 or Higher (osascript support for notifications.)

	cmd := exec.Command("sw_vers", "-productVersion")
	check, _ := cmd.Output()

	version := strings.Split(string(check), ".")

	// semantic versioning of macOS

	major, _ := strconv.Atoi(version[0])
	minor, _ := strconv.Atoi(version[1])

	if major < 10 {
		return false
	} else if major == 10 && minor < 9 {
		return false
	} else {
		return true
	}
}
