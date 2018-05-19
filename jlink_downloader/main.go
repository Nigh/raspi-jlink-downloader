package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	//"runtime"
	"syscall"
	"time"

	"strings"

	rpio "github.com/stianeikeland/go-rpio"
)

var cmd bytes.Buffer

//device nrf52832_xxaa
//si 1
//speed 8000
//r
//h
//loadfile xxx.hex
//setpc 0
//g
//q
var hexfile string = "/dev/shm/firmware.hex"
var cmdfile string = "download.jlink"

func createRamFolder() {
	// 建立tmp目录，若存在则忽略
	//exe = exec.Command("mkdir", "-p /tmpjlink")
	//	exe.Output()
	// 将tmp目录挂载到ram中，并分配16m的空间
	//	exe = exec.Command("mount", "-t ramfs -o size=16m ramfs /tmpjlink")
	//	exe.Output()
}

func decHex() {
	Decrypt("/home/pi/firmware.enc", hexfile)
	//	exe := exec.Command("mv", "firmware.hex /home/pi/tmp")
	//	exe.Output()
}

func init() {
	// runtime.GOMAXPROCS(1)
	rand.Seed(time.Now().UnixNano())
	cmd.WriteString("device nrf52832_xxaa\n")
	cmd.WriteString("si 1\n")
	cmd.WriteString("speed 8000\n")
	cmd.WriteString("r\nh\n")
	cmd.WriteString("loadfile " + hexfile + "\n")
	cmd.WriteString("setpc 0\n")
	cmd.WriteString("g\nq\n")
	if checkFileIsExist(cmdfile) == true {
		os.Remove(cmdfile)
	}
	f, _ := os.Create(cmdfile) //创建文件
	io.WriteString(f, cmd.String())
	f.Close()
	createRamFolder()
	decHex()
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

var button rpio.Pin
var button2 rpio.Pin

var R rpio.Pin
var G rpio.Pin
var B rpio.Pin
var Buzzer rpio.Pin

func main() {
	fmt.Printf("Kurumi Programming System Initial...")
	if rpio.Open() == nil {
		fmt.Println("[OK]")
	} else {
		fmt.Println("[ERROR]")
		return
	}
	defer rpio.Close()
	Buzzer = rpio.Pin(4)
	R = rpio.Pin(23)
	G = rpio.Pin(24)
	B = rpio.Pin(25)

	Buzzer.Output()
	R.Output()
	G.Output()
	B.Output()

	R.Low()
	G.Low()
	B.Low()
	Buzzer.Low()
	<-time.After(777 * time.Millisecond)

	Buzzer.High()
	R.High()
	G.High()
	B.High()

	heartbeat := time.Tick(1337 * time.Millisecond)
	go func() {
		for {
			<-heartbeat
			B.Toggle()
		}
	}()

	button = rpio.Pin(27)
	button.Input()
	button.PullUp()
	button2 = rpio.Pin(16)
	button2.Input()
	button2.PullUp()

	var keyevent = make(chan string)

	go keyscan(keyevent)

	go eventhandler(keyevent)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func download() {
	G.High()
	R.High()
	var esc = make(chan bool)
	go func(esc chan bool) {
		for {
			tick := time.After(100 * time.Millisecond)
			select {
			case <-tick:
				B.Toggle()
			case <-esc:
				return
			}
		}
	}(esc)
	exe := exec.Command("JLinkExe", cmdfile)
	//	err := exe.Run()
	//	if err != nil {
	//		fmt.Print(err)
	//	}
	out, err := exe.Output()
	if err != nil {
		fmt.Println(err)
	}
	esc <- false
	fmt.Println(string(out))
	if strings.Contains(string(out), "Verifying") {
		fmt.Println("[Success]")
		go func() {
			G.Low()
			Buzzer.Low()
			<-time.After(400 * time.Millisecond)
			Buzzer.High()
		}()
	} else {
		fmt.Println("[Failure]")
		go func() {
			R.Low()
			for j := 1; j <= 7; j++ {
				Buzzer.Low()
				<-time.After(100 * time.Millisecond)
				Buzzer.High()
				<-time.After(100 * time.Millisecond)
			}
		}()
	}
}

func shutdown() {
	exe := exec.Command("shutdown", "now")
	exe.Run()
}

func keyscan(event chan string) {
	var state rpio.State = button.Read()
	var state2 rpio.State = button2.Read()
	var t int = 0
	scanner := time.NewTicker(40 * time.Millisecond)
	for {
		select {
		case <-scanner.C:
			temp := button.Read()
			if temp == 0 {
				t += 1
				if t == 70 {
					go func() {
						for j := 1; j <= 3; j++ {
							Buzzer.Low()
							<-time.After(100 * time.Millisecond)
							Buzzer.High()
							R.High()
							G.High()
							B.High()
							<-time.After(100 * time.Millisecond)
							R.Low()
							G.Low()
							B.Low()
						}
					}()
				}
			}
			if state != temp {
				state = temp
				if state == 1 {
					if t < 25 {
						event <- "B1 RisingEdge"
					}
					if t > 70 {
						event <- "B1 Longpress"
					}
				} else {
					t = 0
					event <- "B1 FallingEdge"
				}
			}
			temp = button2.Read()
			if state2 != temp {
				state2 = temp
				if state2 == 1 {
					event <- "B2 RisingEdge"
				} else {
					event <- "B2 FallingEdge"
				}
			}
		}
	}
}

func eventhandler(event chan string) {
	for {
		select {
		case str := <-event:
			fmt.Println(str)
			if str == "B1 RisingEdge" {
				download()
			} else if str == "B1 Longpress" {
				shutdown()
			}
		}
	}
}
