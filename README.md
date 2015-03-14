# ARCD
## Anonymous Relay Chat Daemon

### status

probably full of bugs, use at own risk

### dependancies

* tor
* go 1.4
* irc client
* a brain

### obtain

    export GOPATH=$HOME/git
    mkdir -p $HOME/git
    go get github.com/majestrate/arcd/arcd
    
### compile

    cd $HOME/git/src/github.com/majestrate/arcd 
    make

### update

    cd $HOME/git/src/github.com/majestrate/arcd
    git pull 
    make 

### useage

join irc server at ::1 port 6667 after running the following command:

    ./arcd.bin

active channels: 

* #arcnet
* #benis (off topic channel)

