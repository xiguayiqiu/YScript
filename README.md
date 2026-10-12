<div align="center">
  <img src="icon.png" alt="YScript Logo" width="128" height="128">
</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-555555?style=flat-square&logo=linux&logoColor=white)](#快速开始)
[![CUDA](https://img.shields.io/badge/CUDA-optional-76B900?style=flat-square&logo=nvidia&logoColor=white)](#特性速览)
[![License](https://img.shields.io/badge/License-Apache--2.0-D22128?style=flat-square&logo=apache&logoColor=white)](LICENSE)

</div>

# YScript

<div align="center">

**面向网络安全工作的脚本语言**

用 Go 实现，单文件 `ysc` 即可运行，无需编译环境。
把「协议解析 / 抓包分析 / 密码学运算 / 漏洞验证」这类工作从零散的 Shell 命令，
变成可读、可复用、可版本管理的脚本。

### [在线学习文档（60 篇）](https://github.com/xiguayiqiu/YScript/wiki)

</div>

```yscript
$ cat hello.ys
package main

func main() {
    println("Hello, YScript!")
}

$ ysc hello.ys
Hello, YScript!
```

> 📦 **本仓库为发布仓库**：只包含可直接阅读、运行的**示例脚本与测试套件**。
> 解释器源码与完整中文手册在开发仓库中维护（见文末[仓库定位](#仓库定位)）。

---

## 什么是 YScript

YScript 是一门**为安全工作者设计**的动态类型脚本语言。它不是通用语言的替代品，
而是补上了安全工具链里长期缺失的一环：**用一门真正的编程语言，去表达安全工作中
那些「结构复杂、逻辑密集、步骤繁琐」的任务**。

### 为什么需要它

传统做法通常有三种，都不太舒服：

| 做法 | 问题 |
|------|------|
| 拼 Shell 命令 | 解析二进制、状态分支、错误处理都很别扭；`bash` 处理结构化数据能力弱 |
| 写 Python 脚本 | 环境依赖多（`scapy`/`cryptography` 等）、分发困难、启动开销大 |
| 写 C / Go 工具 | 一个几十行的协议解析工具要编译、要建工程，改一行就要重新构建 |

YScript 把这些痛点合并成一个答案：**语法接近 Python 的脚本体验，
能力接近 Go 的标准库**。

### 它擅长什么

```yscript
// 1. 协议解析 —— binary 命名空间 100 个成员，大小端/流抽象一应俱全
let head = binary.BigEndian().PutUint16(80)     // b"\x00\x50"
let port = binary.Uint16(head)                  // 80

// 2. 网络通信 —— socket 统一 TCP/UDP/TLS，一个对象三种协议
let s = socket.Socket("tcp")
s.connect("192.168.1.1", 22)
s.set_timeout(3000)
let banner = s.recv_line()                      // SSH banner
println(binary.UTF8(banner))

// 3. 密码学 —— 45 个 crypto 函数，含 WPA2 全套
let pmk = crypto.WPA2_PMK(passphrase, ssid)
crypto.WPA2_MIC_Verify(pmk, eapol, mic)

// 4. 抓包取证 —— raw + pcap，直接读原始帧
let h = raw.PcapOpen("eth0", "tcp port 80")
let pkt = raw.PcapNext(h)                      // 实时抓包（libpcap）
let rec = raw.PcapReadFile("cap.pcap")         // 或纯 Go 解析文件

// 5. Shell 与系统 —— 反引号即命令，进程/路径/定时器一应俱全
let out = `nmap -sV -p 22,80,443 {target}`
os.exec("aircrack-ng", ["-b", "wpa.cap"])
```

### 写起来是什么感觉

```yscript
// 异常：一行兜底，不用写 try/catch 样板代码
let n = parse_int(user_input) catch 0

// 带类型异常分流：连接失败和超时分开处理
catch {
    socket.Socket("tcp").connect(host, port)
} match {
    ConnectError(e) -> log.error("连接失败: " + e)
    TimeoutError(e) -> log.error("超时: " + e)
    _               -> log.error("其它错误")
} ensure {
    cleanup()          // 无论成败都执行
}

// 继承多态：把「基类 - 子类 - 子类」写成清晰的层次
struct Animal {
    name: string
}
func this.speak() -> string { return "..." }

struct Dog extends Animal {
    breed: string
}
func this.speak() -> string { return "汪汪" }
func this.describe() -> string { return super.describe() + "，是只" + this.breed }
```

### 安全领域的实际用例

仓库中的示例都是真实场景，可直接阅读：

| 场景 | 文件 |
|------|------|
| **WPA2 握手包破解**（PMK/PTK 派生 + MIC 校验，支持 GPU 加速） | [`test/wifi_crack.ys`](test/wifi_crack.ys) |
| WPA2 破解流程（英文版 / 结果校验） | [`test/wifi_verify.ys`](test/wifi_verify.ys) |
| WPA2 字典破解性能测试 | [`test/wifi_speed.ys`](test/wifi_speed.ys) |
| **服务端口扫描与指纹识别** | [`service_scanner.ys`](service_scanner.ys) |
| 网络协议与 socket 实战 | [`test/net_ext.ys`](test/net_ext.ys) |
| 二进制格式解析与修补 | [`test/binary_lib.ys`](test/binary_lib.ys) |

### 安全边界：沙箱与许可

安全工具需要执行危险操作，YScript 把权限控制做成**语言层面的显式声明**：

```yscript
#!permit file_read:/etc/*, exec:nmap, network
```

脚本声明所需权限，未授权的操作会被拦截；也可通过命令行策略覆盖。
详见《39 预处理器指令》。

---

## 特性速览

| 能力 | 说明 |
|------|------|
| **密码学** | `crypto` 45 个函数：SHA/SHA3/SHAKE、AES-GCM、ChaCha20、HMAC、PBKDF2、ECDSA、Ed25519、**WPA2 全套** |
| **协议分析** | `binary` 100 个函数：大小端编解码、字节容器、流抽象、**等长安全修补** |
| **网络** | `socket` 统一 TCP/UDP/TLS 对象；`net` 底层连接；`raw` 原始帧 + **pcap 抓包**；`ssl` 证书与 TLS |
| **系统与 Shell** | 反引号直接执行命令；`os` 进程管理、`sys` 系统信息、`path` 路径处理 |
| **并发** | `warp` 轻量协程 + `sync` 全套同步原语（Mutex/RWMutex/Chan/TLS/WaitGroup/Once/WorkerPool） |
| **异常** | `expr catch h` 一行兜底、`catch {} match { T(e) -> }` 按类型分流、`ensure` 保证清理 |
| **面向对象** | `struct` 方法、接口多态、继承多态（`extends` + `super`）、泛型函数/结构体、类型推导与约束 |
| **GPU 加速** | `cuda` 命名空间支持密码学运算批处理，无 GPU 时自动降级 CPU |
| **FFI** | `ffi` 真实 ABI 调用动态库（0-8 参数，int/string/void 返回） |
| **工程** | 预处理器、REPL 会话、`.ybc` 字节码落盘、中英双语 i18n |

---

## 快速开始

本仓库不含可执行文件，需先安装 `ysc`（YScript 解释器）：

```bash
ysc hello.ys             # 执行脚本
ysc                       # 无参数 = 进入 REPL 会话
ysc -e 'println(1+1)'     # 直接执行一段代码
ysc -c hello.ys           # 仅做语法检查，不执行
ysc --emit a.ybc x.ys     # 导出字节码
ysc pkg -target linux/amd64 hello.ys # 打包为独立 Linux 可执行程序
ysc test ./test/...       # 运行 *_test.ys 测试
ysc fmt ./src/...         # 递归格式化 YScript 源文件
ysc fmt -check .          # 检查当前目录格式
cat app.ys | ysc fmt -t    # 从 stdin 读入并向 stdout 输出，不修改文件
ysc mod init myapp        # 初始化项目清单
ysc mod get https://github.com/org/lib.git v1.0.0  # 版本参数可省略，默认锁定当前 HEAD
ysc mod sync              # 按 import 同步、校验及清理依赖
ysc doc                  # 查看当前目录包文档
ysc doc ./pkg Serve      # 查看本地包的符号文档
ysc doc Serve            # 搜索当前项目和全局库中的符号
ysc doc thirdparty/pkg   # 查看全局库仓库中的第三方包
ysc doc net              # 查看 ysc 内置标准库命名空间及成员
ysc doc net.LookupHost   # 查看内置标准库成员
ysc -h                    # 查看全部选项
```

### 打包为独立程序

`.ybc` 是可由 `ysc` 直接执行的编译后字节码文件：文件采用带魔数、格式版本、解压长度与校验和的二进制容器，并压缩保存编译 IR。`ysc --emit` 与 `ysc pkg --ybc` 都输出此格式；`ysc program.ybc` 可直接加载运行。当前运行时也能读取旧版 JSON `.ybc`，但新编译一律输出二进制格式。`ysc pkg` 将入口脚本编译为优化后的 `.ybc`，嵌入 Go 运行时并调用 Go 工具链生成目标平台的独立可执行文件。使用 `ysc pkg --ybc` 时只生成优化后的 `.ybc` 中间字节码，不会构建可执行文件，也不需要 Go 构建运行时。普通脚本只打包指定入口及其静态 import 闭包；传入含 `ysc.models` 的项目目录时，会扫描项目 `.ys` 文件并自动选择唯一声明 `main()` 的文件作为入口，不要求文件名为 `main.ys`。如果项目中有多个 `main()`，可显式传入入口文件。项目打包会递归同步校验项目及依赖库锁定的依赖，再将入口依赖的项目源码合并进字节码，因此这些静态依赖不需要运行时访问项目文件或依赖缓存。通过 `load.load` / `load.install` 等动态路径加载的脚本或模块不会自动嵌入，仍需按程序使用的路径部署。打包时按 import 生成标准库注册表和运行时桥接，仅链接引用的标准库模块；未使用的 eval/动态加载编译器、调试器及源码行映射也会移除。Go 运行时源码已随 `ysc` 内嵌，无需单独安装或定位 YScript Go 源码仓库；构建独立可执行文件时仍需 Go 工具链和 Go 模块依赖（可联网下载或使用本地缓存）。默认关闭 cgo 以便交叉编译；本机依赖已安装时可用 `-cgo` 启用，交叉 cgo 还需自行配置目标 C 编译器和目标依赖库。

```bash
ysc pkg hello.ys
ysc pkg --ybc hello.ys # 只生成优化后的 hello.ybc
ysc pkg -target linux/amd64 hello.ys
ysc pkg -target linux/amd64 -target windows/amd64 -o dist/hello hello.ys
ysc pkg ./my-project # 自动选择项目中声明 main() 的唯一 .ys 文件
ysc pkg -cgo hello.ys # 启用本机 cgo 能力
```

默认产物写入当前工作目录：单目标时使用脚本名，项目打包时使用 `ysc.models` 声明的名称；Windows 目标会添加 `.exe`。指定多个目标时会追加系统和架构后缀以避免覆盖。目标支持 Linux、Windows、macOS（`darwin`）、FreeBSD、OpenBSD、NetBSD、DragonFly、Solaris 和 illumos，以及 `386`、`amd64`、`arm`、`arm64`。详情见[独立程序打包指南](../doc/YScript.wiki/59_pkg打包.md)。

`--ybc` 默认写出 `<脚本名或项目名>.ybc` 到当前工作目录；可用 `-o` 指定路径（未写扩展名时自动补 `.ybc`）。该模式不能与 `-target` 或 `-cgo` 同时使用。

### 查询 YScript 文档

`ysc doc` 按 Go Doc 的包 / 符号查询方式读取 `.ys` 源码注释。无参数时查看当前目录；传入本地目录或源文件路径可查看本地项目，传入 `项目名/包路径` 可查看 `~/.ysc/pkg`（或 `YSC_PKG_DIR`）中的第三方库。符号查询可写成 `ysc doc <包> <符号>`，也可在当前目录使用 `ysc doc <符号>`。传入 ysc 内置标准库名（如 `net`、`json`）可查看命名空间成员；使用 `ysc doc net.LookupHost` 可查看单个成员。

```bash
ysc doc
ysc doc ./src/http
ysc doc mylib/http Client
ysc doc -all ./src/http      # 兼容选项；默认已列出全部声明
ysc doc -src ./src/http Get  # 显示该符号的源码
```

`#` 行注释或 `#* ... *#` 块注释紧邻包声明或符号声明时，会作为文档显示；函数注释也可放在文件任意位置，用 `@函数名 HEAD` 与 `@函数名 END` 标记，再由 `ysc doc` 按函数名关联。注解文档优先于紧邻注释，标记及区块内容不会影响脚本执行。YScript 不以首字母大小写区分导出符号，包概览默认展示所有函数（包括小写函数和 `main`）；`-all` 为兼容选项。`ysc doc -h` 查看选项用法。

```yscript
@banner_text HEAD
# 将服务 Banner 字节转换为可读文本。
@banner_text END

func banner_text(data) {
    return string(data)
}
```

### 测试 YScript 代码

测试文件使用 `*_test.ys` 命名，并在 `package main` 中声明 `Test` 开头、后接大写字母或数字的函数。测试函数可不带参数，也可接收测试上下文 `t`。测试命令的帮助、选项错误、测试发现/编译错误、测试上下文 API 错误和运行摘要会按当前 locale 显示中文或英文；测试名称、测试日志与断言消息保留脚本原文：

```yscript
func TestAddition(t) {
    if 1 + 1 != 2 {
        t.Errorf("1 + 1 = %d, want 2", 1 + 1)
    }
}
```

测试上下文提供 `Error` / `Errorf`、`Fatal` / `Fatalf`、`Fail` / `FailNow`、`Failed`、`Log` / `Logf`、`Skip` / `SkipNow`、`Skipped`、`Name`、`Cleanup`、`TempDir` 和 `Run` 子测试。运行器按包目录合并测试文件、编译并在同一 VM 中执行测试函数；测试文件不要声明 `main()`。

```bash
ysc test ./test/...                         # 递归发现并运行测试
ysc test -v ./test/generic                  # 显示名称、日志和通过结果
ysc test -run 'TestGenericCollections' ./test/generic
ysc test -count=3 ./test/generic            # 重复运行
ysc test -list 'TestGeneric' ./test/generic # 列出匹配测试
ysc test ./test/generic/generic_test.ys     # 指定测试文件
```

当传入目录时，只发现该目录直接包含的 `*_test.ys`；`...` 表示递归包目录。`-run` 和 `-list` 接受 Go 风格正则表达式。测试失败会返回非零退出码。

### 第一个程序：探测一个端口

```yscript
package main

func main() {
    let host = "192.168.1.1"
    let port = 22

    // 探测端口是否开放：一行即可，失败自动返回 false
    if socket.tcp_probe(host, port) {
        println("SSH 端口开放")

        // 连上去读 banner
        let s = socket.Socket("tcp")
        s.connect(host, port)
        s.set_timeout(3000)
        println(binary.UTF8(s.recv_line()))
        s.close()
    } else {
        println("端口关闭或被过滤")
    }

    // 解析失败不中断脚本：一行兜底
    let n = parse_int("abc") catch 0
    println("解析失败取默认值: " + string(n))
}
```

对照一下，同样的事如果用 Shell 写，需要 `nc`/`timeout`/`grep` 串联，
且无法优雅地区分「连接被拒」和「超时」——YScript 里这只是一个 `match`。

---

## 示例脚本

### [`service_scanner.ys`](service_scanner.ys) — 服务端口扫描器

一个可读的完整实战示例：`init_services()` 建立端口 → 服务名映射表，`main()` 负责交互和扫描，运行时结合 `socket` 与 Shell 命令完成端口探测。

```yscript
func init_services() {
    let services = {}
    services.set(21,  "FTP")
    services.set(22,  "SSH")
    services.set(80,  "HTTP")
    services.set(443, "HTTPS")
    # ... 共 100+ 常用服务
    return services
}

func main(args) {
    let services = init_services()
    # 交互及扫描逻辑
}
```

---

## 测试套件（含独立测试文件与 `.ys` 示例）

`test/` 下的脚本既是**回归测试**，也是**按特性组织的语法示例**，可直接阅读学习。

| 特性 | 文件 |
|------|------|
| 入口 / 总览 | [`main.ys`](test/main.ys) · [`comprehensive.ys`](test/comprehensive.ys) · [`features.ys`](test/features.ys) |
| 继承多态 | [`poly.ys`](test/poly.ys) |
| 异常捕获 | [`exception_catch.ys`](test/exception_catch.ys) · [`error_codes.ys`](test/error_codes.ys) · [`safe.ys`](test/safe.ys) |
| 切片 | [`slice.ys`](test/slice.ys) |
| 二进制 | [`binary_lib.ys`](test/binary_lib.ys) · [`memory.ys`](test/memory.ys) |
| 网络 | [`net_ext.ys`](test/net_ext.ys) |
| 并发 | [`sync.ys`](test/sync.ys) |
| 面向对象 | [`struct.ys`](test/struct.ys) · [`interface.ys`](test/interface.ys) · [`enum.ys`](test/enum.ys) · [`generics.ys`](test/generics.ys) |
| 集合 / 迭代 | [`rf2.ys`](test/rf2.ys) |
| 文件与系统 | [`sys.ys`](test/sys.ys) · [`shell_ext.ys`](test/shell_ext.ys) · [`io_ext.ys`](test/io_ext.ys) |
| 文本 / 时间 | [`time_lib.ys`](test/time_lib.ys) |
| 配置解析 | [`toml_ext.ys`](test/toml_ext.ys) · [`ini_ext.ys`](test/ini_ext.ys) · [`yaml_ext.ys`](test/yaml_ext.ys) |
| WiFi 实战 | [`wifi_crack.ys`](test/wifi_crack.ys) · [`wifi_verify.ys`](test/wifi_verify.ys) · [`wifi_speed.ys`](test/wifi_speed.ys) |

### 代码示例

**继承多态**（[`test/poly.ys`](test/poly.ys)）—— Java 风格的 `extends` / 方法重写 / `super` 复用：

```yscript
struct Animal {
    name: string
}
func this.speak() -> string { return "..." }
func this.describe() -> string { return "我是 " + this.name }

struct Dog extends Animal {
    breed: string
}
func this.speak() -> string { return "汪汪" }
func this.describe() -> string {
    return super.describe() + "，是只" + this.breed
}
```

**带类型异常分流**（[`test/exception_catch.ys`](test/exception_catch.ys)）—— `match` 按异常类型分支，`ensure` 保证清理：

```yscript
func scan_host(host: string) {
    catch {
        let sock = connect(host)
        println("  已连接 " + sock)
    } match {
        NetError(e)     -> println("  连接失败: " + e),
        TimeoutError(e) -> println("  超时: " + e),
    } ensure {
        println("  释放资源 " + host)
    }
}
```

> `ensure` 在未命中 `match`（异常继续透传）时**也会执行**，
> 因此异常无论被谁捕获，资源都一定被释放。

**一行兜底**：

```yscript
let n = parse_int("abc") catch 0          // 失败取默认值
let s = connect(host, port) catch return  // 失败即返回
```

---

## 标准库命名空间

40+ 命名空间，按用途分组：

| 分组 | 命名空间 |
|------|----------|
| 内建转换 | `hex` `alpha` `alnum` `ascii` `errors` |
| 文本 | `strings` `encoding` `json` `regex` `color` `array` |
| 字节流 | `binary` |
| 文件 / IO | `io` `path` `stdio` `os` `load` |
| 序列化 | `csv` `xml` `yaml` `toml` `ini` |
| 网络 | `socket` `net` `raw` `ssl` `http` `url` |
| 加密 | `crypto` `aes` `rsa` `hash` |
| 并发 | `thread` `sync` `event` |
| 系统 | `sys` `time` `rand` `log` |
| 其它 | `ffi` `cuda` `reflect` `iter` `from` `compress` |

`load` 支持运行时动态管理 YScript 代码：`load.load` / `load.reload` 处理普通函数插件；`load.install`、`load.start`、`load.stop`、`load.uninstall` 管理可包含全局变量、类型和模块入口的完整脚本模块。完整说明见 `doc/YScript.wiki/54_热加载与热插拔.md`。

`event.wait()` 可在 VM 主执行线程等待宿主投递的回调；收到通知后，VM 会在安全点执行回调。`event.wait(timeout_ms)` 可设置毫秒超时，并在超时返回 `false`、收到通知时返回 `true`。该接口用于嵌入式宿主事件循环；普通 CLI 不会自行产生 GUI 事件，且 queued callback 内不能再次调用 `event.wait()`。

Go 宿主可导入 `yscript/host` 编译并运行脚本。`Bind` 将宿主后端操作注册为脚本可调用函数，`Post` 则按顶层函数名安全投递事件回调；回调参数在入队前转换并复制：

```go
package main

import (
    "context"
    "yscript/host"
)

func run(source string) error {
    script, err := host.New(source, host.Options{Filename: "ui.ys"})
    if err != nil { return err }
    defer script.Close()
    ready := make(chan struct{})
    if err := script.Bind("host_ready", func([]any) (any, error) {
        close(ready)
        return nil, nil
    }); err != nil { return err }

    runDone := make(chan error, 1)
    go func() { runDone <- script.Run() }()
    <-ready
    task, err := script.Post("on_click", 12, 30)
    if err != nil { return err }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    if _, err := task.Wait(ctx); err != nil { return err }
    return <-runDone
}
```

每个进程当前只能有一个活动 `Host`（VM 的标准库桥接回调仍是进程级状态）；关闭它后才能创建另一个。`Bind` 必须在 `Run` 前完成；`Run` 和 `RunEventLoop` 必须串行调用，`Post` 可从宿主的其他 goroutine 调用。需要生命周期取消时使用 `RunContext(ctx)`；它会在字节码指令边界响应取消，也会解除阻塞的 `event.wait()`。已经进入的 Go 绑定函数需由该函数自身配合取消；`Close` 会取消脚本执行、阻止新任务、失败尚在队列中的任务，并唤醒等待。

`host.Options` 还可注入 `Stdin`、`Stdout`、`Stderr`，分别供 YScript 的 `io.Stdin()`、`io.Stdout()`、`io.Stderr()` 使用；未提供时使用进程标准流。绑定的 Go 函数在 VM 执行线程同步调用；若 GUI 工具包要求固定 UI 线程，宿主函数应将操作转发到该工具包的线程，并通过 `Post` 把完成事件送回 VM。宿主仍需提供真正的窗口、绘制和输入后端。

### `binary` — 字节流处理（速览）

```yscript
// 大小端编解码
binary.BigEndian()
let head = binary.PutUint16(80)         // b"\x00\x50"
let port = binary.Uint16(b"\x00\x50")    // 80

// 流抽象
let w = binary.NewWriter()
binary.WriteString(w, "hello")
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

📖 **在线阅读（推荐）**：**[YScript Wiki →](https://github.com/xiguayiqiu/YScript/wiki)** —— 共 **60 篇**中文手册，可直接浏览与搜索；测试用例编写和 `ysc test` 命令详见[测试框架指南](https://github.com/xiguayiqiu/YScript/wiki/56_YScript测试框架)，`ysc fmt` 用法详见[代码格式化指南](https://github.com/xiguayiqiu/YScript/wiki/57_YScript代码格式化)，`ysc doc` 用法详见[文档查询指南](https://github.com/xiguayiqiu/YScript/wiki/58_YScript文档查询)，白名单 native 调用详见[lstd 指南](https://github.com/xiguayiqiu/YScript/wiki/60_lstd安全动态库调用)。

完整手册同时以 Markdown 形式存放于开发仓库的 `doc/` 目录（含按命名空间分类的**函数速查表**：45 命名空间 + 6 类型 / 976 个命名空间成员 + 113 个类型方法）：

| 主题 | 章节 |
|------|------|
| 语言基础 | 01 类型系统 · 02 变量与常量 · 03 流程控制 · 04 函数 · 04 运算符 |
| 数据结构 | 06 字符串操作 · 07 字节序列 · 08 集合类型 · 09 数组 · 44 切片 |
| 面向对象 | 10 结构体与方法 · 11 接口 · 46 继承与多态 |
| 异常 | 18 错误处理 · 48 异常捕获 |
| 并发 | 27 并发 · 28 warp 协程 · 29 同步原语 · 47 协程与互斥锁 |
| 字节流 | 20 字节流处理 |
| 网络 | 30 网络通信 · 45 TCP 与 UDP · 31 raw 原始帧 · 32 SSL |
| 系统 | 16 文件与路径 · 25 系统与进程 · 37 日志库 |
| Native 互操作 | 35 FFI 动态库 · 60 lstd 白名单动态库调用 |
| 速查 | 00 实现状态 · 05 关键字速查 · 函数速查表 |

编辑器支持：**vim 插件** · **VSCode 插件**（语法高亮 / LSP 智能补全 / 悬停文档 / 跳转定义）。

---

## 仓库定位

| 仓库 | 用途 | 内容 |
|------|------|------|
| **YScript**（本仓库） | **发布仓库** | 示例脚本、测试套件、许可证 |
| 开发仓库 | 源码维护 | Go 解释器源码、`doc/` 完整手册、`vim/` `vscode/` 插件、`Makefile` 构建 |

本仓库通过 `.gitignore` 采用「默认忽略 + 白名单」策略（`/*` 之后仅放行
`service_scanner.ys`、`test/**`、`README*`），确保发布内容干净可控。

---

## 版本

当前版本 **v0.2.2**。完整变更记录见开发仓库的 `doc/Log.md`。

## 许可

Apache-2.0，见 [LICENSE](LICENSE)。
