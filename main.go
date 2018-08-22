package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var (
	// populate with key management servers.
	servers []string
	// populate with keys for the windows version you want to activate.
	keys map[winVer][]string
)

func main() {
	print("Activate Windows 10 All Versions\n", underscore)
	if err := uninstallProductKey(); err != nil {
		infof("%v", errors.Wrap(err, "uninstalling existing product key"))
	}
	if err := clearRegistry(); err != nil {
		infof("%v", errors.Wrap(err, "clearing registry"))
	}
	version, err := version()
	if err != nil {
		failedf("%v", errors.Wrap(err, "detecting windows version"))
	}
	fmt.Printf("Activating for %v version\n", version)
	keylist, ok := keys[version]
	if !ok {
		failedf("no configured keys for %v version", version)
	}
	for _, key := range keylist {
		ok, err := installProductKey(key)
		if err != nil {
			failedf("%v", errors.Wrap(err, "installing product key"))
		}
		if ok {
			break
		}
	}
	var successful bool
	for _, server := range servers {
		if err := setKeyManagementServer(server); err != nil {
			failedf("%v", errors.Wrap(err, "setting key management server"))
		}
		if err := activateWindows(); err == nil {
			successful = true
			fmt.Printf("Activated!\n")
			break
		}
		infof("could not connect to %q: %v, trying another\n", server, err)
	}
	if !successful {
		failedf("activation failed.\n")
	}
}

const (
	underscore = iota
)

func print(msg string, flag int) {
	fmt.Printf(msg)
	if flag == underscore {
		if msg[len(msg)-1] != '\n' {
			fmt.Printf("\n")
		}
		for ii := 0; ii < len(msg)-1; ii++ {
			fmt.Printf("=")
		}
		fmt.Printf("\n")
	}
}

type winVer string

const (
	enterprise winVer = "Enterprise"
	pro               = "Pro"
	home              = "Home"
	unknown           = "Unknown"
)

func version() (winVer, error) {
	out, err := exec.Command(`C:\Windows\System32\wbem\wmic.exe`, "os").CombinedOutput()
	if err != nil {
		return unknown, err
	}
	stdout := string(out)
	switch {
	case strings.Contains(stdout, "enterprise"):
		return enterprise, nil
	case strings.Contains(stdout, "10 Pro"):
		return pro, nil
	case strings.Contains(stdout, "home"):
		return home, nil
	default:
		return unknown, fmt.Errorf("could not determine Windows version")
	}
}

func uninstallProductKey() error {
	out, err := slmgr("/upk")
	if err != nil {
		return errors.Wrap(err, "/upk")
	}
	if !strings.Contains(out, "successfully") {
		return fmt.Errorf(out)
	}
	return nil
}

func clearRegistry() error {
	out, err := slmgr("/cpky")
	if err != nil {
		return errors.Wrap(err, "/cpky")
	}
	if !strings.Contains(out, "successfully") {
		return fmt.Errorf(out)
	}
	return nil
}

func installProductKey(key string) (bool, error) {
	out, err := slmgr("/ipk", key)
	if err != nil {
		return false, errors.Wrapf(err, "/ipk %s", key)
	}
	if !strings.Contains(out, "successfully") {
		return false, nil
	}
	return true, nil
}

func setKeyManagementServer(server string) error {
	out, err := slmgr("/skms", server)
	if err != nil {
		return errors.Wrap(err, "/skms")
	}
	if !strings.Contains(out, "successfully") {
		return fmt.Errorf(out)
	}
	return nil
}

func activateWindows() error {
	out, err := slmgr("/ato")
	if err != nil {
		return errors.Wrap(err, "/ato")
	}
	if !strings.Contains(out, "successfully") {
		return fmt.Errorf("activation failed")
	}
	return nil
}

func slmgr(args ...string) (string, error) {
	cmd := []string{
		"cscript",
		"//nologo",
		`C:\Windows\System32\slmgr.vbs`,
	}
	cmd = append(cmd, args...)
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	return string(out), err
}

func failedf(format string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf("[error] %s", format), v...)
	os.Exit(1)
}

func infof(format string, v ...interface{}) {
	fmt.Printf(fmt.Sprintf("[info] %v", format), v...)
}
