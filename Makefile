run: build
	@./bin/cache

runfollower: build
	@./bin/cache --listenAddr :4000 --leaderAddr :3000

build:
	go build -o bin/cache

test:
	go test -v ./...