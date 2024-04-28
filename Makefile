BINARY_NAME := ksl-plan

# Default target
all: $(BINARY_NAME)

# Rule for building the binary
$(BINARY_NAME): *.go empty_bootstrap.yaml
	@go build -o $(BINARY_NAME) *.go

# Clean the workspace
clean:
	@rm -f $(BINARY_NAME)

# Phony targets
.PHONY: all clean