package main
import "C"

// #include <stdio.h>
// #include <stdlib.h>
/*
void print(char *str) {
    printf("%s\n", str);
}
*/

import "unsafe"

func main() {
    s := "Hello Cgo"
    cs := C.CString(s)
    C.print(cs)
    C.free(unsafe.Pointer(cs))
}