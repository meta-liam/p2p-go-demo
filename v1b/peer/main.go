package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

// go run v1b/peer/main.go
const HAND_SHAKE_MSG = "我是打洞消息"
const serverIP = "localhost" //"207.148.70.129"
const serverPort = 9981

var tag string
var port = 9983 //9982
var i = 0

func main() {
	port = randPort()
	// 当前进程标记字符串,便于显示
	tag = fmt.Sprintf("%d", port)                         //os.Args[1]
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: port} // 注意端口必须固定
	dstAddr := &net.UDPAddr{IP: net.ParseIP(serverIP), Port: serverPort}
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	if _, err = conn.Write([]byte("hello, I'm new peer:" + tag)); err != nil {
		log.Panic(err)
	}
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Printf("error during read: %s", err)
	}
	conn.Close()
	anotherPeer := parseAddr(string(data[:n]))
	fmt.Printf("local:%s server:%s another:%s\n", srcAddr, remoteAddr, anotherPeer.String())
	// 开始打洞
	bidirectionalHole(srcAddr, &anotherPeer)
}

func bidirectionalHole(srcAddr *net.UDPAddr, anotherAddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", srcAddr, anotherAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息

	if _, err = conn.Write([]byte(fmt.Sprintf("%s from %d", HAND_SHAKE_MSG, port))); err != nil {
		log.Println("send handshake:", err)
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			i++
			if _, err = conn.Write([]byte("from [" + tag + "]" + fmt.Sprintf("%s %d", HAND_SHAKE_MSG, i))); err != nil {
				log.Println("send msg fail", err)
			}
		}
	}()
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("error during read: %s\n", err)
		} else {
			log.Printf("收到数据:%s\n", data[:n])
		}
	}
}

func randPort() int {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000) + 9982
	//fmt.Println(r)
	return r
}

func parseAddr(addr string) net.UDPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.UDPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
	}
}
