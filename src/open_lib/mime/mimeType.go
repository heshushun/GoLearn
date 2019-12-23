package main

import (
	"fmt"
	"mime"
	"path"
)

func main()  {
	filepath := "./1.png"
	mimetype := mime.TypeByExtension(path.Ext(filepath))
	fmt.Println(mimetype)

	filepath = "./2.txt"
	mimetype = mime.TypeByExtension(path.Ext(filepath))
	fmt.Println(mimetype)

}
