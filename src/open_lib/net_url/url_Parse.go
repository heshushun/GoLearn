package main

import (
	"fmt"
	"log"
	"net/url"
)

func main() {
    u, err := url.Parse("https://blog.csdn.net/wangshubo1989/article/details/75017632")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(u)
    fmt.Println("1 Scheme: ",u.Scheme)
    fmt.Println("2 Opaque: ",u.Opaque)
    fmt.Println("3 Host: ",u.Host)
    fmt.Println("4 Path: ",u.Path)
    fmt.Println("5 RawPath: ",u.RawPath)
    fmt.Println("6 ForceQuery: ",u.ForceQuery)
    fmt.Println("7 RawQuery: ",u.RawQuery)
    fmt.Println("8 Fragment: ",u.Fragment)
}

