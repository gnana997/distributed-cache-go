package main

import (
	"context"
	"fmt"
	"gnana997/distributed-cache/client"
	"log"
	"time"

	"github.com/hashicorp/raft"
)

type Server struct {
	raft *raft.Raft
}

func main() {
	TestClient()
}

func TestClient() {
	client, err := client.NewCLient(":3000", client.Options{})
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		var (
			key   = []byte(fmt.Sprintf("key_%d", i))
			value = []byte(fmt.Sprintf("val_%d", i))
		)

		SendCommand(client, key, value)
		time.Sleep(time.Second)

		val, err := GetCommand(client, key)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(val))

		// client.Close()
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
