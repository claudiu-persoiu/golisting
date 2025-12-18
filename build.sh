GOOS=linux GOARCH=arm go build -o bin/golisting-linux-arm cmd/golisting.go
GOOS=linux GOARCH=amd64 go build -o bin/golisting-linux-amd64 cmd/golisting.go
GOOS=linux GOARCH=386 go build -o bin/golisting-linux-368 cmd/golisting.go

GOOS=windows GOARCH=arm go build -o bin/golisting-win-arm cmd/golisting.go
GOOS=windows GOARCH=amd64 go build -o bin/golisting-win-amd64 cmd/golisting.go
GOOS=windows GOARCH=386 go build -o bin/golisting-win-368 cmd/golisting.go

GOOS=darwin GOARCH=amd64 go build -o bin/golisting-darwin-amd64 cmd/golisting.go
GOOS=darwin GOARCH=arm64 go build -o bin/golisting-darwin-amr64 cmd/golisting.go
