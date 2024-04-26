BINARY_NAME := kessel-plan

# Default target
all: $(BINARY_NAME)

# Rule for building the binary
$(BINARY_NAME): *.go
	@go build -o $(BINARY_NAME) $^

# Clean the workspace
clean:
	@rm -f $(BINARY_NAME)

# Phony targets
.PHONY: all clean