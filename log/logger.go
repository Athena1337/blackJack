package log

import (
	"fmt"
	. "github.com/logrusorgru/aurora"
	"os"
)

var DEBUG bool

func Fatal(msg string){
	fmt.Println(Bold(Red("[-] "+msg)))
	os.Exit(0)
}

func Error(msg string){
	fmt.Println(Red("[-] "+msg))
}

func Warn(msg string){
	fmt.Println(Yellow("[*] "+msg))
}

func Info(msg string){
	fmt.Println(Blue("[+] "+msg))
}

func Debug(msg string){
	if DEBUG{
		fmt.Println(Green("[+] "+msg))
	}
}
