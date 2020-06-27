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
	servers []string = []string{
		"kms.srv.crsoo.com",
		"cy2617.jios.org",
		"kms.digiboy.ir",
		"kms.cangshui.net",
		"kms.library.hk",
		"hq1.chinancce.com",
		"kms.loli.beer",
		"kms.v0v.bid",
		"54.223.212.31",
		"kms.jm33.me",
		"nb.shenqw.win",
		"kms.izetn.cn",
		"kms.cin.ink",
		"222.184.9.98",
		"kms.ijio.net",
		"fourdeltaone.net:1688",
		"kms.iaini.net",
		"kms.cnlic.com",
		"kms.51it.wang",
		"key.17108.com",
		"kms.chinancce.com",
		"kms.ddns.net",
		"windows.kms.app",
		"kms.ddz.red",
		"franklv.ddns.net",
		"kms.mogeko.me",
		"k.zpale.com",
		"amrice.top",
		"m.zpale.com",
		"mvg.zpale.com",
		"kms.shuax.com",
		"kensol263.imwork.net:1688",
		"xykz.f3322.org",
		"kms789.com",
		"dimanyakms.sytes.net:1688",
		"kms8.MSGuides.com",
		"kms.03k.org:1688",
		"kms.ymgblog.com",
		"kms.bige0.com",
		"kms9.MSGuides.com",
		"kms.cz9.cn",
		"kms.lolico.moe",
		"kms.ddddg.cn",
		"kms.zhuxiaole.org",
		"kms.moeclub.org",
		"kms.lotro.cc",
		"zh.us.to",
		"noair.strangled.net:1688",
	}
	// populate with keys for the windows version you want to activate.
	keys map[winVer][]string = map[winVer][]string{
		pro: {},
	}
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
