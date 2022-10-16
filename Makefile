SRCS=$(wildcard *.go)

rv: $(SRCS)
	go build -o rv *.go

test: rv
	go test ./...

testv: rv
	go test -v ./...

clean:
	rm -f rv

.PHONY: test testv clean
