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

	if opts.IsLeader {
		go func() {
			time.Sleep(10 * time.Second)
			TestClient()
		}()
	}

	server := NewServer(opts, cache.NewCache())

	server.Start()
}

func TestClient() {
	for i := 0; i < 100; i++ {
		go func(i int) {
			client, err := client.NewCLient(":3000", client.Options{})
			if err != nil {
				log.Fatal(err)
			}

			var (
				key   = []byte(fmt.Sprintf("key_%d", i))
				value = []byte(fmt.Sprintf("val_%d", i))
			)

			SendCommand(client, key, value)
			time.Sleep(200 * time.Millisecond)

			val, err := GetCommand(client, key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(val))

			client.Close()
		}(i)
	}
}

// func randomBytes(n int) []byte {
// 	buf := make([]byte, n)
// 	io.ReadFull(rand.Reader, buf)
// 	return buf
// }

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
