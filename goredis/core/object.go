package core

const C_ERR = -1
const C_OK = 0

const ObjectTypeString = 0
const OBJ_LIST = 1
const OBJ_SET = 2
const OBJ_ZSET = 3
const OBJ_HASH = 4

const OBJ_ENCODING_RAW = 0        /* Raw representation */
const OBJ_ENCODING_INT = 1        /* Encoded as integer */
const OBJ_ENCODING_HT = 2         /* Encoded as hash table */
const OBJ_ENCODING_ZIPMAP = 3     /* Encoded as zipmap */
const OBJ_ENCODING_LINKEDLIST = 4 /* No longer used: old list encoding. */
const OBJ_ENCODING_ZIPLIST = 5    /* Encoded as ziplist */
const OBJ_ENCODING_INTSET = 6     /* Encoded as intset */
const OBJ_ENCODING_SKIPLIST = 7   /* Encoded as skiplist */
const OBJ_ENCODING_EMBSTR = 8     /* Embedded sds string encoding */
const OBJ_ENCODING_QUICKLIST = 9  /* Encoded as linked list of ziplists */

/*
*
* GoredisObject
*
**/
type GoredisObject struct {
	ObjectType int
	Ptr        interface{}
}

func CreateObject(t int, ptr interface{}) *GoredisObject {
	o := new(GoredisObject)
	o.ObjectType = t
	o.Ptr = ptr
	return o
}

type cmdFunc func(c *Client, s *Server)

type dict map[string]*GoredisObject

/*
*
* GoredisCommand
*
**/
type GoredisCommand struct {
	Name string
	Proc cmdFunc
}

/*
*
* GoredisDb
*
**/
type GoredisDb struct {
	Dict    dict
	Expires dict
	ID      int32
}
