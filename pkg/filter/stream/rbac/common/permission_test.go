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

package common

import "testing"

func TestPermissionAny(t *testing.T) {
	engine, _, err := genRoleBasedAccessControlEngine("./test_conf/deny-all.json")
	if err != nil {
		t.Error("TestPermissionAny failed")
		return
	}

	allowed, _ := engine.Allowed(nil, nil)
	if allowed {
		t.Error("TestPermissionAny failed")
		return
	}
}

func TestPermissionOrIds(t *testing.T) {
	engine, _, err := genRoleBasedAccessControlEngine("./test_conf/permission-or.json")
	if err != nil {
		t.Error("TestPermissionOrIds failed")
		return
	}

	allowed, _ := engine.Allowed(nil, nil)
	if allowed {
		t.Error("TestPermissionOrIds failed")
		return
	}

	engine, _, err = genRoleBasedAccessControlEngine("./test_conf/permission-and.json")
	if err != nil {
		t.Error("TestPermissionOrIds failed")
		return
	}

	allowed, _ = engine.Allowed(nil, nil)
	if !allowed {
		t.Error("TestPermissionOrIds failed")
		return
	}
}

func TestPermissionAndIds(t *testing.T) {
	engine, _, err := genRoleBasedAccessControlEngine("./test_conf/permission-and.json")
	if err != nil {
		t.Error("TestPermissionAndIds failed")
		return
	}

	allowed, _ := engine.Allowed(nil, nil)
	if !allowed {
		t.Error("TestPermissionAndIds failed")
		return
	}

	engine, _, err = genRoleBasedAccessControlEngine("./test_conf/permission-or.json")
	if err != nil {
		t.Error("TestPermissionAndIds failed")
		return
	}

	allowed, _ = engine.Allowed(nil, nil)
	if allowed {
		t.Error("TestPermissionAndIds failed")
		return
	}
}

func TestPermissionDestinationIp(t *testing.T) {
	engine, _, err := genRoleBasedAccessControlEngine("./test_conf/permission-dst-ip.json")
	if err != nil {
		t.Error("TestPermissionDestinationIp failed")
		return
	}

	cb := &mockStreamReceiverFilterCallbacks{
		conn: &mockConn{
			localAddr: &mockAddr{
				IP:   "1.2.3.4",
				Port: 8080,
			},
		},
	}
	allowed, _ := engine.Allowed(cb, nil)
	if !allowed {
		t.Error("TestPermissionDestinationIp failed")
		return
	}

	cb.conn.localAddr.IP = "1.2.3.100"
	allowed, _ = engine.Allowed(cb, nil)
	if !allowed {
		t.Error("TestPermissionDestinationIp failed")
		return
	}

	cb.conn.localAddr.IP = "1.2.4.1"
	allowed, _ = engine.Allowed(cb, nil)
	if allowed {
		t.Error("TestPermissionDestinationIp failed")
		return
	}
}

func TestPermissionDestinationPort(t *testing.T) {
	engine, _, err := genRoleBasedAccessControlEngine("./test_conf/permission-dst-port.json")
	if err != nil {
		t.Error("TestPermissionDestinationPort failed")
		return
	}

	cb := &mockStreamReceiverFilterCallbacks{
		conn: &mockConn{
			localAddr: &mockAddr{
				IP:   "1.2.3.4",
				Port: 8080,
			},
		},
	}
	allowed, _ := engine.Allowed(cb, nil)
	if !allowed {
		t.Error("TestPermissionDestinationPort failed")
		return
	}

	cb.conn.localAddr.Port = 8888
	allowed, _ = engine.Allowed(cb, nil)
	if allowed {
		t.Error("TestPermissionDestinationPort failed")
		return
	}
}

func TestPermissionHeader(t *testing.T) {
	// Present
	engine, _, err := genRoleBasedAccessControlEngine("./test_conf/permission-headers-present.json")
	if err != nil {
		t.Error("TestPermissionHeader failed")
		return
	}

	headers := &mockHeaderMap{
		headers: map[string]string{
			"X-Custom-Header": "123",
		},
	}

	allowed, _ := engine.Allowed(nil, headers)
	if allowed {
		t.Error("TestPermissionHeader failed")
		return
	}

	delete(headers.headers, "X-Custom-Header")

	allowed, _ = engine.Allowed(nil, headers)
	if !allowed {
		t.Error("TestPermissionHeader failed")
		return
	}

	// Header Value
	engine, _, err = genRoleBasedAccessControlEngine("./test_conf/permission-headers-value.json")
	if err != nil {
		t.Error("TestPermissionHeader failed")
		return
	}

	headers.headers["X-Mosn-Path"] = "/control-api/resources"
	allowed, _ = engine.Allowed(nil, headers)
	if allowed {
		t.Error("TestPermissionHeader failed")
		return
	}

	headers.headers["X-Mosn-Path"] = "/sources/test.java"
	allowed, _ = engine.Allowed(nil, headers)
	if allowed {
		t.Error("TestPermissionHeader failed")
		return
	}

	headers.headers["X-Mosn-Path"] = "/deny/custom"
	allowed, _ = engine.Allowed(nil, headers)
	if allowed {
		t.Error("TestPermissionHeader failed")
		return
	}

	delete(headers.headers, "X-Mosn-Path")
	headers.headers["X-Mosn-Method"] = "HEAD"
	allowed, _ = engine.Allowed(nil, headers)
	if allowed {
		t.Error("TestPermissionHeader failed")
		return
	}
}
