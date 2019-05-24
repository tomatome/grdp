# Go Retome Desktop Protocol

SSL standard security
需要
```
reg add "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp" /v UserAuthentication /t REG_DWORD /d 0 /f
```


// https://blog.csdn.net/songbohr/article/details/5309650

## 调用层次：
rdp_--->sec_--->mcs_--->iso_--->tcp_

协议包编解码层次：
rdp_hdr->sec_hdr->mcs_hdr->iso_hdr->data，所有这些指针组成一个STREAM.

## 主过程：
rdp_connect： 
按照调用层次依次调用sec_connect……，
然后调用rdp_send_logon_info发送登录请求验证信息.
    其中rdp_send_logon_info调用sec_init初始化数据包,
    调用sec_send发送数据包，
    根据flags（包含加密标识）调用加密处理逻辑.
然后进入rdp_main_loop循环，调用rdp_recv,根据触发的事件类型做相应处理。
rdp_disconnect，按照调用层次依次调用sec_disconnect……断开。
特殊的，在iso_disconnect中首先调用iso_send_msg(ISO_PDU_DR)发送PDU消息包，
然后再调用tcp_disconnect 断开连接。


