run: doc-gen
	go run github.com/w-k-s/simple-budget-tracker/cmd/server
	
test:
	mkdir -p ~/.budget/migrations.d/
	cp migrations/*.sql ~/.budget/migrations.d/
	go test -v -coverprofile=coverage.txt -coverpkg=test/...,./... ./...
fmt:
	gofmt -w */**

doc-gen:
	cp ./api/openapiv3.yaml ./assets/swaggerui
	statik -src="$(shell pwd)/assets/swaggerui"
