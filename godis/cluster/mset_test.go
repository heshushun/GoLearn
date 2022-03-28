package cluster

import (
	"godis/redis/connection"
	"godis/redis/reply/asserts"
	"testing"
)

func TestMSet(t *testing.T) {
	conn := &connection.FakeConn{}
	allowFastTransaction = false
	ret := MSet(testCluster, conn, toArgs("MSET", "a", "a", "b", "b"))
	asserts.AssertNotError(t, ret)
	ret = testCluster.Exec(conn, toArgs("MGET", "a", "b"))
	asserts.AssertMultiBulkReply(t, ret, []string{"a", "b"})
}

func TestMSetNx(t *testing.T) {
	conn := &connection.FakeConn{}
	allowFastTransaction = false
	FlushAll(testCluster, conn, toArgs("FLUSHALL"))
	ret := MSetNX(testCluster, conn, toArgs("MSETNX", "a", "a", "b", "b"))
	asserts.AssertNotError(t, ret)
}
