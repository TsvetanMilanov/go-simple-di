.PHONY: \
	test \
	run-docs-server

test:
	go test -v -cover github.com/TsvetanMilanov/go-simple-di/di

run-docs-server:
	godoc -http=":6060"
