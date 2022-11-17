deps:
	go get -v -t -d ./...

setup:
	go install github.com/golang/mock/mockgen@v1.6.0

gen:
	go generate

test-unit:
	go test -race -coverprofile=coverage.out -covermode=atomic
