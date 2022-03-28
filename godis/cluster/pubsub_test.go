package cluster

import (
	"godis/lib/utils"
	"godis/redis/connection"
	"godis/redis/parser"
	"godis/redis/reply/asserts"
	"testing"
)

func TestPublish(t *testing.T) {
	channel := utils.RandString(5)
	msg := utils.RandString(5)
	conn := &connection.FakeConn{}
	Subscribe(testCluster, conn, utils.ToCmdLine("SUBSCRIBE", channel))
	conn.Clean() // clean subscribe success
	Publish(testCluster, conn, utils.ToCmdLine("PUBLISH", channel, msg))
	data := conn.Bytes()
	ret, err := parser.ParseOne(data)
	if err != nil {
		t.Error(err)
		return
	}
	asserts.AssertMultiBulkReply(t, ret, []string{
		"message",
		channel,
		msg,
	})

	// unsubscribe
	UnSubscribe(testCluster, conn, utils.ToCmdLine("UNSUBSCRIBE", channel))
	conn.Clean()
	Publish(testCluster, conn, utils.ToCmdLine("PUBLISH", channel, msg))
	data = conn.Bytes()
	if len(data) > 0 {
		t.Error("expect no msg")
	}

	// unsubscribe all
	Subscribe(testCluster, conn, utils.ToCmdLine("SUBSCRIBE", channel))
	UnSubscribe(testCluster, conn, utils.ToCmdLine("UNSUBSCRIBE"))
	conn.Clean()
	Publish(testCluster, conn, utils.ToCmdLine("PUBLISH", channel, msg))
	data = conn.Bytes()
	if len(data) > 0 {
		t.Error("expect no msg")
	}
}
