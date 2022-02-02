package core

const ObjectTypeString = 1

/*
*
* GoredisObject
*
**/
type GoredisObject struct {
	ObjectType int
	Ptr        interface{}
}

func NewObject(t int, ptr interface{}) *GoredisObject {
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
