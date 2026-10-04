package std

import (
	"sync"
	"testing"

	"yscript/internal/value"
)

// call 通过命名空间表取出函数并调用，模拟脚本层的调用路径。
func call(t *testing.T, name string, args ...value.Value) value.Value {
	t.Helper()
	fnVal, ok := BinaryNS.DictGet(name)
	if !ok {
		t.Fatalf("binary.%s 未注册", name)
	}
	if fnVal.Typ != value.TypeBuiltin || fnVal.Blt == nil {
		t.Fatalf("binary.%s 不是内置函数", name)
	}
	v, err := fnVal.Blt.Fn(args)
	if err != nil {
		t.Fatalf("binary.%s 返回错误: %v", name, err)
	}
	return v
}

// callErr 调用并断言返回错误。
func callErr(t *testing.T, name string, args ...value.Value) {
	t.Helper()
	fnVal, ok := BinaryNS.DictGet(name)
	if !ok {
		t.Fatalf("binary.%s 未注册", name)
	}
	if _, err := fnVal.Blt.Fn(args); err == nil {
		t.Fatalf("binary.%s 应当返回错误，但没有", name)
	}
}

func bs(s string) value.Value { return value.NewBytes([]byte(s)) }
func is(v value.Value) int64  { n, _ := v.ToInt(); return n }

func TestPadLeftAndPadRightAreDistinct(t *testing.T) {
	// 回归测试：PadLeft 曾与 PadRight 实现完全相同（复制粘贴缺陷）。
	left := call(t, "PadLeft", bs("AB"), value.NewInt(4), value.NewInt(0x30)).String()
	right := call(t, "PadRight", bs("AB"), value.NewInt(4), value.NewInt(0x30)).String()
	if left != "00AB" {
		t.Errorf("PadLeft = %q, 期望 %q", left, "00AB")
	}
	if right != "AB00" {
		t.Errorf("PadRight = %q, 期望 %q", right, "AB00")
	}
}

func TestEOFDoesNotConsume(t *testing.T) {
	// 回归测试：EOF 曾通过 Read 探测，导致吞掉一个字节。
	r := is(call(t, "NewReader", bs("ABC")))
	if call(t, "EOF", value.NewInt(r)).Bool {
		t.Fatal("非空 reader 不应报 EOF")
	}
	if got := call(t, "ReadAll", value.NewInt(r)).String(); got != "ABC" {
		t.Errorf("EOF 后数据被破坏: %q, 期望 %q", got, "ABC")
	}
	if !call(t, "EOF", value.NewInt(r)).Bool {
		t.Error("读完后应报 EOF")
	}
	call(t, "CloseReader", value.NewInt(r))
}

func TestBufferSeekRoundTrip(t *testing.T) {
	b := is(call(t, "NewBuffer", bs("0123456789")))
	if got := call(t, "BufferRead", value.NewInt(b), value.NewInt(3)).String(); got != "012" {
		t.Errorf("BufferRead = %q", got)
	}
	// 回绕到开头后可重读（bytes.Buffer 无法回退，故使用显式游标）。
	call(t, "BufferSeek", value.NewInt(b), value.NewInt(0), value.NewInt(0))
	if got := call(t, "BufferRead", value.NewInt(b), value.NewInt(3)).String(); got != "012" {
		t.Errorf("回绕后重读 = %q, 期望 %q", got, "012")
	}
	// whence=2 表示相对末尾。
	call(t, "BufferSeek", value.NewInt(b), value.NewInt(-3), value.NewInt(2))
	if got := call(t, "BufferRead", value.NewInt(b), value.NewInt(3)).String(); got != "789" {
		t.Errorf("末尾寻址读取 = %q, 期望 %q", got, "789")
	}
	call(t, "CloseBuffer", value.NewInt(b))
}

