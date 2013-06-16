package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	log.SetFlags(log.Lshortfile)
	devs := getProcInputs()
	xinputs := getXinputs()
	printXinputs(xinputs)
	if len(devs) >= 4 {
		twoSeats(devs, xinputs)
	} else {
		oneSeat(devs, xinputs)
	}
}

type procDev struct {
	name, dev, event string
	keep             bool
}

func getProcInputs() []procDev {
	f, err := os.Open("/proc/bus/input/devices")
	if err != nil { log.Fatalln(err) }
	scanner := bufio.NewScanner(f)
	var devs []procDev
	var d procDev
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			if d.keep {
				devs = append(devs, d)
			}
			d.keep = false
			continue
		}
		switch line[0] {
		case 'N':
			f1 := strings.Split(line, "=")
			f2 := strings.Split(f1[1], "\"")
			d.name = strings.TrimSpace(f2[1])
		case 'P':
			if strings.HasSuffix(line, "input0") &&
				!(strings.HasSuffix(d.name, "Button\"") ||
					strings.HasSuffix(d.name, "Speaker\"")) {
				d.keep = true
			}
		case 'H':
			f1 := strings.Split(line, "=")
			f2 := strings.Fields(f1[1])
			d.dev = f2[0]
			if len(f2) > 2 {
				d.dev = f2[1] // sepcial case for icon7
			}
			if len(f2) > 1 && strings.HasPrefix(f2[len(f2)-1], "event") {
				d.event = f2[len(f2)-1]
			} else {
				d.keep = false
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	f.Close()
	return devs
}
func twoSeats(devs []procDev, xinputs []DevType) {
	fmt.Fprintln(os.Stderr, "two seats")
	devMouse := LastDev(devs, "mouse")
	devKbd := LastDev(devs, "kbd")
	fmt.Fprintln(os.Stderr, "Mouse=", devMouse)
	fmt.Fprintln(os.Stderr, "Kbd=", devKbd)
	genScript(devMouse, devKbd, xinputs)
}
func genScript(devMouse, devKbd *procDev, xinputs []DevType) {
	fmt.Println("sudo ls -l /dev/input")
	fmt.Println("# disable on lastMouse & keyboard on master device")
	ms := getXinputByEvent(xinputs, devMouse.event)
	kb := getXinputByEvent(xinputs, devKbd.event)
	geo := "1024x768"
	dsp := ":5"
	fmt.Println("xinput set-prop", ms.id, "141 0")
	fmt.Println("xinput set-prop", kb.id, "141 0")
	fmt.Println("# initial another xserver")
	fmt.Println("sudo Xephyr", dsp, "-audit 0 -ac -screen", geo,
		"-keybd evdev,,device=/dev/input/"+kb.event,
		"-mouse evdev,,device=/dev/input/"+ms.event,
		"-dpi 96 -noreset -nolisten tcp &")
	fmt.Println("sleep 2")
	fmt.Println("export DISPLAY=" + dsp)
	fmt.Println("setxkbmap -model pc105 -layout us" +
		" -variant \"\" -rules evdev -option grp:switch" +
		" -option grp:alt_shift_toggle -option grp_led:caps" +
		" -option terminate:ctrl_alt_bksp")
	fmt.Println("xterm&")
	fmt.Println("wait")
}
func getXinputByEvent(xinputs []DevType, ev string) *DevType {
	for _, master := range xinputs {
		for _, dev := range master.devs {
			if dev.event == ev {
				return &dev
			}
		}
	}
	return nil
}
func LastDev(d []procDev, pattern string) *procDev {
	for i := len(d) - 1; i >= 0; i-- {
		if strings.HasPrefix(d[i].dev, pattern) {
			return &d[i]
		}
	}
	return nil
}
func oneSeat(devs []procDev, xinputs []DevType) {
	fmt.Println("one seat")
}

type DevType struct {
	name      string
	id        string
	isMaster  bool
	isPointer bool
	event     string
	devs      []DevType
}

func (p *DevType) String() string {
	keyOrPointer := "keyboard"
	if p.isPointer {
		keyOrPointer = "pointer"
	}
	return fmt.Sprint(p.id, ". ", keyOrPointer, ",", p.name, ",", p.event)
}

func getXinputs() []DevType {
	var devs []DevType
	out, err := exec.Command("/usr/bin/xinput", "list").CombinedOutput()
	if err != nil { log.Fatal(string(out), "\n", err, "\n") }
	list_long := strings.Split(string(out), "\n")
	out, err = exec.Command("/usr/bin/xinput", "list",
		"--name-only").CombinedOutput()
	if err != nil { log.Fatal(string(out), "\n", err, "\n") }
	list_name := strings.Split(string(out), "\n")
	var devMaster *DevType
	for i, line := range list_long {
		//fmt.Println("dbg:",string(line))
		flds := strings.Fields(string(line))
		if len(flds) < 5 {
			continue
		}
		n := len(flds)
		name := string(list_name[i])
		dev := DevType{name: name,
			id:        strings.Split(flds[n-4], "=")[1],
			isMaster:  (flds[n-3] == "[master"),
			isPointer: (flds[n-2] == "pointer")}
		dev.event = getXinputProps(dev.id)
		if dev.isMaster {
			devs = append(devs, dev)
			devMaster = &(devs[len(devs)-1])
		} else if !(strings.HasPrefix(name, "Virtual") ||
			strings.Contains(name, "XTEST")) {
			devMaster.devs = append(devMaster.devs, dev)
		}
	}
	return devs
}
func getXinputProps(id string) string {
	out, err := exec.Command("/usr/bin/xinput", "--list-props", id).CombinedOutput()
	if err != nil { log.Fatal(string(out), "\n", err, "\n") }
	list := strings.Split(string(out), "\n")
	for _, line := range list {
		//Device Node (259):	"/dev/input/event1"
		if strings.Contains(line, "(259)") {
			return strings.Split(strings.Split(line, "input/")[1], "\"")[0]
		}
	}

	return ""
}
func printXinputs(devs []DevType) {
	for _, dev := range devs {
		if dev.isMaster {
			fmt.Fprint(os.Stderr, dev.String(), "\n")
			for _, subdev := range dev.devs {
				fmt.Fprint(os.Stderr, "  ", subdev.String(), "\n")
			}
		}
	}
}
