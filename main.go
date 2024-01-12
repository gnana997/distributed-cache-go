package main

import (
	"flag"
	"gnana997/distributed-cache/cache"
)

func main() {

	var (
		raftAddr   = flag.String("raftAddr", "127.0.0.1:4000", "raft listen address of the server")
		listenAddr = flag.String("listenAddr", "127.0.0.1:3000", "listen address of the server")
		leaderAddr = flag.String("leaderAddr", "", "leader address of the server")
		raftDir    = flag.String("raftDir", "raft", "raft directory for the server")
	)

	flag.Parse()

	opts := ServerOpts{
		RaftListenAddr: *raftAddr,
		ListenAddr:     *listenAddr,
		IsLeader:       len(*leaderAddr) == 0,
		LeaderAddr:     *leaderAddr,
		RaftDir:        *raftDir,
	}

	server := NewServer(opts, cache.NewCache())

	server.Start()
}
