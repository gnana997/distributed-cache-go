package main

import (
	"context"
	"flag"
	"fmt"
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

		SendCommand(client)
		time.Sleep(2000 * time.Millisecond)

		val, err := GetCommand(client)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(val))

		client.Close()
		time.Sleep(2 * time.Second)
	}()

	server := NewServer(opts, cache.NewCache())

	server.Start()
}

func SendCommand(c *client.Client) {

	err := c.Set(context.Background(), []byte("gg"), []byte("gnana997"), 0)
	if err != nil {
		log.Fatal(err)
	}
}

func GetCommand(c *client.Client) ([]byte, error) {

	val, err := c.Get(context.Background(), []byte("gg"))
	if err != nil {
		return nil, err
	}

	return val, nil
}
