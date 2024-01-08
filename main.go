package main

import (
	"flag"
	"gnana997/distributed-cache/cache"
	"log"
	"log/slog"
	"net"
	"time"
)

func main() {

	var (
		listenAddr = flag.String("listenAddr", ":3000", "listen address of the server")
		leaderAddr = flag.String("leaderAddr", "", "leader address of the server")
	)

	flag.Parse()

	opts := ServerOpts{
		ListenAddr: *listenAddr,
		IsLeader:   true,
		LeaderAddr: *leaderAddr,
	}

	go func() {
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
	}()

	server := NewServer(opts, cache.NewCache())

	server.Start()
}
