package main

import (
	"fmt"
	"github.com/icodeface/grdp"
)

func main() {
	client := grdp.NewClient("192.168.0.3:3889")
	err := client.Login("Administrator", "123456")
	if err != nil {
		fmt.Println("login failed", err)
	} else {
		fmt.Println("login success")
	}
}