func TestBufferWriteAppends(t *testing.T) {
	// 写入必须追加到末尾（bytes.Buffer 语义），而非覆盖读游标处。
	b := is(call(t, "NewBuffer", bs("init data")))
	call(t, "BufferWrite", value.NewInt(b), bs(" more"))
	call(t, "BufferSeek", value.NewInt(b), value.NewInt(0), value.NewInt(0))
	if got := call(t, "BufferRead", value.NewInt(b), value.NewInt(4)).String(); got != "init" {
		t.Errorf("BufferRead = %q, 期望 %q", got, "init")
	}
	call(t, "BufferSeek", value.NewInt(b), value.NewInt(-5), value.NewInt(2))
	if got := call(t, "BufferRead", value.NewInt(b), value.NewInt(5)).String(); got != " more" {
		t.Errorf("末尾读取 = %q, 期望 %q", got, " more")
	}
	call(t, "CloseBuffer", value.NewInt(b))
}

func TestBufferBytesReturnsCopy(t *testing.T) {
	// 回归测试：BufferBytes 曾直接返回内部切片。
	b := is(call(t, "NewBuffer", bs("abc")))
	snap := call(t, "BufferBytes", value.NewInt(b))
	call(t, "BufferReset", value.NewInt(b))
	if snap.String() != "abc" {
		t.Errorf("快照被后续写入破坏: %q", snap.String())
	}
	call(t, "CloseBuffer", value.NewInt(b))
}

func TestWriterTypedWrites(t *testing.T) {
	w := is(call(t, "NewWriter"))
	call(t, "WriteString", value.NewInt(w), value.NewString("AB"))
	call(t, "WriteUint16", value.NewInt(w), value.NewInt(0x0102))
	call(t, "WriteUint32", value.NewInt(w), value.NewInt(0x01020304))
	call(t, "WriteInt8", value.NewInt(w), value.NewInt(-1))
	data := call(t, "WriterBytes", value.NewInt(w))
	// String 返回 writer 全部内容的 UTF-8 解释（含二进制尾部）。
	if got := len(call(t, "String", value.NewInt(w)).String()); got != 9 {
		t.Errorf("String(w) 长度 = %d, 期望 9", got)
	}
	if got := is(call(t, "BytesWritten", value.NewInt(w))); got != 9 {
		t.Errorf("已写入 %d 字节, 期望 9", got)
	}
	if got := is(call(t, "Uint16", call(t, "Slice", data, value.NewInt(2), value.NewInt(4)))); got != 0x0102 {
		t.Errorf("Uint16 = %#x", got)
	}
	if got := is(call(t, "Uint32", call(t, "Slice", data, value.NewInt(4), value.NewInt(8)))); got != 0x01020304 {
		t.Errorf("Uint32 = %#x", got)
	}
	if got := is(call(t, "Int8", call(t, "Slice", data, value.NewInt(8), value.NewInt(9)))); got != -1 {
		t.Errorf("Int8 = %d", got)
	}
	call(t, "CloseWriter", value.NewInt(w))
}

func TestEndiannessRoundTrip(t *testing.T) {
	call(t, "BigEndian")
	if got := call(t, "PutUint16", value.NewInt(0x0102)).String(); got != "\x01\x02" {
		t.Errorf("大端 PutUint16 = %q", got)
	}
	call(t, "LittleEndian")
	if got := call(t, "PutUint16", value.NewInt(0x0102)).String(); got != "\x02\x01" {
		t.Errorf("小端 PutUint16 = %q", got)
	}
	if got := is(call(t, "Uint16", call(t, "PutUint16", value.NewInt(0x0102)))); got != 0x0102 {
		t.Errorf("小端往返 = %#x", got)
	}
	call(t, "BigEndian")
}

func TestUint8Int8(t *testing.T) {
	if got := is(call(t, "Uint8", call(t, "PutUint8", value.NewInt(0x41)))); got != 65 {
		t.Errorf("PutUint8/Uint8 = %d", got)
	}
	if got := is(call(t, "Int8", call(t, "PutInt8", value.NewInt(-1)))); got != -1 {
		t.Errorf("PutInt8/Int8 = %d", got)
	}
}

