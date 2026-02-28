# Mini-GCC

A functional C compiler written in Go for macOS (ARM64/Apple Silicon). It compiles a subset of C into ARM64 assembly, using the system toolchain (`cc`) for final assembly and linking.

## Features

*   **Types**: `int`, `char`, `void`, pointers (`int*`), arrays (`int a[10]`).
*   **Control Flow**: `if`, `else`, `while`, `for`.
*   **Functions**: Recursive functions, parameters, external calls (`printf`).
*   **Memory**: Global variables, stack variables, string literals.
*   **Intrinsics**: `__ldrb` (load byte) and `__strb` (store byte) for implementing standard library functions.

## Usage

### 1. Build using Make (Recommended)
```bash
make all    # Build the compiler 'mini-gcc'
make clean  # Clean up binaries and temp files
make demo   # Run a quick "Hello World" demo
```

### 2. Manual Build
```bash
go build -o mini-gcc main.go lexer.go parser.go codegen.go codegen_binary.go instructions.go macho_writer.go
```

### 3. Compile C Code
```bash
# Compile to executable (a.out)
./mini-gcc test.c

# Compile to specific output
./mini-gcc -o myprog test.c

# Compile to Assembly only
./mini-gcc -S test.c
```

### Example

**input.c**:
```c
int printf(char *fmt, ...);

int factorial(int n) {
    if (n == 0) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main() {
    int res = factorial(5);
    printf("Factorial of 5 is %d\n", res);
    return 0;
}
```

**Run**:
```bash
./mini-gcc input.c
./a.out
# Output: Factorial of 5 is 120
```

## Architecture
`mini-gcc` uses a Pratt Parser to generate an AST, which is then traversed to emit ARM64 assembly instructions. It manages stack frames, variable offsets, and calling conventions compatible with the macOS ABI.

## Limitations
*   No support for `struct`, `union`, `enum`, `float`.
*   No preprocessor (`#include`, `#define`).
*   No error recovery (stops on first error).