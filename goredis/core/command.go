package core

import "strconv"

func SetCommand(c *Client, s *Server) {
	objKey := c.Argv[1]
	objValue := c.Argv[2]
	if c.Argc != 3 {
		c.addReplyError("(error) ERR wrong number of arguments for 'set' command")
	}
	if stringKey, ok1 := objKey.Ptr.(string); ok1 {
		if stringValue, ok2 := objValue.Ptr.(string); ok2 {
			c.Db.Dict[stringKey] = CreateObject(ObjectTypeString, stringValue)
		}
	}
	s.Dirty++
	c.addReplyStatus("OK")
}

func GetCommand(c *Client, s *Server) {
	key := c.Argv[1].Ptr.(string)
	obj := c.lookupObject(key)
	if obj != nil {
		c.addReplyStatus(obj.Ptr.(string))
	} else {
		c.addReplyStatus("nil")
	}
}

func SubscribeCommand(c *Client, s *Server) {
	for j := 1; j < c.Argc; j++ {
		subscribeChannel(c, c.Argv[j], s)
	}
	c.Flags |= CLIENT_PUBSUB
}

func PublishCommand(c *Client, s *Server) {
	receivers := publishMessage(c.Argv[1], c.Argv[2], s)
	c.addReplyStatus(strconv.Itoa(receivers))
}
