package main

import (
	"context"
	"fmt"
	"gnana997/distributed-cache/client"
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

type Server struct {
	raft *raft.Raft
}

func main() {
	var (
		cfg            = raft.DefaultConfig()
		fsm            = &raft.MockFSM{}
		logStore       = raft.NewInmemStore()
		snapshoteStore = raft.NewInmemSnapshotStore()
		stableStore    = raft.NewInmemStore()
		timeout        = time.Second * 5
	)

	cfg.LocalID = "gn"

	ips, err := net.LookupIP("localhost")
	if err != nil {
		log.Fatal(err)
	}
	if len(ips) == 0 {
		log.Fatalf("localhost did not resolve to any IPs")
	}
	addr := &net.TCPAddr{IP: ips[0], Port: 4000}

	tr, err := raft.NewTCPTransport(":4000", addr, 10, timeout, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	r, err := raft.NewRaft(cfg, fsm, logStore, stableStore, snapshoteStore, tr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", r)

	select {}
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

		// val, err := GetCommand(client, key)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Println(string(val))

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
