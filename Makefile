.PHONY: chimera-admission
chimera-admission: chimera-admission-amd64

.PHONY: phony-explicit

chimera-admission-%: phony-explicit
	sh -c 'CGO_ENABLED=1 GOOS=linux GOARCH=$* GO111MODULE=on go build -o chimera-admission-$*'

.PHONY: run
run:
	sh -c 'GO111MODULE=on go run main.go'
