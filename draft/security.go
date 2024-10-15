package main

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"strings"
	"fmt"
	"io"
)

func main() {
    log.SetFlags(log.Lshortfile)

    cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        log.Println(err)
        return
    }

    config := &tls.Config{Certificates: []tls.Certificate{cer}}
    ln, err := tls.Listen("tcp", ":443", config) 
    if err != nil {
        log.Println(err)
        return
    }
    defer ln.Close()

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        go handleConnection(conn)
    }
}

// func handleConnection(conn net.Conn) {
//     defer conn.Close()

// 	l := log.Default()
// 	l.Printf("remote addr: %s", conn.RemoteAddr().String())
// 	l.Printf("local addr: %s", conn.LocalAddr().String())

//     r := bufio.NewReader(conn)
	
// 	var data []byte

// 	_, err := r.Read(data)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	log.Println(data)
// }

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

    log.Printf("Client connected: %s", clientConn.RemoteAddr().String())

    // Чтение первого запроса от клиента
    clientReader := bufio.NewReader(clientConn)
    requestLine, err := clientReader.ReadString('\n')
    if err != nil {
        log.Println("Error reading request:", err)
        return
    }

	log.Print(requestLine)

	if strings.HasPrefix(requestLine, "CONNECT") {
        targetAddress := extractTargetAddress(requestLine)
        if targetAddress == "" {
            log.Println("Invalid CONNECT request")
            return
        }

        log.Printf("Target address: %s", targetAddress)

        // Соединение с целевым сервером
        targetConn, err := net.Dial("tcp", targetAddress)
        if err != nil {
            log.Printf("Failed to connect to target server %s: %v", targetAddress, err)
            return
        }
        defer targetConn.Close()

        // Подтверждаем клиенту, что соединение установлено
        fmt.Fprint(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")

        // Пересылаем данные между клиентом и целевым сервером
        go transfer(clientConn, targetConn) // Клиент -> Сервер
        transfer(targetConn, clientConn)    // Сервер -> Клиент
    } else {
        log.Println("Unsupported request method")
        return
    }
}


// Функция извлечения адреса сервера назначения из строки запроса CONNECT
func extractTargetAddress(requestLine string) string {
    parts := strings.Split(requestLine, " ")
    if len(parts) < 2 {
        return ""
    }
    return parts[1] // Второй элемент - это "example.com:443"
}

// Функция передачи данных между соединениями
func transfer(destination net.Conn, source net.Conn) {
    _, err := io.Copy(destination, source)
    if err != nil {
        log.Println("Error during data transfer:", err)
    }
}