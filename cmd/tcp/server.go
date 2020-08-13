package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	// 新用户到来，通过该channel登记
	enteringChannel = make(chan *User)

	// 用户离开，通过该channel进行登记
	leavingChannel = make(chan *User)

	// 广播专用的用户消息的channel，缓冲是尽可能避免异常情况阻塞，这里简单给了8
	// 具体值根据情况调整
	messageChannel = make(chan string, 8)
)

func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	// 用于记录聊天室用户，并进行消息广播
	// 1. 新用户进来
	// 2. 用户普通消息
	// 3. 用户离开
	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// 1. 新用户进来，构建该用户实例
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}

	// 2. 由于当前是在一个新的goroutine中进行读操作，所以需要开一个goroutine，用于写操作
	// 读写goroutine之间可以通过channel进行通信
	go sendMessage(conn, user.MessageChannel)

	// 3. 给当前用户发送欢迎消息，向所有用户告知新用户到来
	user.MessageChannel <- "Welcome, " + user.String()
	messageChannel <- "user:`" + strconv.Itoa(user.ID) + "`has enter"

	// 4. 记录到全局表中，避免用锁
	enteringChannel <- user

	// 5. 循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messageChannel <- strconv.Itoa(user.ID) + "" + input.Text()
	}

	if err := input.Err(); err != nil {
		log.Println("读取错误：", err)
	}

	// 6. 用户离开
	leavingChannel <- user
	messageChannel <- "user: `" + strconv.Itoa(user.ID) + "` has left"
}

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}

func (u *User) String() string {
	return u.Addr + ", UID:" + strconv.Itoa(u.ID) + ", Enter At:" +
		u.EnterAt.Format("2006-01-02 15:04:05+8000")
}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, _ = fmt.Fprintln(conn, msg)
	}
}

func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {

		// 用户进入
		case user := <-enteringChannel:
			users[user] = struct{}{}

		// 用户离开
		case user := <-leavingChannel:
			delete(users, user)

			// 避免goroutine泄漏问题
			close(user.MessageChannel)

		// 给所有在线用户发消息
		case msg := <-messageChannel:
			for user := range users {
				user.MessageChannel <- msg
			}
		}

	}
}

// 生成用户 ID
var (
	globalID int
	idLocker sync.Mutex
)

func GenUserID() int {
	idLocker.Lock()
	defer idLocker.Unlock()

	globalID++
	return globalID
}
