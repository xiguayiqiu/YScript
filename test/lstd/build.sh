#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
out_root=${OUT_DIR:-"$script_dir/build"}

host_target=$(go env GOOS)/$(go env GOARCH)

compiler_for() {
    target_os=$1
    target_arch=$2
    case "$target_os/$target_arch" in
        linux/amd64)
            if [ -n "${CXX_LINUX_AMD64:-}" ]; then
                printf '%s\n' "$CXX_LINUX_AMD64"
                return
            fi
            if [ "$host_target" = linux/amd64 ]; then
                printf '%s\n' "${CXX:-c++}"
            else
                printf '%s\n' "${CXX:-x86_64-linux-gnu-g++}"
            fi
            ;;
        linux/arm64)
            printf '%s\n' "${CXX_LINUX_ARM64:-aarch64-linux-gnu-g++}"
            ;;
        linux/386)
            printf '%s\n' "${CXX_LINUX_386:-i686-linux-gnu-g++}"
            ;;
        windows/amd64)
            printf '%s\n' "${CXX_WINDOWS_AMD64:-x86_64-w64-mingw32-g++}"
            ;;
        windows/386)
            printf '%s\n' "${CXX_WINDOWS_386:-i686-w64-mingw32-g++}"
            ;;
        darwin/amd64)
            printf '%s\n' "${CXX_DARWIN_AMD64:-clang++}"
            ;;
        darwin/arm64)
            printf '%s\n' "${CXX_DARWIN_ARM64:-clang++}"
            ;;
        *)
            echo "unsupported target: $target_os/$target_arch" >&2
            return 1
            ;;
    esac
}

build_target() {
    target=$1
    case "$target" in
        */*) ;;
        *)
            echo "target must be os/arch, got: $target" >&2
            return 1
            ;;
    esac
    target_os=${target%/*}
    target_arch=${target#*/}
    compiler=$(compiler_for "$target_os" "$target_arch") || return
    if ! command -v "$compiler" >/dev/null 2>&1; then
        echo "missing C++ compiler for $target: $compiler" >&2
        return 1
    fi

    destination="$out_root/$target"
    mkdir -p "$destination"
    case "$target_os" in
        linux)
            "$compiler" -std=c++17 -O2 -fPIC -shared \
                "$script_dir/native/lstd_fixture.cpp" \
                -o "$destination/liblstd_fixture.so"
            ;;
        darwin)
            darwin_arch=$target_arch
            if [ "$target_arch" = amd64 ]; then
                darwin_arch=x86_64
            fi
            "$compiler" -target "$darwin_arch-apple-macosx11.0" -std=c++17 -O2 \
                -fPIC -dynamiclib -nostdlib -fuse-ld=lld \
                -Wl,-arch,"$darwin_arch" -Wl,-platform_version,macos,11.0,11.0 \
                "$script_dir/native/lstd_fixture.cpp" \
                -o "$destination/liblstd_fixture.dylib"
            ;;
        windows)
            "$compiler" -std=c++17 -O2 -shared \
                "$script_dir/native/lstd_fixture.cpp" \
                "$script_dir/native/lstd_fixture.def" \
                -o "$destination/lstd_fixture.dll"
            ;;
        *)
            echo "unsupported operating system: $target_os" >&2
            return 1
            ;;
    esac
    printf 'built %s for %s\n' "$destination" "$target"
}

if [ "$#" -eq 0 ]; then
    set -- "$host_target"
fi

for target do
    build_target "$target"
done
