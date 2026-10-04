# YScript

一套用 Go 实现的脚本语言：动态类型、`catch/match/ensure` 异常、`warp` 协程、
40+ 标准库命名空间，面向网络协议解析、二进制处理与系统自动化场景。

```
$ cat hello.ys
package main

func main() {
    println("Hello, YScript!")
}

$ ysc hello.ys
Hello, YScript!
```

---

## 特性速览

| 能力 | 说明 |
|------|------|
| **异常** | `expr catch h`、`catch { } match { T(e) -> }`、`ensure` 清理块、`raise T("msg")` |
| **并发** | `warp` 轻量协程 + `sync` 全套同步原语（Mutex/RWMutex/Chan/TLS/WaitGroup/Once/WorkerPool） |
| **面向对象** | `struct` 方法、接口多态、**继承多态**（`extends` + `super`）、泛型注解 |
| **二进制** | `binary` 命名空间 99 个函数：字节容器、大小端编解码、流抽象、**等长安全修补** |
| **网络** | `socket` 统一 TCP/UDP/TLS 对象；`net` / `raw` / `ssl` 底层能力；pcap 抓包 |
| **系统** | 进程、Shell、文件路径、定时器、日志、CUDA 加速、FFI 动态库调用 |
| **工程** | 预处理器、REPL 会话、`.ybc` 字节码落盘（`--emit`）、交叉编译、中英双语 i18n |

---

## 快速开始

### 构建

```bash
cd yscript
make            # 编译本机 ysc（自动检测 CUDA）
make all        # 交叉编译全部平台并打包为 tar.xz（产物在 dist/）
make all -j8    # 并行编译，更快
make static     # 纯静态、无 cgo（可移植）
make linux      # 交叉编译 Linux amd64
make windows    # 交叉编译 Windows amd64
make darwin     # 交叉编译 macOS arm64
make help       # 查看全部目标
```

`make all` 默认为下列 6 个平台各产出一个 `tar.xz`（并生成 `SHA256SUMS`）：

| 平台 | 产物 |
|---|---|
| `linux/amd64`、`linux/arm64` | `dist/ysc-<版本>-linux-<arch>.tar.xz` |
| `darwin/amd64`、`darwin/arm64` | `dist/ysc-<版本>-darwin-<arch>.tar.xz` |
| `windows/amd64`、`windows/arm64` | `dist/ysc-<版本>-windows-<arch>.tar.xz` |

```bash
make all PLATFORMS="linux/amd64 windows/amd64"   # 只打指定平台
make dist-linux-amd64                            # 只做单个平台
make dist                                        # 额外保留解包目录
make dist-clean                                  # 清理 dist/
```

> **发布版产物**：`-s -w` 去掉符号表与 DWARF 调试信息，`-trimpath` 去掉
> 构建机绝对路径（源码目录 / GOPATH），既防信息泄露又保证可复现构建。
> 二者由 `RELEASE_LDFLAGS` / `TRIMPATH` 强制指定，覆盖 `LDFLAGS` 也不会失效
> （`LDFLAGS` 仅作为附加标志追加，例如 `-X main.version=...`）。
>
> 交叉编译一律关闭 CGO（无法跨平台链接），故 CUDA 自动降级为 CPU stub；
> 需要 CUDA 请在本机执行 `make`（仅 linux/amd64 生效）。

产物为 `ysc`。运行：

```bash
ysc script.ys             # 执行脚本
ysc                       # 无参数 = 进入 REPL 会话
ysc -e 'println(1+1)'     # 直接执行一段代码
ysc -c script.ys          # 仅做语法检查，不执行
ysc --emit a.ybc x.ys     # 导出字节码
ysc --sandbox policy.json # 按沙箱策略限制权限
ysc -h                    # 查看全部选项
```

### 第一个程序

```yscript
package main

func main() {
    // 集合与迭代
    let xs = [1, 2, 3, 4]
    let evens = []
    for x in xs {
        if x % 2 == 0 { evens.append(x) }
    }
    println("偶数: " + string(evens))

    // 异常处理：一行兜底
    let n = parse_int("abc") catch 0
    println("解析失败取默认值: " + string(n))
}
```

---

## 标准库命名空间

| 分组 | 命名空间 |
|------|----------|
| 内建转换 | `hex` `alpha` `alnum` `ascii` `errors` |
| 文本 | `strings` `encoding` `json` `regex` `color` `array` |
| 字节流 | `binary` |
| 文件 / IO | `io` `path` `stdio` `os` |
| 序列化 | `csv` `xml` `yaml` `toml` `ini` |
| 网络 | `socket` `net` `raw` `ssl` `http` `url` |
| 加密 | `crypto` `aes` `rsa` `hash` |
| 并发 | `thread` `sync` |
| 系统 | `sys` `time` `rand` `log` |
| 其它 | `ffi` `cuda` `reflect` `iter` `from` `compress` |

### `binary` — 字节流处理（速览）

