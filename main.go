package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Не указан аргумент: необходимо передать 'server' или 'client'")
		return
	}

	arg := os.Args[1]

	switch arg {
	case "server":
		runServer()
	case "client":
		runClient()
	default:
		fmt.Println("Неизвестный аргумент. Используйте 'server' или 'client'")
	}
}

func runClient() {
	connection, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("Ошибка соединения:", err)
		return
	}
	defer connection.Close()

	fmt.Print("Введите свой ник: ")
	reader := bufio.NewReader(os.Stdin)
	nick, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения никнейма:", err)
		return
	}
	nick = strings.TrimSpace(nick)

	sendMessage(connection, nick)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go receiveMessage(connection, &waitGroup)

	for {
		fmt.Print("Текст, который отправляем: ")
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения сообщения:", err)
			return
		}
		message = strings.TrimSpace(message)

		if message == "exit" {
			fmt.Println("Выход...")
			break
		}

		sendMessage(connection, message)
	}

	waitGroup.Wait()
}

func sendMessage(connection net.Conn, message string) {
	fmt.Println("Отправляю сообщение:", message)
	_, err := fmt.Fprintf(connection, message+"\n")
	if err != nil {
		fmt.Println("Ошибка отправки сообщения:", err)
	}
}

func receiveMessage(connection net.Conn, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	for {
		message, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения сообщения с сервера:", err)
			return
		}
		fmt.Println("Полученно сообщение:", message)
	}
}

type Client struct {
	connection net.Conn
	nickname   string
}

var clients = make(map[string]*Client)

func handleConnection(connection net.Conn) {
	reader := bufio.NewReader(connection)

	fmt.Println("Новое соединение!")

	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения никнейма пользователя: ", err)
		return
	}

	greetingMessage := "Привет! " + strings.ToUpper(msg)
	connection.Write([]byte(greetingMessage + "\n"))

	client := &Client{connection: connection, nickname: "@" + msg[:len(msg)-1]}

	clients[client.nickname] = client

	fmt.Println(client.nickname)

	go handleMessages(client)
}

func handleMessages(client *Client) {
	defer client.connection.Close()

	reader := bufio.NewReader(client.connection)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Ошибка чтения сообщения:", err)
				break
			}
		}

		message = strings.TrimSpace(message)

		splitedMessage := strings.SplitN(message, " ", 2)
		firstPart := splitedMessage[0]

		if strings.HasPrefix(firstPart, "@") {
			fmt.Printf(firstPart + " - получатель")
			if clientToSend, ok := clients[firstPart]; ok {
				fmt.Println(firstPart + " - получатель")
				clientToSend.connection.Write([]byte(fmt.Sprintf("%s: %s\n", client.nickname, message)))
			}
		} else {
			for _, c := range clients {

				if c.nickname == client.nickname {
					continue
				}

				fmt.Println("______")
				fmt.Println("Пользователь")
				fmt.Println(client.nickname)
				fmt.Println("Отправил сообщение")
				fmt.Println(c.nickname)
				fmt.Println("______")

				c.connection.Write([]byte(fmt.Sprintf("%s: %s\n", client.nickname, message)))
			}
		}
		client.connection.Write([]byte(fmt.Sprintf("%s\n", "ОТПРАВЛЕНО")))
	}

	delete(clients, client.nickname)
	fmt.Printf("User %s disconnected\n", client.nickname)
}

func runServer() {

	fmt.Println("Запуск сервера...")

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Сервер запущен на порту 9000")

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(connection)
	}
}
