package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"gnana997/distributed-cache/cache"
	"gnana997/distributed-cache/client"
	"io"
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
		for i := 0; i < 1000; i++ {
			go func() {
				client, err := client.NewCLient(":3000", client.Options{})
				if err != nil {
					log.Fatal(err)
				}

				var (
					key   = randomBytes(10)
					value = randomBytes(10)
				)

				SendCommand(client, key, value)
				time.Sleep(20 * time.Millisecond)

				val, err := GetCommand(client, key)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(string(val))

				client.Close()
				time.Sleep(2 * time.Second)
			}()
		}
	}()

	server := NewServer(opts, cache.NewCache())

	server.Start()
}

func randomBytes(n int) []byte {
	buf := make([]byte, n)
	io.ReadFull(rand.Reader, buf)
	return buf
}

func SendCommand(c *client.Client, key, value []byte) {

	err := c.Set(context.Background(), key, value, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func GetCommand(c *client.Client, key []byte) ([]byte, error) {

	val, err := c.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	return val, nil
}
