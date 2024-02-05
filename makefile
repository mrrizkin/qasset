.PHONY: all info build serve clean

# Print information about available commands
info:
	$(info ------------------------------------------)
	$(info -                 QAsset                 -)
	$(info ------------------------------------------)
	$(info This Makefile helps you manage the project.)
	$(info )
	$(info Available commands:)
	$(info - build: Build QAsset)
	$(info - serve: Run QAsset)
	$(info - clean: Clean build artifacts.)
	$(info )
	$(info Usage: make <command>)

install: clean
	@echo "=== Installing dependencies ==="
	@go mod tidy

all: install build move

serve: clean
	@echo "=== Running Server ==="
	@air -c .air.toml

build: clean
	@echo "=== Building ==="
	@go build -o qasset -v main.go

clean:
	@echo "=== Cleaning build artifacts ==="
	@rm -f qasset

move:
	@echo "=== Moving build artifacts ==="
	@mv qasset /usr/local/bin
	@echo "=== Done ==="
	@echo "=== Instaling service ==="
	@cp qasset.service /etc/systemd/system
	@systemctl enable qasset
	@systemctl start qasset
	@echo "=== Done ==="
