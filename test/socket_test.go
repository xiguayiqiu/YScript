package std

import (
	"testing"
	"time"

	"yscript/internal/value"
)

func newSock(t *testing.T, proto string) value.Value {
	t.Helper()
	v, err := socketNew([]value.Value{value.NewString(proto)})
	if err != nil {
		t.Fatalf("socket.Socket(%q) 失败: %v", proto, err)
	}
	return v
}

// TestSocketNewRejectsBadProtocol 非法协议必须被拒绝。
func TestSocketNewRejectsBadProtocol(t *testing.T) {
	for _, p := range []string{"sctp", "http", "", "TCP/UDP"} {
		if _, err := socketNew([]value.Value{value.NewString(p)}); err == nil {
			t.Errorf("协议 %q 应当被拒绝", p)
		}
	}
	for _, p := range []string{"tcp", "udp", "tls", "TCP", "UDP"} {
		if _, err := socketNew([]value.Value{value.NewString(p)}); err != nil {
			t.Errorf("协议 %q 应当被接受: %v", p, err)
		}
	}
}

// TestSocketTypedErrors 错误必须带正确的异常类型，
// 否则 catch ... match 无法按类型分流。
func TestSocketTypedErrors(t *testing.T) {
	cases := []struct {
		name string
		run  func() error
		want string
	}{
		{"协议非法", func() error {
			_, err := socketNew([]value.Value{value.NewString("sctp")})
			return err
		}, "SocketError"},
		{"未连接时 send", func() error {
			_, err := socketSend([]value.Value{newSock(t, "tcp"), value.NewString("x")})
			return err
		}, "SocketError"},
		{"未连接时 recv", func() error {
			_, err := socketRecv([]value.Value{newSock(t, "tcp"), value.NewInt(10)})
			return err
		}, "SocketError"},
		{"端口越界", func() error {
			_, err := socketConnect([]value.Value{
				newSock(t, "tcp"), value.NewString("h"), value.NewInt(99999)})
			return err
		}, "SocketError"},
		{"TCP 上 bind", func() error {
			_, err := socketBind([]value.Value{
				newSock(t, "tcp"), value.NewString("127.0.0.1"), value.NewInt(0)})
			return err
		}, "SocketError"},
		{"UDP 上 listen", func() error {
			_, err := socketListen([]value.Value{newSock(t, "udp"), value.NewInt(0)})
			return err
		}, "SocketError"},
		{"未监听时 accept", func() error {
			_, err := socketAccept([]value.Value{newSock(t, "tcp")})
			return err
		}, "SocketError"},
		{"未知选项", func() error {
			_, err := socketSetOption([]value.Value{
				newSock(t, "tcp"), value.NewString("bogus"), value.NewBool(true)})
			return err
		}, "SocketError"},
		{"reuse_addr 需建连前", func() error {
			_, err := socketSetOption([]value.Value{
				newSock(t, "tcp"), value.NewString("reuse_addr"), value.NewBool(true)})
			return err
		}, "SocketError"},
		{"recv 长度为负", func() error {
			_, err := socketRecv([]value.Value{newSock(t, "tcp"), value.NewInt(-1)})
			return err
		}, "SocketError"},
		{"recv 长度过大", func() error {
			_, err := socketRecv([]value.Value{
				newSock(t, "tcp"), value.NewInt(socketMaxRecv + 1)})
			return err
		}, "SocketError"},
		{"非 socket 对象", func() error {
			_, err := socketRecv([]value.Value{value.NewInt(1), value.NewInt(10)})
			return err
		}, "SocketError"},
	}
	for _, c := range cases {
		err := c.run()
		if err == nil {
			t.Errorf("%s: 应当返回错误", c.name)
			continue
		}
		te, ok := err.(*value.TypedError)
		if !ok {
			t.Errorf("%s: 错误类型 %T，应为 *value.TypedError", c.name, err)
			continue
		}
		if te.Type != c.want {
			t.Errorf("%s: 异常类型 = %q, 期望 %q", c.name, te.Type, c.want)
		}
	}
}

// TestSocketCloseIdempotent close 必须幂等，且关闭后使用报错而非崩溃。
func TestSocketCloseIdempotent(t *testing.T) {
	s := newSock(t, "tcp")
	if _, err := socketClose([]value.Value{s}); err != nil {
		t.Fatalf("首次 close 应成功: %v", err)
	}
	if _, err := socketClose([]value.Value{s}); err != nil {
		t.Fatalf("重复 close 应幂等: %v", err)
	}
	closed, err := socketIsClosed([]value.Value{s})
	if err != nil || !closed.Bool {
		t.Error("is_closed 应为 true")
	}
	// 关闭后操作必须返回错误而非 panic
	if _, err := socketSend([]value.Value{s, value.NewString("x")}); err == nil {
		t.Error("关闭后 send 应当报错")
	}
	if _, err := socketRecv([]value.Value{s, value.NewInt(10)}); err == nil {
		t.Error("关闭后 recv 应当报错")
	}
	if _, err := socketListen([]value.Value{s, value.NewInt(0)}); err == nil {
		t.Error("关闭后 listen 应当报错")
	}
}

