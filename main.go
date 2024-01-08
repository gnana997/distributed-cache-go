package main

import (
	"gnana997/distributed-cache/cache"
	"log"
	"net"
	"time"
)

func main() {

	opts := ServerOpts{
		ListenAddr: ":3000",
		IsLeader:   true,
	}

	i := 0
	for i < 10 {
		go func() {
			time.Sleep(2 * time.Second)
			conn, err := net.Dial("tcp", ":3000")
			if err != nil {
				log.Fatal(err)
			}

			conn.Write([]byte("SET Foo Bar 2500"))
		}()
		i++
	}

	server := NewServer(opts, cache.NewCache())

	server.Start()
}
