all:
	go build nacl/*.go
	go build arc/*.go
	go build -o arcd main.go 
clean:
	rm -f arcd