// TestSocketTCPExchange 端到端 TCP 回显。
func TestSocketTCPExchange(t *testing.T) {
	srv := newSock(t, "tcp")
	pv, err := socketListen([]value.Value{srv, value.NewInt(0)})
	if err != nil {
		t.Fatalf("listen 失败: %v", err)
	}
	port := pv.Int
	if port <= 0 {
		t.Fatalf("listen 未返回有效端口: %d", port)
	}

	done := make(chan error, 1)
	go func() {
		cli, aerr := socketAccept([]value.Value{srv})
		if aerr != nil {
			done <- aerr
			return
		}
		line, rerr := socketRecvLine([]value.Value{cli})
		if rerr == nil {
			_, rerr = socketSend([]value.Value{cli, value.NewBytes(line.B)})
		}
		_, _ = socketClose([]value.Value{cli})
		done <- rerr
	}()

	c := newSock(t, "tcp")
	if _, cerr := socketConnect([]value.Value{
		c, value.NewString("127.0.0.1"), value.NewInt(port)}); cerr != nil {
		t.Fatalf("connect 失败: %v", cerr)
	}
	if _, serr := socketSend([]value.Value{c, value.NewString("ping\n")}); serr != nil {
		t.Fatalf("send 失败: %v", serr)
	}
	got, rerr := socketRecvLine([]value.Value{c})
	if rerr != nil {
		t.Fatalf("recv_line 失败: %v", rerr)
	}
	if string(got.B) != "ping\n" {
		t.Errorf("回显 = %q, 期望 %q", got.B, "ping\n")
	}
	_, _ = socketClose([]value.Value{c})
	if werr := <-done; werr != nil {
		t.Errorf("服务端: %v", werr)
	}
	_, _ = socketClose([]value.Value{srv})
}

// TestSocketRecvTimeout 沉默对端应触发 TimeoutError。
func TestSocketRecvTimeout(t *testing.T) {
	srv := newSock(t, "tcp")
	pv, err := socketListen([]value.Value{srv, value.NewInt(0)})
	if err != nil {
		t.Fatalf("listen 失败: %v", err)
	}
	// 服务端 accept 后不发数据也不关闭，触发客户端读超时
	ready := make(chan struct{})
	go func() {
		if cli, aerr := socketAccept([]value.Value{srv}); aerr == nil {
			close(ready)
			// 保持连接直到超时触发
			time.Sleep(300 * time.Millisecond)
			_, _ = socketClose([]value.Value{cli})
		}
	}()

	c := newSock(t, "tcp")
	if _, cerr := socketConnect([]value.Value{
		c, value.NewString("127.0.0.1"), pv}); cerr != nil {
		t.Fatalf("connect 失败: %v", cerr)
	}
	<-ready
	if _, terr := socketSetTimeout([]value.Value{c, value.NewInt(150)}); terr != nil {
		t.Fatalf("set_timeout 失败: %v", terr)
	}
	_, rerr := socketRecv([]value.Value{c, value.NewInt(10)})
	if rerr == nil {
		t.Fatal("沉默对端应触发超时")
	}
	te, ok := rerr.(*value.TypedError)
	if !ok || te.Type != "TimeoutError" {
		t.Errorf("超时异常类型错误: %#v", rerr)
	}
	_, _ = socketClose([]value.Value{c})
	_, _ = socketClose([]value.Value{srv})
}

// TestSocketMethodsRegistered 确保方法表完整。
func TestSocketMethodsRegistered(t *testing.T) {
	for _, m := range []string{
		"connect", "connect_udp", "send", "sendto", "send_addr",
		"recv", "recv_line", "recv_all", "recvfrom", "close",
		"bind", "listen", "accept", "set_option", "set_timeout",
		"set_nonblocking", "set_insecure", "peer_addr", "peer_port",
		"local_addr", "local_port", "start_tls", "is_closed",
	} {
		if _, ok := GetSocketMethod(value.Value{Typ: value.TypeSocket}, m); !ok {
			t.Errorf("socket.%s 未注册", m)
		}
	}
	if _, ok := GetSocketMethod(value.Value{Typ: value.TypeInt}, "connect"); ok {
		t.Error("非 socket 类型不应命中 socket 方法")
	}
	for _, f := range []string{
		"Socket", "tcp_request", "udp_request", "tls_request",
		"tcp_probe", "port_scan", "default_timeout",
	} {
		if _, ok := SocketNS.DictGet(f); !ok {
			t.Errorf("socket.%s 未注册到命名空间", f)
		}
	}
}
