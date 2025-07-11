.PHONY: all build test test-example clean

all: build

build:
	go build -o vanitycal

test:
	go test -v ./...

test-example: build
	@echo "Generating example output..."
	./vanitycal -config example.toml -output example.ics
	@echo "Generated: example.ics"
	@echo ""
	@echo "First 50 lines of output:"
	@echo "========================"
	@head -50 example.ics
	@echo ""
	@echo "Event summary:"
	@echo "============="
	@echo "Anniversary events: $$(grep -c "Company Founded\|Product Launch\|Series A\|100th Employee" example.ics)"
	@echo "Countdown events: $$(grep -c "Project Deadline\|Conference Talk" example.ics)"
	@echo "Recurring events: $$(grep -c "New Year's Day\|Bastille Day\|Christmas\|Halloween" example.ics)"
	@echo "Total events: $$(grep -c "BEGIN:VEVENT" example.ics)"

clean:
	rm -f vanitycal

help:
	@echo "Available targets:"
	@echo "  make build        - Build the vanitycal binary"
	@echo "  make test         - Run unit tests"
	@echo "  make test-example - Generate example.ics from example.toml"
	@echo "  make clean        - Remove generated files"
	@echo "  make help         - Show this help message"