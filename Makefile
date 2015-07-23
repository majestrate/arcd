testnet: 
	go build -o testnet_arcd testnet_main.go
main: arc nacl
	go build -o arcd main.go
clean:
	rm -f arcd
