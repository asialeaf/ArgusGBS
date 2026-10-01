.PHONY: cms sms all test

all: cms sms

cms:
	mkdir -p bin
	go build -o bin/arguscms ./cmd/arguscms

sms:
	$(MAKE) -C sms
	mkdir -p bin
	cp sms/argussms bin/argussms

test:
	go test ./...

run-sms:
	./bin/argussms -config configs/argussms.ini

run-cms:
	./bin/arguscms -config configs/arguscms.ini
