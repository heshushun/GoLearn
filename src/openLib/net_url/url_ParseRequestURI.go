package main

import (
    "fmt"
    "net/url"
)

func main() {
    values, err := url.ParseRequestURI("https://www.jd.com/?cu=true&utm_source=www.baidu.com&utm_medium=tuiguang&utm_campaign=t_1000003625_hao123mz&utm_term=3575ade7d11248bba9f9e6543c540777")
    fmt.Println(values)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("1 Scheme: ",values.Scheme)
    fmt.Println("2 Opaque: ",values.Opaque)
    fmt.Println("3 Host: ",values.Host)
    fmt.Println("4 Path: ",values.Path)
    fmt.Println("5 urlParam: ",values.RawQuery)
    fmt.Println("6 port: ",values.Port())

}

