package main

import (
	"fmt"

	"github.com/icodeface/grdp"
	"github.com/icodeface/grdp/glog"
)

func main() {
	//client := grdp.NewClient("192.168.18.101:3389", glog.DEBUG)
	//err := client.Login(".", "administrator", "wren")
	client := grdp.NewClient("192.168.0.132:3389", glog.DEBUG)
	err := client.Login("DEV", "jhadmin", "Letmein123")
	if err != nil {
		fmt.Println("login failed,", err)
	} else {
		fmt.Println("login success")
	}
}
