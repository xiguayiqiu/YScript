# lstd native integration tests

This fixture exercises `lstd` against a small C++ shared library, rather than
only against platform system libraries. The YScript integration test covers
symbol lookup and binding, integer and floating-point calls, strings, structs
passed and returned by value, callbacks, native memory, error conversion, and
allowlist/closed-library failures.

Run the native interpreter test from the repository root:

```sh
./test/lstd/test.sh
```

The runner also compiles the same YScript file with `ysc pkg -cgo` and executes
the packaged binary against the fixture, then verifies default-deny, symbol
allowlist, and closed-library failures.

The test requires cgo, a C++ compiler, `pkg-config`, and the libffi development
package. Set `YSCRIPT_BIN` to use an existing interpreter binary; otherwise the
script runs the current source with `go run`.

Build shared libraries for individual targets with `build.sh`:

```sh
./test/lstd/build.sh linux/amd64
./test/lstd/build.sh linux/arm64
./test/lstd/build.sh windows/amd64 windows/386
./test/lstd/build.sh darwin/arm64 darwin/amd64
```

The compiler defaults cover common GNU/MinGW toolchain names; the macOS
cross-link requires Clang and `ld64.lld`. Override a
compiler with `CXX_LINUX_ARM64`, `CXX_WINDOWS_AMD64`, `CXX_WINDOWS_386`,
`CXX_DARWIN_ARM64`, or `CXX_DARWIN_AMD64`. Cross-built libraries are compile
checks only and must be run on their target OS/architecture. To build and run
for a non-host location, set `OUT_DIR` and `YSCRIPT_BIN` as appropriate.
