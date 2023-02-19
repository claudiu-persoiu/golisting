GOOS=linux GOARCH=arm go build -o bin/golisting-linux-arm
GOOS=linux GOARCH=amd64 go build -o bin/golisting-linux-amd64
GOOS=linux GOARCH=386 go build -o bin/golisting-linux-368

GOOS=windows GOARCH=arm go build -o bin/golisting-win-arm
GOOS=windows GOARCH=amd64 go build -o bin/golisting-win-amd64
GOOS=windows GOARCH=386 go build -o bin/golisting-win-368

GOOS=darwin GOARCH=amd64 go build -o bin/golisting-darwin-amd64
GOOS=darwin GOARCH=386 go build -o bin/golisting-darwin-368
