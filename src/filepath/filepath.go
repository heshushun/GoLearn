package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	s := "F:\\golang_workspace\\leetcode\\aa.js"
	fmt.Println("path: ", s)

	// 1、ToSlash 将 path 中平台相关的路径分隔符转换为 '/'
	s = filepath.ToSlash(s)
	fmt.Println("ToSlash: ", s)
	// ToSlash:  F:/golang_workspace/leetcode

	// 2、FromSlash 将 path 中的 '/' 转换为系统相关的路径分隔符
	s = filepath.FromSlash(s)
	fmt.Println("FromSlash: ", s)
	// FromSlash:  F:\golang_workspace\leetcode

	// 3、Dir 获取 path 中最后一个分隔符之前的部分（不包含分隔符）
	s = "/golang_workspace/leetcode/aa.js"
	s = filepath.Dir(s)
	fmt.Println("Dir: ", s)
	// Dir:  \golang_workspace\leetcode

	// 4、Base 获取path中最后一个分隔符之后的部分(不包含分隔符)
	s = "/golang_workspace/leetcode/aa.js"
	s = filepath.Base(s)
	fmt.Println("Base: ", s)
	// Base:  aa.js

	// 5、Base 获取 path 中最后一个分隔符前后的两部分
	// 之前包含分隔符，之后不包含分隔符
	s = "/golang_workspace/leetcode/aa.js"
	d, s := filepath.Split(s)
	fmt.Println("Split: ", d, s)
	// Split:  /golang_workspace/leetcode/ aa.js

	// 6、Base 获取路径字符串中的文件扩展名
	s = "/golang_workspace/leetcode/aa.js"
	s = filepath.Ext(s)
	fmt.Println("Ext: ", s)
	// Ext:  .js

	// 7、Rel 获取 targpath 相对于 basepath 的路径。
	s = "/golang_workspace/leetcode/aa.js"
	s2 := "/golang_workspace/"
	s, _ = filepath.Rel(s2,s)
	fmt.Println("Rel: ", s)
	// Rel:  leetcode\aa.js

	// 8、Join 将 elem 中的多个元素合并为一个路径，忽略空元素，清理多余字符。
	s = "golang_workspace"
	s2 = "leetcode/aa.js"
	s = filepath.Join(s,s2)
	fmt.Println("Join: ", s)
	// Join:  golang_workspace\leetcode\aa.js

	// 9、Clean 清理路径中的多余字符，比如 /// 或 ../ 或 ./
	//返回等价的最短路径
	//1.用一个斜线替换多个斜线
	s = filepath.Clean("/.../..../////abc/abc")
	//2.清除当前路径.
	s = filepath.Clean("./1.txt")
	//3.清除内部的..和他前面的元素
	s = filepath.Clean("C:/a/b/../c")
	//4.以/..开头的，变成/
	s = filepath.Clean("/../1.txt")
	fmt.Println("Clean: ", s)
	// Clean:  \1.txt

	// 10、判断路径是否为绝对路径
	s = "/home/gopher"
	s2 = ".bashrc"
	f := filepath.IsAbs(s)
	f = filepath.IsAbs(s2)
	fmt.Println("IsAbs: ", f)
	// IsAbs:  true  IsAbs:  false

	// 11、返回所给目录的绝对路径
	s = ".bashrc"
	s,_ = filepath.Abs(s)
	fmt.Println("Abs: ", s)
	// Abs:  F:\golang_workspace\GoLearn\.bashrc

	// 12、将路径序列 操作系统特别的连接符组成的path
	s = "/a/b/c:/usr/bin"
	sList := filepath.SplitList(s)
	fmt.Println("SplitList: ", sList)
	// SplitList:  [/a/b/c:/usr/bin]

	// 13、返回路径字符串中的卷名
	s = "F:\\golang_workspace\\leetcode\\aa.js"
	s = filepath.VolumeName(s)
	fmt.Println("VolumeName: ", s)
	// VolumeName:  F:

	// 14、根据pattern来判断name是否匹配，如果匹配则返回true
	r := "/home/catch/*"
	s = "/home/catch/foo"
	f, _ = filepath.Match(r, s)
	fmt.Println("Match: ", f)
	// Match:  true

	// 15、列出与指定的模式 pattern 完全匹配的文件或目录（匹配原则同match）
	r = "F:\\golang_workspace\\[s]*"
	sList,_ = filepath.Glob(r)
	fmt.Println("Glob: ", sList)
	// Glob:  [F:\golang_workspace\shenqi_server F:\golang_workspace\src]

	// 16、遍历指定目录(包括子目录)，对遍历的项目用walkFn函数进行处理
	pwd,_ := os.Getwd()
	filepath.Walk(pwd,func(fpath string, info os.FileInfo, err error) error {
		if match,err := filepath.Match("???",filepath.Base(fpath)); match {
			fmt.Println("Walk path:",fpath)
			fmt.Println("Walk info:",info)
			return err
		}
		return nil
	})
	// Walk path:  F:\golang_workspace\GoLearn\src

}
