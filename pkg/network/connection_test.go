/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package network

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"mosn.io/mosn/pkg/types"
)

type MyEventListener struct{}

func (el *MyEventListener) OnEvent(event types.ConnectionEvent) {}

func testAddConnectionEventListener(n int, t *testing.T) {
	c := connection{}

	for i := 0; i < n; i++ {
		el0 := &MyEventListener{}
		c.AddConnectionEventListener(el0)
	}

	if len(c.connCallbacks) != n {
		t.Errorf("Expect %d, but got %d after AddConnectionEventListener(el0)", n, len(c.connCallbacks))
	}
}

func testAddBytesReadListener(n int, t *testing.T) {
	c := connection{}

	for i := 0; i < n; i++ {
		fn1 := func(bytesRead uint64) {}
		c.AddBytesReadListener(fn1)
	}

	if len(c.bytesReadCallbacks) != n {
		t.Errorf("Expect %d, but got %d after AddBytesReadListener(fn1)", n, len(c.bytesReadCallbacks))
	}
}

func testAddBytesSendListener(n int, t *testing.T) {
	c := connection{}

	for i := 0; i < n; i++ {
		fn1 := func(bytesSent uint64) {}
		c.AddBytesSentListener(fn1)
	}

	if len(c.bytesSendCallbacks) != n {
		t.Errorf("Expect %d, but got %d after AddBytesSentListener(fn1)", n, len(c.bytesSendCallbacks))
	}
}

func TestAddConnectionEventListener(t *testing.T) {
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("AddConnectionEventListener(%d)", i)
		t.Run(name, func(t *testing.T) {
			testAddConnectionEventListener(i, t)
		})
	}
}

func TestAddBytesReadListener(t *testing.T) {
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("AddBytesReadListener(%d)", i)
		t.Run(name, func(t *testing.T) {
			testAddBytesReadListener(i, t)
		})
	}
}

func TestAddBytesSendListener(t *testing.T) {
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("AddBytesSendListener(%d)", i)
		t.Run(name, func(t *testing.T) {
			testAddBytesSendListener(i, t)
		})
	}
}

func TestConnectTimeout(t *testing.T) {
	timeout := time.Second

	remoteAddr, _ := net.ResolveTCPAddr("tcp", "2.2.2.2:22222")
	conn := NewClientConnection(nil, timeout, nil, remoteAddr, nil)
	begin := time.Now()
	err := conn.Connect()
	if err == nil {
		t.Errorf("connect should timeout")
		return
	}

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		t.Errorf("connect should timeout")
		return
	}

	sub := time.Now().Sub(begin)
	if sub < timeout-10*time.Millisecond {
		t.Errorf("connect should timeout %v, but get %v", timeout, sub)
	}
}

func TestClientConectionRemoteaddrIsNil(t *testing.T) {
	conn := NewClientConnection(nil, 0, nil, nil, nil)
	err := conn.Connect()
	if err == nil {
		t.Errorf("connect should Failed")
		return
	}
}

type zeroReadConn struct {
	net.Conn
}

func (r *zeroReadConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (r *zeroReadConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (r *zeroReadConn) LocalAddr() net.Addr {
	return nil
}

func TestIoBufferZeroRead(t *testing.T) {
	conn := &connection{}
	conn.rawConnection = &zeroReadConn{}
	err := conn.doRead()
	if err != io.EOF {
		t.Errorf("error should be io.EOF")
	}
}

func TestConnState(t *testing.T) {
	testAddr := "127.0.0.1:11234"
	remoteAddr, _ := net.ResolveTCPAddr("tcp", testAddr)
	l, err := net.Listen("tcp", testAddr)
	if err != nil {
		t.Logf("listen error %v", err)
		return
	}
	rawc, err := net.Dial("tcp", testAddr)
	if err != nil {
		t.Logf("net.Dial error %v", err)
		return
	}
	c := NewServerConnection(context.Background(), rawc, nil)
	if c.State() != types.ConnActive {
		t.Errorf("ConnState should be ConnActive")
	}
	c.Close(types.NoFlush, types.LocalClose)
	if c.State() != types.ConnClosed {
		t.Errorf("ConnState should be ConnClosed")
	}

	cc := NewClientConnection(nil, 0, nil, remoteAddr, nil)
	if cc.State() != types.ConnInit {
		t.Errorf("ConnState should be ConnInit")
	}
	if err := cc.Connect(); err != nil {
		t.Errorf("conn Connect error: %v", err)
	}
	if cc.State() != types.ConnActive {
		t.Errorf("ConnState should be ConnActive")
	}
	l.Close()

	time.Sleep(10 * time.Millisecond)
	if cc.State() != types.ConnClosed {
		t.Errorf("ConnState should be ConnClosed")
	}
}
