package client

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClientLogin(t *testing.T) {
	c := NewClient("192.168.0.132:3389", "administrator", "Jhadmin123", TC_RDP, nil)
	err := c.Login()
	if err != nil {
		fmt.Println("Login:", err)
	}
	c.OnBitmap(func(b []Bitmap) {
		fmt.Println("ready:", b)
	})
	time.Sleep(100 * time.Second)
}

func TestClientConnect(t *testing.T) {
	host := os.Getenv("RDP_HOST")
	port := os.Getenv("RDP_POST")
	if port == "" {
		port = "3389"
	}
	domain := os.Getenv("RDP_DOMAIN")
	username := os.Getenv("RDP_USERNAME")
	password := os.Getenv("RDP_PASSWORD")
	c := NewRdpClient(host, port, domain, username, password)
	err := c.Connect()
	if err != nil {
		panic(err)
	}
}
