# Mini-GCC Project Progress

## Overview
`mini-gcc` is a functional C compiler written in Go, targeting macOS (ARM64/Apple Silicon). It compiles a significant subset of C into native assembly, capable of producing executable binaries by linking with the system toolchain.

## Architecture
The compiler follows a standard pipeline:
1.  **Lexer (`lexer.go`)**: Tokenizes source code (e.g., `INT`, `IDENT`, `FOR`, `STRING`).
2.  **Parser (`parser.go`)**: Parses tokens into an Abstract Syntax Tree (AST) using a Pratt Parser. Handles operator precedence, complex statements, and types.
3.  **Code Generator (Text) (`codegen.go`)**: Traverses the AST and emits ARM64 assembly instructions (`.s`).
4.  **Driver (`main.go`)**: Orchestrates the compilation process.

## Implemented Features

### 1. Core Language
-   **Functions**: Definitions, calls, parameters, recursion.
-   **Variables**:
    -   **Local**: Stack-allocated.
    -   **Global**: Data-segment allocated.
    -   **Arrays**: Stack/Global allocation, index access (`a[i]`).
    -   **Pointers**: Address-of (`&`), Dereference (`*`), Assignment (`*p = val`).
-   **Control Flow**: `if`, `else`, `while`, `for`.
-   **Types**: `int`, `int*` (and pointers to pointers), `char` (basic support via intrinsics), `void`.

### 2. Data & Memory
-   **Strings**: String literals (`"hello"`) stored in `.cstring`.
-   **Memory Access**: Direct load/store via pointers and arrays.
-   **Intrinsics**: `__ldrb` (load byte), `__strb` (store byte) for low-level memory manipulation.

### 3. Integration
-   **External Calls**: Can call system functions like `printf` (linking with `libc`).
-   **Stdlib Capability**: Sufficiently powerful to implement `strlen`, `strcpy`, `strcmp` in C.

## Verification
We have verified the compiler with comprehensive test cases:
1.  **Basic Math**: `test.c` (returns `25`).
2.  **Variables**: `test_var.c` (returns `30`).
3.  **Control Flow**: `test_control.c` (Sum 0..9 = `45`, returns `1`).
4.  **Functions**: `test_func.c` (Call `add(10, 20)` returns `30`).
5.  **Recursion**: `test_factorial.c` (Factorial(5) returns `120`).
6.  **Global Variables**: `test_global.c` (State persistence across calls).
7.  **For Loops**: `test_for.c` (Sum loop returns `45`).
8.  **Pointers**: `test_ptr.c` (Dereference assignment returns `20`).
9.  **Arrays**: `test_array.c` (Index access returns `15`).
10. **Strings**: `test_string.c` (Prints "Hello World").

## Known Limitations
-   **OS Security**: Native binary generation (without `cc`) is blocked by macOS kernel security. We rely on `cc` as a backend assembler/linker.
-   **Variadic Functions**: Complex calls to variadic functions (like `printf` with mixed types inside loops) may have ABI alignment issues on Apple Silicon.
-   **Missing C Features**: Structs, Enums, Floating Point, Preprocessor macros (`#define`).

## Roadmap
1.  **Structs**: Implement memory layout for `struct`.
2.  **Preprocessor**: Add `#include` support.
3.  **Optimization**: Basic constant folding or register allocation.