func TestBitwise(t *testing.T) {
	a, b := bs("\x0f\xf0"), bs("\xff\x00")
	for _, c := range []struct{ fn, want string }{
		{"And", "\x0f\x00"},
		{"Or", "\xff\xf0"},
		{"Xor", "\xf0\xf0"},
	} {
		if got := call(t, c.fn, a, b).String(); got != c.want {
			t.Errorf("%s = %q, 期望 %q", c.fn, got, c.want)
		}
	}
	if got := call(t, "Not", a).String(); got != "\xf0\x0f" {
		t.Errorf("Not = %q", got)
	}
	// 位运算要求等长。
	callErr(t, "And", a, bs("\x00"))
}

// TestHugeAllocDoesNotPanic 回归测试：
// 极端尺寸曾触发 Go 的 "makeslice: len out of range" panic，
// 该 panic 无法被脚本层 try/catch 捕获，会直接杀死整个解释器。
func TestHugeAllocDoesNotPanic(t *testing.T) {
	const huge = 1<<63 - 1
	// 这些调用必须返回 error（可被脚本捕获），而不是 panic。
	fnVal, _ := BinaryNS.DictGet("PadLeft")
	if _, err := fnVal.Blt.Fn([]value.Value{bs("abcdef"), value.NewInt(huge)}); err == nil {
		t.Error("PadLeft 极端尺寸应当报错")
	}
	fnVal, _ = BinaryNS.DictGet("PadRight")
	if _, err := fnVal.Blt.Fn([]value.Value{bs("abcdef"), value.NewInt(huge)}); err == nil {
		t.Error("PadRight 极端尺寸应当报错")
	}
	fnVal, _ = BinaryNS.DictGet("Repeat")
	if _, err := fnVal.Blt.Fn([]value.Value{bs("abcdef"), value.NewInt(huge)}); err == nil {
		t.Error("Repeat 极端次数应当报错")
	}
	fnVal, _ = BinaryNS.DictGet("Grow")
	w := is(call(t, "NewWriter"))
	if _, err := fnVal.Blt.Fn([]value.Value{value.NewInt(w), value.NewInt(huge)}); err == nil {
		t.Error("Grow 极端尺寸应当报错")
	}
	// Read 应夹取到剩余字节数，而不是尝试分配超大缓冲。
	r := is(call(t, "NewReader", bs("abcdef")))
	v := call(t, "Read", value.NewInt(r), value.NewInt(huge))
	if v.String() != "abcdef" {
		t.Errorf("Read 极端长度应返回全部剩余字节, 实际 %q", v.String())
	}
	call(t, "CloseReader", value.NewInt(r))
	call(t, "CloseWriter", value.NewInt(w))
}

// TestNegativeAllocRejected 确认负数尺寸被拒绝（而非静默产生怪结果）。
func TestNegativeAllocRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []value.Value
	}{
		{"PadLeft", []value.Value{bs("a"), value.NewInt(-1)}},
		{"PadRight", []value.Value{bs("a"), value.NewInt(-1)}},
		{"Repeat", []value.Value{bs("a"), value.NewInt(-1)}},
		{"Chunk", []value.Value{bs("a"), value.NewInt(-1)}},
		{"Read", []value.Value{value.NewInt(-1), value.NewInt(-1)}},
		{"BufferRead", []value.Value{value.NewInt(-1), value.NewInt(-1)}},
	} {
		fnVal, ok := BinaryNS.DictGet(tc.name)
		if !ok {
			continue
		}
		if _, err := fnVal.Blt.Fn(tc.args); err == nil {
			t.Errorf("%s 对负数尺寸应当报错", tc.name)
		}
	}
}

