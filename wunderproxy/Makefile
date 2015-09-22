build:
	go get ./...

deploy:
	@GOOS=linux go get ./...
	ws-disco host list app | tail -n +2 | awk '{ print $$5 }' | while read ip; do \
		scp ${GOPATH}/bin/linux_amd64/wunder{proxy,status} ubuntu@$${ip}:.; \
		ssh ubuntu@$${ip} sudo mv /home/ubuntu/wunder{proxy,status} /opt; \
	done

