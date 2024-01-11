package main

import (
	"flag"
	"gnana997/distributed-cache/cache"
)

func main() {

	var (
		listenAddr = flag.String("listenAddr", ":3000", "listen address of the server")
		leaderAddr = flag.String("leaderAddr", "", "leader address of the server")
		raftDir    = flag.String("raftDir", "raft", "raft directory for the server")
	)

	flag.Parse()

	opts := ServerOpts{
		ListenAddr: *listenAddr,
		IsLeader:   len(*leaderAddr) == 0,
		LeaderAddr: *leaderAddr,
		RaftDir:    *raftDir,
	}

	server := NewServer(opts, cache.NewCache())

	server.Start()
}
