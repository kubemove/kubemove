all: kubemove-cli

kubemove-cli:
	@echo "Building kubemove-cli"
	@rm -rf _output/bin/kubemove
	@go build -o _output/bin/kubemove cmd/kubemove/main.go
	@echo "Done"

clean:
	@echo "Removing old binaries"
	@rm -rf _output
	@echo "Done"