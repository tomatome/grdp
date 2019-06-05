# Golang Retome Desktop Protocol

grdp is a pure Golang implementation of the Microsoft RDP (Remote Desktop Protocol) protocol (**client side authorization only**).

## Status

* [ ] SSL Authentication (soon)
* [ ] NLA Authentication

## Example

```golang
client := grdp.NewClient("192.168.0.2:3389", glog.DEBUG)
err := client.Login("Administrator", "123456")
if err != nil {
    fmt.Println("login failed,", err)
} else {
    fmt.Println("login success")
}
```

## Take ideas from

* [rdpy](https://github.com/citronneur/rdpy)
* [node-rdpjs](https://github.com/citronneur/node-rdpjs)
* [gordp](https://github.com/Madnikulin50/gordp)
