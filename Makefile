test:
	mkdir -p ~/.budget/migrations.d/
	cp migrations/*.sql ~/.budget/migrations.d/
	go test -v ./...

fmt:
	gofmt -w */**