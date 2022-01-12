package log

import (
	"fmt"
	"github.com/pterm/pterm"
	"os"
)

var DEBUG bool

func Fatal(msg interface{}){
	pterm.Println(pterm.Red(fmt.Sprintf("[x] %v",msg)))
	os.Exit(0)
}

func Error(msg string){
	pterm.Println(pterm.Red("[-] " + msg))
}

func Warn(msg string){
	pterm.Println(pterm.Yellow("[*] " + msg))
}

func Info(msg string){
	pterm.Println(pterm.Blue("[+] " + msg))
}

func Debug(msg string){
	if DEBUG{
		pterm.Println(pterm.Green("[+] " + msg))
	}
}
