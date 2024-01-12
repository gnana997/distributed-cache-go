run: build
	@./bin/cache --listenAddr 127.0.0.1:3000 --raftAddr 127.0.0.1:4000

runfollower1: build
	@./bin/cache --listenAddr 127.0.0.1:4010 --leaderAddr 127.0.0.1:3000 --raftAddr 127.0.0.1:4001 --raftDir raftFollower1

runfollower2: build
	@./bin/cache --listenAddr 127.0.0.1:4011 --leaderAddr 127.0.0.1:3000 --raftAddr 127.0.0.1:4002 --raftDir raftFollower2

build:
	go build -o bin/cache

test:
	go test -v ./...