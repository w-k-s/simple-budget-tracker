test:
	TEST_MIGRATIONS_DIRECTORY="$(shell pwd)/migrations" go test -v ./...

fmt:
	gofmt -w */**