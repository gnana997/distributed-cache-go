package main

import (
	"log"
	"log/slog"
	"net"
	"time"
)

func main() {
	time.Sleep(2 * time.Second)
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte("SET Foo Bar 2500"))

	time.Sleep(2 * time.Second)

	conn.Write([]byte("GET Foo"))

	buf := make([]byte, 2048)

	n, err := conn.Read(buf)

	if err != nil {
		log.Fatal(err)
	}

	slog.Info("received message from server: ", "msg", string(buf[:n]))

	select {}
}
