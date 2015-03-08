all:
	go build arcd/*.go
	go build -o arcd.bin arcd.go 
	
clean:
	rm -f arcd.bin