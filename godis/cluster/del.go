package cluster

import (
	"godis/interface/redis"
	"godis/redis/reply"
	"strconv"
)

// Del atomically removes given writeKeys from cluster, writeKeys can be distributed on any node
// if the given writeKeys are distributed on different node, Del will use try-commit-catch to remove them
func Del(cluster *Cluster, c redis.Connection, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'del' command")
	}
	keys := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		keys[i-1] = string(args[i])
	}
	groupMap := cluster.groupBy(keys)
	if len(groupMap) == 1 && allowFastTransaction { // do fast
		for peer, group := range groupMap { // only one peerKeys
			return cluster.relay(peer, c, makeArgs("DEL", group...))
		}
	}
	// prepare
	var errReply redis.Reply
	txID := cluster.idGenerator.NextID()
	txIDStr := strconv.FormatInt(txID, 10)
	rollback := false
	for peer, peerKeys := range groupMap {
		peerArgs := []string{txIDStr, "DEL"}
		peerArgs = append(peerArgs, peerKeys...)
		var resp redis.Reply
		if peer == cluster.self {
			resp = execPrepare(cluster, c, makeArgs("Prepare", peerArgs...))
		} else {
			resp = cluster.relay(peer, c, makeArgs("Prepare", peerArgs...))
		}
		if reply.IsErrorReply(resp) {
			errReply = resp
			rollback = true
			break
		}
	}
	var respList []redis.Reply
	if rollback {
		// rollback
		requestRollback(cluster, c, txID, groupMap)
	} else {
		// commit
		respList, errReply = requestCommit(cluster, c, txID, groupMap)
		if errReply != nil {
			rollback = true
		}
	}
	if !rollback {
		var deleted int64 = 0
		for _, resp := range respList {
			intResp := resp.(*reply.IntReply)
			deleted += intResp.Code
		}
		return reply.MakeIntReply(int64(deleted))
	}
	return errReply
}