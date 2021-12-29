package main

import (
	"flag"
	"os"
	path "path/filepath"
	"runtime"
	"strings"
)

var (
	cfg     *config
	curpath string
	exit    chan bool

	output   string
	buildPkg string
	cmdArgs  string

	started chan bool
)

func init() {
	flag.StringVar(&output, "o", "", "go build output")
	flag.StringVar(&buildPkg, "p", "", "go build packages")
	flag.StringVar(&cmdArgs, "args", "", "app run args,separated by commas. like: -args='-host=:8080,-name=demo'")
}

var ignoredFilesRegExps = []string{
	`.#(\w+).go`,
	`.(\w+).go.swp`,
	`(\w+).go~`,
	`(\w+).tmp`,
}

func main() {
	flag.Parse() // 解析参数

	cfg = parseConfig()
	curpath, _ = os.Getwd() // 当前目录

	// AppName
	if cfg.AppName == "" {
		if output == "" {
			cfg.AppName = path.Base(curpath)
		} else {
			cfg.AppName = path.Base(output)
		}
	}

	// Output
	outputExt := ""
	if runtime.GOOS == "windows" {
		outputExt = ".exe"
	}
	if output != "" {
		cfg.Output = output + "/" + cfg.AppName + outputExt
	} else {
		cfg.Output = "./" + cfg.AppName + outputExt
	}

	// CmdArgs
	if cmdArgs != "" {
		cfg.CmdArgs = strings.Split(cmdArgs, ",")
	}

	// WatchExts 监听的文件后缀
	cfg.WatchExts = append(cfg.WatchExts, ".go")

	runApp()
}

func runApp() {

	// WatchPaths
	var paths []string
	if len(cfg.WatchPaths) != 0 {
		for _, watchPath := range cfg.WatchPaths {
			readAppDirectories(watchPath, &paths)
		}
	} else {
		readAppDirectories(curpath, &paths)
	}

	// BuildPkg 需要编译的文件
	files := []string{}
	if buildPkg == "" {
		buildPkg = cfg.BuildPkg
	}
	if buildPkg != "" {
		files = strings.Split(buildPkg, ",")
	}

	// TODO: 编译是当前工作目录文件；监听是指定目录（默认当前工作目录）
	NewWatcher(paths, files)

	AutoBuild(files)

	for {
		select {
		case <-exit:
			runtime.Goexit()
		}
	}
}
