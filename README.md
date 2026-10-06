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

### [在线学习文档（51 篇）](https://github.com/xiguayiqiu/YScript/wiki)

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
// 1. 协议解析 —— binary 命名空间 99 个函数，大小端/流抽象一应俱全
let head = binary.BigEndian().PutUint16(80)     // b"\x00\x50"
let port = binary.Uint16(head)                  // 80

// 2. 网络通信 —— socket 统一 TCP/UDP/TLS，一个对象三种协议
let s = socket.Socket("tcp")
s.connect("192.168.1.1", 22)
s.set_timeout(3000)
let banner = s.recv_line()                      // SSH banner
println(binary.UTF8(banner))

// 3. 密码学 —— SHA-2 / SHA-3 / BLAKE2 / HMAC、AEAD、签名、KDF 与 WPA2
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
| **WPA2 握手包破解**（PMK/PTK 派生 + MIC 校验，支持 GPU 加速） | [`wifi_crack.ys`](test/os/wifi_crack.ys) |
| WPA2 破解流程（英文版 / 结果校验） | [`wifi_verify.ys`](test/os/wifi_verify.ys) |
| WPA2 字典破解性能测试 | [`wifi_speed.ys`](test/os/wifi_speed.ys) |
| **服务端口扫描与指纹识别** | [`service_scanner.ys`](service_scanner.ys) |
| 网络协议与 socket 实战 | [`net_ext.ys`](test/os/net_ext.ys) |
| 二进制格式解析与修补 | [`binary_lib.ys`](test/os/binary_lib.ys) |

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
| **密码学** | `crypto` 35 个函数：SHA/SHA3/SHAKE、AES-GCM、ChaCha20、HMAC、PBKDF2、ECDSA、Ed25519、**WPA2 全套** |
| **协议分析** | `binary` 99 个函数：大小端编解码、字节容器、流抽象、**等长安全修补** |
| **网络** | `socket` 统一 TCP/UDP/TLS 对象；`net` 底层连接；`raw` 原始帧 + **pcap 抓包**；`ssl` 证书与 TLS |
| **系统与 Shell** | 反引号直接执行命令；`os` 进程管理、`sys` 系统信息、`path` 路径处理 |
| **并发** | `warp` 轻量协程 + `sync` 全套同步原语（Mutex/RWMutex/Chan/TLS/WaitGroup/Once/WorkerPool） |
| **异常** | `expr catch h` 一行兜底、`catch {} match { T(e) -> }` 按类型分流、`ensure` 保证清理 |
| **面向对象** | `struct` 方法、接口多态、继承多态（`extends` + `super`）、泛型注解 |
| **GPU 加速** | `cuda` 命名空间支持密码学运算批处理，无 GPU 时自动降级 CPU |
| **FFI** | `ffi` 真实 ABI 调用动态库（0-8 参数，int/string/void 返回） |
| **工程** | 预处理器、REPL 会话、`.ybc` 字节码落盘、交叉编译、中英双语 i18n |

### v0.1.5.1 更新

- **HTTP**：新增 `http.Server`、状态码释义、可配置重试退避、响应后钩子，以及会话的表单、上传、下载和报文辅助方法。
- **本地站点**：`net.nginx()` 可托管静态目录，默认查找 `index.html`，也可通过 `php-cgi` 执行 PHP 页面。
- **CLI**：新增 `cli` 参数框架，支持长短选项、组合短开关、位置参数和彩色帮助。
- **密码学**：补充 SHA-224/384、SHA-512/224/256、SHA3-224、BLAKE2s/BLAKE2b 与 SHA3 HMAC，并扩展 `crypto.HashFile()` 算法名支持。

```yscript
import ["cli", "color"]

let app = cli.New("scan", "端口扫描工具")
app.flag("verbose", "v", false, "显示详细信息")
app.option("target", "t", "127.0.0.1", "目标地址")
app.color(color.Cyan())
let opts = app.parse(args)
```

静态站点示例：

```yscript
import ["net"]
let site = net.nginx("./public", ":8080")
println("listening on " + site.addr)
```

PHP 页面需在运行环境安装 `php-cgi`，或在 `net.nginx` options 中指定 `php_cgi` 路径；未配置 CGI 时 PHP 请求返回 501，不会以静态文件方式暴露源码。

---

## 快速开始

本仓库不含可执行文件，需先安装 `ysc`（YScript 解释器）：

