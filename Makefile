run:
	go run github.com/w-k-s/simple-budget-tracker/cmd/server
	
test:
	mkdir -p ~/.budget/migrations.d/
	cp migrations/*.sql ~/.budget/migrations.d/
	go test -v -coverprofile=coverage.txt -coverpkg=test/...,./... `go list ./...`
fmt:
	gofmt -w */**