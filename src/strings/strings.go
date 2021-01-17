package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	s := "hello"
	fmt.Println("string: ", s)

	// 1、EqualFold 判断两个utf-8编码字符串 是否相同。
	f := strings.EqualFold(s, "Hello")
	fmt.Println("EqualFold: ", f)
	// EqualFold:  true

	// 2、HasPrefix 判断s是否有前缀字符串prefix。
	f = strings.HasPrefix(s, "he")
	fmt.Println("HasPrefix: ", f)
	// HasPrefix:  true

	// 3、HasSuffix 判断s是否有后缀字符串suffix。
	f = strings.HasSuffix(s, "llo")
	fmt.Println("HasSuffix: ", f)
	// HasSuffix:  true

	// 4、Contains 判断字符串s是否包含子串substr。
	f = strings.Contains(s, "ll")
	fmt.Println("Contains: ", s)
	// Contains:  hello

	// 5、ContainsAny 判断字符串s是否包含字符串chars中的任一字符。
	f = strings.ContainsAny(s, "h")
	f2 := strings.ContainsAny(s, "mp")
	fmt.Println("ContainsAny: ", f, f2)
	// ContainsAny:  true false

	// 6、Count 返回字符串s中有几个不重复的sep子串。
	c := strings.Count(s, "l")
	fmt.Println("Count: ", c)
	// Count:  2

	// 7、Index 子串sep在字符串s中第一次出现的位置，不存在则返回-1。
	i := strings.Index(s,"l")
	fmt.Println("Index: ", i)
	// Index:  2

	// 8、IndexAny 字符串chars中的任一utf-8码值在s中第一次出现的位置，如果不存在或者chars为空字符串则返回-1。
	i = strings.IndexAny(s, "hmp")
	fmt.Println("IndexAny: ", i)
	// IndexAny:  0

	// 9、LastIndex 子串sep在字符串s中最后一次出现的位置，不存在则返回-1。
	i = strings.LastIndex(s, "l")
	fmt.Println("LastIndex: ", i)
	// LastIndex:  3

	// 10、Title 返回s中每个单词的首字母都改为标题格式的字符串拷贝。
	s = strings.Title(s)
	fmt.Println("Title: ", s)
	// Title:  Hello

	// 11、ToLower 返回将所有字母都转为对应的小写版本的拷贝。
	s = "HELLO"
	s = strings.ToLower(s)
	fmt.Println("ToLower: ", s)
	// ToLower:  hello

	// 12、ToUpper 返回将所有字母都转为对应的大写版本的拷贝。
	s = "hello world"
	s = strings.ToUpper(s)
	fmt.Println("ToUpper: ", s)
	// ToUpper:  HELLO WORLD

	// 13、Replace 返回路径字符串中的卷名
	s = "hello"
	k := strings.Replace(s, "h", "k", -1)
	fmt.Println("Replace: ", k)
	// Replace:  kello

	// 14、Trim 根据pattern来判断name是否匹配，如果匹配则返回true
	s = "!!! Achtung! Achtung! !!!"
	sList := strings.Trim(s, "! ")
	fmt.Println("Trim: ", sList)
	// Trim:  Achtung! Achtung

	// 15、TrimSpace 返回将s前后端所有空白（unicode.IsSpace指定）都去掉的字符串。
	s = "hello  "
	s = strings.TrimSpace(s)
	fmt.Println("TrimSpace: ", s)
	// TrimSpace:  hello

	// 16、TrimPrefix 返回去除s可能的前缀prefix的字符串。
	s = "hello world"
	s = strings.TrimPrefix(s, "he")
	fmt.Println("TrimPrefix: ", s)
	// TrimPrefix:  llo world

	// 17、TrimSuffix 返回去除s可能的前缀prefix的字符串。
	s = "hello world"
	s = strings.TrimSuffix(s, "ld")
	fmt.Println("TrimSuffix: ", s)
	// TrimSuffix:  hello wor

	// 18、Fields 返回将字符串按照空白分割的多个字符串。
	s = "hello world"
	sList2 := strings.Fields(s)
	fmt.Println("Fields: ", sList2)
	// Fields:  [hello world]

	// 19、FieldsFunc 类似Fields，但使用函数f来确定分割符。
	s = "hello world"
	fu := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	sList2 = strings.FieldsFunc(s, fu)
	fmt.Println("FieldsFunc: ", sList2)
	// FieldsFunc:  [hello world]


	// 20、Split 用去掉s中出现的sep的方式进行分割，会分割到结尾，并返回生成的所有片段组成的切片
	s = "hello,world, you"
	sList2 = strings.Split(s, ",")
	fmt.Println("Split: ", sList2)
	// Split:  [hello world  you]

	// 21、SplitN 用去掉s中出现的sep的方式进行分割，会分割到结尾，并返回生成的所有片段组成的切片
	s = "hello;world;you"
	sList2 = strings.SplitN(s, ";", 2)
	fmt.Println("SplitN: ", sList2)
	// SplitN:  [hello world;you]

	// 22、SplitAfter 用从s中出现的sep后面切断的方式进行分割，会分割到结尾
	// 并返回生成的所有片段组成的切片
	s = "hello;world;you"
	sList2 = strings.SplitAfter(s, ";")
	fmt.Println("SplitAfter: ", sList2)
	// SplitAfter:  [hello; world; you]

	// 23、Join 将一系列字符串连接为一个字符串，之间用sep来分隔。
	sL := []string{"I", "love", "you"}
	s = strings.Join(sL, ".")
	fmt.Println("Join: ", s)
	// Join:  I.love.you

}
