package main

import (
	"context"
	"flag"
	"gnana997/distributed-cache/cache"
	"gnana997/distributed-cache/client"
	"log"
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
		IsLeader:   len(*leaderAddr) == 0,
		LeaderAddr: *leaderAddr,
	}

	go func() {
		time.Sleep(2 * time.Second)
		client, err := client.NewCLient(":3000", client.Options{})
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			SendCommand(client)
			time.Sleep(200 * time.Millisecond)
		}
		client.Close()
		time.Sleep(2 * time.Second)
	}()

	server := NewServer(opts, cache.NewCache())

	server.Start()
}

func SendCommand(c *client.Client) {

	_, err := c.Set(context.Background(), []byte("gg"), []byte("gnana997"), 2)
	if err != nil {
		log.Fatal(err)
	}
}
