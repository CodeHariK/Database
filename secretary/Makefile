test:
	@rm -r SECRETARY
	@go clean -testcache
	@go test -cover -coverprofile=cover.txt ./...

vtest:
	@rm -r SECRETARY
	@go clean -testcache
	@go test -cover -coverprofile=cover.txt -v ./...

cover:
	@rm -r SECRETARY
	@go clean -testcache
	@go test -cover -coverprofile=cover.txt ./...
	@go tool cover -func=cover.txt
	@go tool cover -html=cover.txt

run:
	@go run example/main.go