```yscript
// 大小端编解码
binary.BigEndian()
let head = binary.PutUint16(80)         // b"\x00\x50"
let port = binary.Uint16(b"\x00\x50")    // 80

// 流抽象
let w = binary.NewWriter()
binary.WriteString(w, "hello ")
binary.WriteUint32(w, 0x01020304)
binary.WriterBytes(w)
binary.CloseWriter(w)

// 二进制修补（等长覆盖，防损坏）
let off  = binary.Index(data, binary.FromUTF8("Hello Word!"))
let out  = binary.PatchEqual(data, off, binary.FromUTF8("Hello Word!"),
                            binary.FromUTF8("Hello PATCH"))
```

> ⚠️ Go 把字符串常量**无分隔符紧密排列**。用 `Replace` 做变长替换会移动其后
> 所有数据 → 运行期 **SIGSEGV**，且 `readelf` 检查不出来。`PatchEqual` 在写入前
> 强制等长 + 偏移合法 + 内容匹配三重校验。

### `socket` — TCP / UDP / TLS（速览）

```yscript
// TCP 服务端
let srv = socket.Socket("tcp")
let port = srv.listen(8080)             # 0 = 由内核分配
let cli = srv.accept()                  # 返回新的 Socket
cli.send("hello\n")
println(binary.UTF8(cli.recv_line()))
cli.close()
srv.close()

// UDP 请求/响应
let us = socket.Socket("udp")
let up = us.bind("0.0.0.0", 9000)
let uc = socket.Socket("udp")
uc.connect_udp("127.0.0.1", up)          # 先固定源端口
uc.send("ping")
let r = uc.recvfrom(2048)               # [数据, 来源地址]

// 一行便捷函数
let resp = binary.UTF8(socket.tcp_request("127.0.0.1", 80, "GET / HTTP/1.1\r\n\r\n"))
if socket.tcp_probe("192.168.1.1", 22) { println("SSH 开放") }
```

失败抛出**带类型异常**，可按类型分流：

```yscript
catch {
    let s = socket.Socket("tcp")
    s.connect(host, port)
} match {
    ConnectError(e) -> println("连接失败: " + e)
    TimeoutError(e) -> println("超时: " + e)
    TLSError(e)     -> println("TLS 失败: " + e)
    SocketError(e)  -> println("其它: " + e)
}
```

---

## 文档

完整手册见 [`doc/`](doc/)，共 48 章 + 更新日志：

| 主题 | 章节 |
|------|------|
| 语言基础 | [01 类型系统](doc/01_类型系统.md) [02 变量与常量](doc/02_变量与常量.md) [03 流程控制](doc/03_流程控制.md) [04 函数](doc/04_函数.md) |
| 数据结构 | [09 数组](doc/09_数组.md) [08 集合类型](doc/08_集合类型.md) [07 字节序列](doc/07_字节序列.md) [44 切片](doc/44_切片.md) |
| 面向对象 | [10 结构体与方法](doc/10_结构体与方法.md) [11 接口](doc/11_接口.md) [46 继承与多态](doc/46_继承与多态.md) |
| 异常 | [18 错误处理](doc/18_错误处理.md) [48 异常捕获](doc/48_异常捕获.md) |
| 并发 | [27 并发](doc/27_并发.md) [28 warp协程](doc/28_warp并发线程.md) [29 warp同步原语](doc/29_warp同步原语.md) [47 协程与互斥锁](doc/47_协程与互斥锁.md) |
| 字节流 | [20 字节流处理](doc/20_字节流处理.md) |
| 网络 | [30 网络通信](doc/30_网络通信.md) [45 TCP与UDP](doc/45_TCP与UDP网络编程.md) [31 raw原始帧](doc/31_raw网络原始帧操作.md) [32 SSL](doc/32_ssl安全套接层.md) |
| 系统 | [16 文件与路径](doc/16_文件与路径.md) [25 系统与进程](doc/25_系统与进程库.md) [37 日志库](doc/37_日志库.md) |
| 速查 | [**49_函数速查表.md**](doc/函数速查表.md)（37 命名空间 + 6 类型 / 729 个函数）· [05 关键字速查](doc/05_关键字速查.md) |
| 其它 | [00 实现状态](doc/00_实现状态.md) [**Log.md 更新日志**](doc/Log.md) |

编辑器支持：[vim 插件](vim/) · [VSCode 插件](vscode/)

---

## 项目结构

```
y_script/
├── yscript/              解释器主仓库（Go）
│   ├── cmd/yscript/      命令行入口
│   ├── internal/
│   │   ├── lexer parser checker compiler   前端
│   │   ├── bytecode vm                      运行时
│   │   ├── value                            值表示与类型系统
│   │   ├── preproc i18n                     预处理器 / 中英双语
│   │   └── std/                             40+ 标准库命名空间
│   ├── test/              YScript 测试套件
│   └── Makefile
├── doc/                  48 章中文手册 + Log.md
├── vim/  vscode/          编辑器插件
├── testexe/              二进制修补实验目标
└── wifi/                 实战项目（WiFi 相关）
```

---

## 开发

```bash
make test    # go test ./cmd/... ./internal/...
make vet     # go vet
make fmt     # gofmt
make clean   # 清理产物
```

---

## 版本

当前版本 **v0.1.4**（见 `yscript/cmd/yscript/main.go` → `const version`）。
完整变更记录见 [doc/Log.md](doc/Log.md)。

## 许可

见 [LICENSE](yscript/LICENSE)。