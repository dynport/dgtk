default:
	go get github.com/dynport/dgtk/cryptostore/cryptostore

kill:
	ps aux | grep -v grep | grep "cryptostore server" | tr -s " " "	" | cut -f 2 | xargs kill
	
restart: kill
	go get github.com/dynport/dgtk/cryptostore/cryptostore
	cryptostore server > ~/logs/cryptostore.log 2>&1 &
