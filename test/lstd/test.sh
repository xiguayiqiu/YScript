#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
target=$(cd "$root" && go env GOOS)/$(cd "$root" && go env GOARCH)

if [ "$(cd "$root" && go env CGO_ENABLED)" != 1 ]; then
    echo "lstd integration tests require CGO_ENABLED=1" >&2
    exit 1
fi
if ! command -v pkg-config >/dev/null 2>&1 || ! pkg-config --exists libffi; then
    echo "lstd integration tests require the libffi development package" >&2
    exit 1
fi

"$script_dir/build.sh" "$target"
cd "$root"

if [ -n "${YSCRIPT_BIN:-}" ]; then
    ysc=$YSCRIPT_BIN
else
    ysc=
fi
run_ys() {
    if [ -n "$ysc" ]; then
        "$ysc" "$@"
    else
        go run ./cmd/yscript "$@"
    fi
}

library_name=liblstd_fixture.so
if [ "${target%/*}" = windows ]; then
    library_name=lstd_fixture.dll
elif [ "${target%/*}" = darwin ]; then
    library_name=liblstd_fixture.dylib
fi
library_path="test/lstd/build/$target/$library_name"
symbols='["ys_lstd_add","ys_lstd_weighted","ys_lstd_strlen","ys_lstd_greeting","ys_lstd_pair_transform","ys_lstd_apply","ys_lstd_upper_first","ys_lstd_ordinal_value","#17"]'

YSCRIPT_LSTD_ALLOW="{\"$library_path\":$symbols}" run_ys test/lstd/lstd.ys

package_dir=$(mktemp -d)
package_binary="$package_dir/lstd-fixture"
if [ "${target%/*}" = windows ]; then
    package_binary="$package_binary.exe"
fi
cleanup() {
    rm -f "$package_binary" "$package_dir/failure.log"
    rmdir "$package_dir"
}
trap cleanup EXIT HUP INT TERM
run_ys pkg -cgo -o "$package_binary" test/lstd/lstd.ys
YSCRIPT_LSTD_ALLOW="{\"$library_path\":$symbols}" "$package_binary"

expect_failure() {
    script=$1
    expected=$2
    output="$package_dir/failure.log"
    : >"$output"
    if YSCRIPT_LSTD_ALLOW="$3" run_ys "$script" >"$output" 2>&1; then
        cat "$output" >&2
        echo "expected $script to fail" >&2
        exit 1
    fi
    if ! grep -E "$expected" "$output" >/dev/null; then
        cat "$output" >&2
        echo "$script failed without expected diagnostic: $expected" >&2
        exit 1
    fi
    printf 'expected failure passed: %s\n' "$script"
}

expect_failure test/lstd/deny_load.ys 'denied by default|默认拒绝' ""
expect_failure test/lstd/deny_symbol.ys 'not in.*allowlist|符号不在白名单' \
    "{\"$library_path\":[\"ys_lstd_add\"]}"
expect_failure test/lstd/closed_library.ys 'closed|关闭' \
    "{\"$library_path\":[\"ys_lstd_add\"]}"
