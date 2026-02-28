# Makefile for mini-gcc

# Go compiler
GO := go
# Output binary name
BIN := mini-gcc

# Source files
SRCS := main.go lexer.go parser.go codegen.go codegen_binary.go instructions.go macho_writer.go

.PHONY: all clean test demo

# Build the compiler
all: $(BIN)

$(BIN): $(SRCS)
	$(GO) build -o $(BIN) $(SRCS)

# Clean up build artifacts and temporary test binaries
clean:
	rm -f $(BIN) *.o *.s a.out
	rm -f test_bin test_mixed test_paren test_var test_ptr test_array test_string test_char test_for test_global test_func test_lib_simple test_lib_strcpy test_lib_full test_simple_global test_control test_factorial test_prog

# Run a quick demo
demo: $(BIN)
	@echo "Creating test program..."
	@echo 'int printf(char *s, ...); int main() { printf("Hello from mini-gcc!\n"); return 42; }' > test_prog.c
	@echo "Compiling..."
	./$(BIN) -S test_prog.c
	cc -o test_prog test_prog.s
	@echo "Running..."
	./test_prog
	@echo "Exit code: $$?"
	@rm test_prog.c test_prog.s test_prog

# Run the test suite (placeholder)
test: $(BIN)
	@echo "Running basic tests..."