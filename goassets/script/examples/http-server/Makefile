.PHONY: assets

assets:
	rm -f assets.go
	go run ../../goassets.go --file assets.go assets

run: assets
	PORT=3001 go run *.go