```bash
ysc hello.ys             # 执行脚本
ysc                       # 无参数 = 进入 REPL 会话
ysc -e 'println(1+1)'     # 直接执行一段代码
ysc -c hello.ys           # 仅做语法检查，不执行
ysc --emit a.ybc x.ys     # 导出字节码
ysc -h                    # 查看全部选项
```

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

一个可读的完整实战示例：`init()` 中建立端口 → 服务名映射表，运行时结合 `socket` 与 Shell 命令完成端口探测。

```yscript
init() {
    SERVICES.set(21,  "FTP")
    SERVICES.set(22,  "SSH")
    SERVICES.set(80,  "HTTP")
    SERVICES.set(443, "HTTPS")
    # ... 共 100+ 常用服务
}
```

---

## 测试套件（45 个 `.ys` 示例）

`test/` 下的脚本既是**回归测试**，也是**按特性组织的语法示例**，可直接阅读学习。

| 特性 | 文件 |
|------|------|
| 入口 / 总览 | [`main.ys`](test/os/main.ys) · [`comprehensive.ys`](test/os/comprehensive.ys) · [`features.ys`](test/os/features.ys) |
| 继承多态 | [`poly.ys`](test/os/poly.ys) |
| 异常捕获 | [`exception_catch.ys`](test/os/exception_catch.ys) · [`error_codes.ys`](test/os/error_codes.ys) · [`safe.ys`](test/os/safe.ys) |
| 切片 | [`slice.ys`](test/os/slice.ys) |
| 二进制 | [`binary_lib.ys`](test/os/binary_lib.ys) · [`memory.ys`](test/os/memory.ys) |
| 网络 | [`net_ext.ys`](test/os/net_ext.ys) · [`http_ext.ys`](test/os/http_ext.ys) |
| 并发 | [`sync.ys`](test/os/sync.ys) |
| 面向对象 | [`struct.ys`](test/os/struct.ys) · [`interface.ys`](test/os/interface.ys) · [`enum.ys`](test/os/enum.ys) · [`generics.ys`](test/os/generics.ys) |
| 集合 / 迭代 | [`rf2.ys`](test/os/rf2.ys) |
| 文件与系统 | [`sys.ys`](test/os/sys.ys) · [`shell_ext.ys`](test/os/shell_ext.ys) · [`io_ext.ys`](test/os/io_ext.ys) |
| 文本 / 时间 | [`time_lib.ys`](test/os/time_lib.ys) |
| 配置解析 | [`toml_ext.ys`](test/os/toml_ext.ys) · [`ini_ext.ys`](test/os/ini_ext.ys) · [`yaml_ext.ys`](test/os/yaml_ext.ys) |
| CLI / HTTP | [`cli.ys`](test/cli/cli.ys) · [`tcp_send.ys`](test/cli/tcp_send.ys) |
| 密码学 | [`crypto_hash.ys`](test/crypto/crypto_hash.ys) |
| 文本处理 | [`strings_flexible.ys`](test/strings/strings_flexible.ys) |
| WiFi 实战 | [`wifi_crack.ys`](test/os/wifi_crack.ys) · [`wifi_verify.ys`](test/os/wifi_verify.ys) · [`wifi_speed.ys`](test/os/wifi_speed.ys) |

### 代码示例

**继承多态**（[`test/os/poly.ys`](test/os/poly.ys)）—— Java 风格的 `extends` / 方法重写 / `super` 复用：

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

**带类型异常分流**（[`test/os/exception_catch.ys`](test/os/exception_catch.ys)）—— `match` 按异常类型分支，`ensure` 保证清理：

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
| 文件 / IO | `io` `path` `stdio` `os` |
| 序列化 | `csv` `xml` `yaml` `toml` `ini` |
| 网络 | `socket` `net` `raw` `ssl` `http` `url` |
| 命令行 | `cli` |
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

📖 **在线阅读（推荐）**：**[YScript Wiki →](https://github.com/xiguayiqiu/YScript/wiki)** —— 共 **51 篇**中文手册，可直接浏览与搜索。

完整手册同时以 Markdown 形式存放于开发仓库的 `doc/` 目录（含按命名空间分类的**函数速查表**：37 命名空间 + 6 类型 / 729 个函数）：

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

当前版本 **v0.1.5.1**。完整变更记录见开发仓库的 `doc/Log.md`。

## 许可

Apache-2.0，见 [LICENSE](LICENSE)。
