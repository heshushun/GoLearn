package core

func subscribeChannel(c *Client, obj *GoredisObject, s *Server) {
	channel := obj.Ptr.(string)
	if clients, ok := s.PubSubChannels[channel]; ok {
		clients.listAddNodeTail(c)
	} else {
		clients = listCreate()
		clients.listAddNodeTail(c)
		s.PubSubChannels[channel] = clients
	}
}

func publishMessage(obj *GoredisObject, message *GoredisObject, s *Server) int {
	receivers := 0
	channel := obj.Ptr.(string)
	if clients, ok := s.PubSubChannels[channel]; ok {
		for list := clients.head; list != nil; list = list.next {
			c := list.value.(*Client)
			c.addReplyStatus(message.Ptr.(string))
			receivers++
		}
	}
	return receivers
}
