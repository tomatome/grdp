package client

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func getHost() string {
	return os.Getenv("RDP_HOST")
}

func getPort() string {
	port := os.Getenv("RDP_POST")
	if port == "" {
		port = "3389"
	}
	return port
}

func getDomain() string {
	return os.Getenv("RDP_DOMAIN")
}

func getUsername() string {
	username := os.Getenv("RDP_USERNAME")
	if username == "" {
		username = "administrator"
	}
	return username
}

func getPassword() string {
	return os.Getenv("RDP_PASSWORD")
}

func TestClientLogin(t *testing.T) {
	c := NewRdpClient(getHost(), getPort(), getDomain(), getUsername(), getPassword())
	err := c.Login(getHost()+":"+getPort(), getUsername(), getPassword(), 800, 600)
	if err != nil {
		fmt.Println("Login:", err)
	}
	c.OnBitmap(func(b []Bitmap) {
		fmt.Println("ready:", b)
	})
	time.Sleep(100 * time.Second)
}

func TestClientConnect(t *testing.T) {
	host := getHost()
	port := getPort()
	domain := getDomain()
	username := getUsername()
	password := getPassword()
	c := NewRdpClient(host, port, domain, username, password)
	err := c.Connect()
	if err != nil {
		panic(err)
	}
	c.OnBitmap(func(bitmaps []Bitmap) {
		fmt.Printf("Ready %d bitmaps\n", len(bitmaps))
		fmt.Println("Ready: ", bitmaps)
	})
	time.Sleep(100 * time.Second)
}
