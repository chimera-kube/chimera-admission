run:
	sh -c 'GO111MODULE=on go run main.go'

build:
	sh -c 'CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o chimera-admission'
