# YScript 脚本语言

YScript 是一门面向网络安全领域的脚本语言，提供丰富的内置网络、系统、编码和加密函数库，用于快速编写端口扫描、数据包分析、漏洞验证、日志处理等安全工具。

## 特性

- **内置网络库** — TCP/UDP 连接、端口扫描、HTTP 请求、SSL/TLS、CIDR 子网计算
- **系统交互** — 进程执行、Shell 命令（反引号）、文件 IO、目录遍历、符号链接
- **编码与加密** — Base64、Hex、JSON、XML、CSV、AES、RSA、Gzip 压缩
- **并发编程** — warp 多线程、Mutex/RWMutex、Channel、信号量、原子操作
- **内存与指针** — `@` 取地址、`@>` 解引用、`alloc`/`free`
- **空安全** — `?.` 安全访问、`??` nil 合并
- **类型系统** — byte/short/int/long/float/double、string、bytes、list、dict、struct、enum、interface、泛型
- **二进制与字节** — `b"..."` 字节字面量、binary 编解码、hexdump
- **正则表达式** — 内置 regex 匹配与提取
- **编译执行** — 直接编译为字节码在轻量级 VM 上运行

## 演示

[完整演示文件](service_scanner.ys)
[二进制处理](test/binary_lib.ys)

## 插件

- **Vim/Nvim**：[Vim/Nvim插件](https://github.com/xiguayiqiu/YScript-vim)
- **VScode**：[VScode插件](https://github.com/xiguayiqiu/YScript-vscode)

## 快速开始

```bash
yscript hello.ys
yscript -e 'println("hello")'
```

## 语法示例

### 端口扫描

```yscript
let conn = net.DialTimeout("192.168.1.1:80", 2)
if conn != nil {
    printf("[+] port 80 open\n")
}
```

### 并发 warp

```yscript
let w = warp {
    println("hello from warp")
}
w.await()
```

### HTTP 请求

```yscript
let resp = net.HTTPGet("https://example.com")
if resp.ok {
    println(resp.body)
}
```

### 字节与编码

```yscript
let b = b"\x90\x90\x90"
let e = encoding.base64_encode(b)
let d = encoding.base64_decode(e)
```

### 空安全

```yscript
let data = nil
let host = data?.host ?? "localhost"
```

### 加密

```yscript
let key = "0123456789abcdef0123456789abcdef"
let encrypted = crypto.aes_encrypt(key, "secret data")
let decrypted = crypto.aes_decrypt(key, encrypted)
```

## 内置模块

| 模块         | 功能                                      |
| ---------- | --------------------------------------- |
| `net`      | TCP/UDP 连接、端口扫描、DNS 解析、CIDR、SSL/TLS     |
| `net.HTTP` | HTTP GET/POST 请求                        |
| `io`       | 文件读写、临时文件、chmod、walk、glob、symlink       |
| `os`       | 进程执行、shell、hostname、getpid              |
| `encoding` | base64、hex 编解码                          |
| `json`     | JSON 解析与序列化                             |
| `regex`    | 正则匹配与提取                                 |
| `compress` | gzip 压缩/解压                              |
| `crypto`   | AES 加密/解密、RSA                           |
| `binary`   | 大端/小端编解码、PutUint16/32/64                |
| `sync`     | Mutex、RWMutex、Atomic、Chan、Semaphore、TLS |
| `sys`      | CPU 信息、OS 信息、当前用户、网络接口                  |
| `time`     | 时间戳、格式化、解析、持续时间                         |
| `path`     | 路径 join、basename、ext                    |
| `array`    | 排序、查找                                   |
| `log`      | 分级日志输出                                  |
| `stdio`    | 交互式提示输入                                 |
| `rand`     | 随机数、UUID 生成                             |
| `ffi`      | C 语言外部函数调用                              |
| `color`    | ANSI 终端颜色输出                             |
| `raw`      | 原始 socket 操作                            |
| `reflect`  | 运行时类型反射                                 |
| `errors`   | 错误码定义与匹配                                |

## 安装

```bash
git clone https://github.com/your/project
cd yscript
go build -o yscript ./cmd/yscript/
sudo cp yscript /usr/local/bin/
```

## 测试

```bash
cd test
yscript main.ys
```

## 编辑器支持

**目前支持** — vim、vscode