// TestPatchEqualGuards 验证 PatchEqual 的三层防护。
func TestPatchEqualGuards(t *testing.T) {
	data := []byte("xxxxHello World!yyyy")
	off := int64(4)

	// 1. 正常等长修补
	fn, _ := BinaryNS.DictGet("PatchEqual")
	oldStr, newStr := []byte("Hello World!"), []byte("Hello PATCH!")
	n := len(oldStr)
	out, err := fn.Blt.Fn([]value.Value{
		value.NewBytes(data), value.NewInt(off),
		value.NewBytes(oldStr), value.NewBytes(newStr),
	})
	if err != nil {
		t.Fatalf("等长修补不应报错: %v", err)
	}
	if len(out.B) != len(data) {
		t.Errorf("结果长度 %d != 原始 %d", len(out.B), len(data))
	}
	if string(out.B[off:off+int64(n)]) != string(newStr) {
		t.Errorf("修补内容 = %q", out.B[off:off+int64(n)])
	}
	// 前缀/后缀必须保持不变。
	if string(out.B[:off]) != "xxxx" || string(out.B[off+int64(n):]) != "yyyy" {
		t.Error("修补影响了区间之外的字节")
	}
	// 输入按值语义处理，不得被修改。
	if string(data) != "xxxxHello World!yyyy" {
		t.Errorf("输入被就地修改: %q", data)
	}

	call := func(args ...value.Value) error {
		_, err := fn.Blt.Fn(args)
		return err
	}
	// 2. 变长必须拒绝（这是防止二进制损坏的核心约束）
	if err := call(value.NewBytes(data), value.NewInt(off),
		value.NewBytes([]byte("Hello World!")), value.NewBytes([]byte("longer!"))); err == nil {
		t.Error("变长替换应当被拒绝")
	}
	if err := call(value.NewBytes(data), value.NewInt(off),
		value.NewBytes([]byte("Hello World!")), value.NewBytes([]byte("s"))); err == nil {
		t.Error("变短替换应当被拒绝")
	}
	// 3. 空 old 必须拒绝
	if err := call(value.NewBytes(data), value.NewInt(off),
		value.NewBytes(nil), value.NewBytes(nil)); err == nil {
		t.Error("空 old 应当被拒绝")
	}
	// 4. 越界
	if err := call(value.NewBytes(data), value.NewInt(-1),
		value.NewBytes([]byte("x")), value.NewBytes([]byte("y"))); err == nil {
		t.Error("负偏移应当被拒绝")
	}
	if err := call(value.NewBytes(data), value.NewInt(int64(len(data))),
		value.NewBytes([]byte("x")), value.NewBytes([]byte("y"))); err == nil {
		t.Error("超出末尾的偏移应当被拒绝")
	}
	if err := call(value.NewBytes(data), value.NewInt(int64(len(data)-5)),
		value.NewBytes([]byte("Hello World!")), value.NewBytes([]byte("Hello PATCH!"))); err == nil {
		t.Error("区间超出末尾应当被拒绝")
	}
	// 5. 偏移错误时内容校验必须失败（防止覆盖无关字节）
	if err := call(value.NewBytes(data), value.NewInt(0),
		value.NewBytes([]byte("Hello World!")), value.NewBytes([]byte("Hello PATCH!"))); err == nil {
		t.Error("偏移处内容不匹配时应当被拒绝")
	}
	if err := call(value.NewBytes(data), value.NewInt(off+1),
		value.NewBytes([]byte("Hello World!")), value.NewBytes([]byte("Hello PATCH!"))); err == nil {
		t.Error("偏移差 1 时应当被拒绝")
	}
	// 6. 极值不得 panic
	if err := call(value.NewBytes(data), value.NewInt(1<<63-1),
		value.NewBytes([]byte("x")), value.NewBytes([]byte("y"))); err == nil {
		t.Error("极大偏移应当返回错误而非 panic")
	}
}

// TestPatchEqualPinsStringTable 模拟 Go 字符串表布局：
// 等长修补必须保持相邻字符串完好，变长则必然破坏它们。
func TestPatchEqualPinsStringTable(t *testing.T) {
	// Go 会把字符串常量紧密排列：Hello World! | short | write
	blob := []byte("Hello World!shortwrite")
	off := int64(0)

	fn, _ := BinaryNS.DictGet("PatchEqual")
	out, err := fn.Blt.Fn([]value.Value{
		value.NewBytes(blob), value.NewInt(off),
		value.NewBytes([]byte("Hello World!")), value.NewBytes([]byte("Hello PATCH!")),
	})
	if err != nil {
		t.Fatalf("等长修补失败: %v", err)
	}
	if len(out.B) != len(blob) {
		t.Fatalf("长度变化: %d -> %d", len(blob), len(out.B))
	}
	got := string(out.B)
	if got != "Hello PATCH!shortwrite" {
		t.Errorf("相邻字符串被破坏: %q", got)
	}
}

