run: build
	@./bin/cache --listenAddr :3000

runfollower1: build
	@./bin/cache --listenAddr :4000 --leaderAddr :3000 --raftDir raftFollower1

runfollower2: build
	@./bin/cache --listenAddr :4001 --leaderAddr :3000 --raftDir raftFollower2

build:
	go build -o bin/cache

test:
	go test -v ./...