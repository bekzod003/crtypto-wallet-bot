.SILENT:

run:
	go run main.go

get-dependencies:
	go mod tidy
	go mod vendor