// TestContainerOps 覆盖此前缺失的容器函数。
func TestContainerOps(t *testing.T) {
	if got := is(call(t, "RFind", bs("abracadabra"), bs("abra"))); got != 7 {
		t.Errorf("RFind = %d, 期望 7", got)
	}
	if got := call(t, "ReplaceAll", bs("banana"), bs("a"), bs("o")).String(); got != "bonono" {
		t.Errorf("ReplaceAll = %q", got)
	}
	if got := len(call(t, "Split", bs("a|b|c"), bs("|")).List.Items); got != 3 {
		t.Errorf("Split = %d 段", got)
	}
	if got := call(t, "Truncate", bs("hello world"), value.NewInt(5)).String(); got != "hello" {
		t.Errorf("Truncate = %q", got)
	}
}

func TestSliceDoesNotAlias(t *testing.T) {
	// Slice 不应与源数据共享底层数组。
	src := []byte("abcdef")
	sl := call(t, "Slice", value.NewBytes(src), value.NewInt(0), value.NewInt(3))
	sl.B[0] = 'X'
	if src[0] != 'a' {
		t.Errorf("Slice 结果与源共享内存: src[0]=%q", src[0])
	}
}

func TestBase64Variants(t *testing.T) {
	if got := call(t, "Base64Encode", bs("AB")).String(); got != "QUI=" {
		t.Errorf("Base64Encode = %q", got)
	}
	// 缺少填充也能解码（手写实现曾静默接受任意长度）。
	if got := call(t, "Base64Decode", value.NewString("3q2+7w")).String(); got != "\xde\xad\xbe\xef" {
		t.Errorf("无填充解码 = %q", got)
	}
	// URL-safe 字母表。
	url := call(t, "Base64Encode", bs("\xfb\xff"), value.NewBool(true)).String()
	if url != "-_8" {
		t.Errorf("URL-safe 编码 = %q", url)
	}
	if got := call(t, "Base64Decode", value.NewString(url)).String(); got != "\xfb\xff" {
		t.Errorf("URL-safe 解码 = %q", got)
	}
	callErr(t, "Base64Decode", value.NewString("!!!!"))
}

func TestHexTolerance(t *testing.T) {
	if got := call(t, "HexEncode", bs("\x01\x02"), value.NewString(" ")).String(); got != "01 02" {
		t.Errorf("带分隔符 HexEncode = %q", got)
	}
	// 容忍抓包工具常见的分隔符与空白。
	if got := call(t, "HexDecode", value.NewString("de:ad be ef")).String(); got != "\xde\xad\xbe\xef" {
		t.Errorf("含噪 HexDecode = %q", got)
	}
	// 奇数长度前置补 0。
	if got := call(t, "HexDecode", value.NewString("abc")).String(); got != "\x0a\xbc" {
		t.Errorf("奇数长度 HexDecode = %q", got)
	}
}

func TestErrorPaths(t *testing.T) {
	callErr(t, "Chunk", bs("abc"), value.NewInt(0))
	callErr(t, "Repeat", bs("A"), value.NewInt(-1))
	callErr(t, "PadLeft", bs("A"), value.NewInt(-1))
	callErr(t, "PadRight", bs("A"), value.NewInt(-1))
	callErr(t, "Split", bs("abc"), bs(""))
	callErr(t, "Uint16", bs("A"))
	callErr(t, "Read", value.NewInt(99999), value.NewInt(1))
	callErr(t, "Write", value.NewInt(99999), bs("x"))
	callErr(t, "BufferRead", value.NewInt(99999), value.NewInt(1))
	callErr(t, "FileRead", value.NewInt(99999), value.NewInt(1))
}

func TestRepeatPreallocation(t *testing.T) {
	if got := call(t, "Repeat", bs("ab"), value.NewInt(3)).String(); got != "ababab" {
		t.Errorf("Repeat = %q", got)
	}
}

func TestFileRoundTrip(t *testing.T) {
	path := t.TempDir() + "/bin_test.bin"
	call(t, "WriteFile", value.NewString(path), bs("hello"))
	if got := call(t, "ReadFile", value.NewString(path)).String(); got != "hello" {
		t.Errorf("ReadFile = %q", got)
	}
	st := call(t, "FileStat", value.NewString(path))
	if got := is(st.Dict.Map["size"]); got != 5 {
		t.Errorf("FileStat size = %d", got)
	}
	f := is(call(t, "Open", value.NewString(path)))
	if got := call(t, "FileRead", value.NewInt(f), value.NewInt(2)).String(); got != "he" {
		t.Errorf("FileRead = %q", got)
	}
	call(t, "FileClose", value.NewInt(f))
	// 重复关闭应幂等，不报错。
	call(t, "FileClose", value.NewInt(f))
}

func TestAppendFile(t *testing.T) {
	path := t.TempDir() + "/append.bin"
	call(t, "WriteFile", value.NewString(path), bs("ab"))
	call(t, "AppendFile", value.NewString(path), bs("cd"))
	if got := call(t, "ReadFile", value.NewString(path)).String(); got != "abcd" {
		t.Errorf("AppendFile 后 = %q, 期望 %q", got, "abcd")
	}
}

// TestAllNewFunctionsRegistered 防止新增函数遗漏注册。
func TestAllNewFunctionsRegistered(t *testing.T) {
	for _, n := range []string{
		"PutUint8", "Uint8", "Int8", "PutInt8",
		"RFind", "ReplaceAll", "Split", "Truncate",
		"And", "Or", "Xor", "Not", "PatchEqual",
		"BufferSeek", "BufferString",
		"WriterBytes", "String", "WriteString",
		"WriteUint8", "WriteUint16", "WriteUint32", "WriteUint64",
		"WriteInt8", "WriteInt16", "WriteInt32", "WriteInt64",
		"WriteFloat32", "WriteFloat64",
		"Seek", "Len",
		"FileStat", "FileSync", "AppendFile", "OpenMode",
		"CloseReader", "CloseWriter", "CloseBuffer",
	} {
		if _, ok := BinaryNS.DictGet(n); !ok {
			t.Errorf("binary.%s 应已注册", n)
		}
	}
}

// TestConcurrentAccess 配合 -race：验证注册表与字节序状态在并发下安全。
func TestConcurrentAccess(t *testing.T) {
	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				switch j % 6 {
				case 0:
					r := is(call(t, "NewReader", bs("abcdef")))
					call(t, "Read", value.NewInt(r), value.NewInt(3))
					call(t, "EOF", value.NewInt(r))
					call(t, "Len", value.NewInt(r))
					call(t, "CloseReader", value.NewInt(r))
				case 1:
					wr := is(call(t, "NewWriter"))
					call(t, "Write", value.NewInt(wr), bs("xyz"))
					call(t, "WriterBytes", value.NewInt(wr))
					call(t, "CloseWriter", value.NewInt(wr))
				case 2:
					bf := is(call(t, "NewBuffer", bs("0123456789")))
					call(t, "BufferRead", value.NewInt(bf), value.NewInt(2))
					call(t, "BufferSeek", value.NewInt(bf), value.NewInt(0), value.NewInt(0))
					call(t, "BufferBytes", value.NewInt(bf))
					call(t, "CloseBuffer", value.NewInt(bf))
				case 3:
					// 字节序是全局状态，并发切换不应崩溃。
					if i%2 == 0 {
						call(t, "LittleEndian")
					} else {
						call(t, "BigEndian")
					}
					call(t, "PutUint32", value.NewInt(0x01020304))
				case 4:
					call(t, "Concat", bs("a"), bs("b"), bs("c"))
					call(t, "And", bs("\x0f"), bs("\xff"))
					call(t, "HexEncode", bs("\x01\x02"))
				case 5:
					call(t, "Base64Encode", bs("hello"))
					call(t, "Truncate", bs("hello"), value.NewInt(3))
				}
			}
		}(i)
	}
	wg.Wait()
}
