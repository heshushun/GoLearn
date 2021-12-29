package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	path "path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/silenceper/log"
)

var (
	cmd          *exec.Cmd
	state        sync.Mutex
	eventTime    = make(map[string]int64)
	scheduleTime time.Time
)

func NewWatcher(paths []string, files []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf(" Fail to create new Watcher[ %s ]\n", err)
		os.Exit(2)
	}

	go func() {
		for {
			select {
			case e := <-watcher.Event:
				isBuild := true

				// Skip ignored files
				if shouldIgnoreFile(e.Name) {
					continue
				}
				// Skip not watch_ext
				if !checkIfWatchExt(e.Name) {
					continue
				}

				mt := getFileModTime(e.Name)
				if t := eventTime[e.Name]; mt == t {
					isBuild = false
				}

				eventTime[e.Name] = mt

				if isBuild {
					go func() {
						// Wait 1s before auto build util there is no file change.
						scheduleTime = time.Now().Add(1 * time.Second)
						for {
							time.Sleep(scheduleTime.Sub(time.Now()))
							if time.Now().After(scheduleTime) {
								break
							}
							return
						}
						AutoBuild(files)
					}()
				}
			case err := <-watcher.Error:
				log.Errorf("%v", err)
				log.Warnf(" %s\n", err.Error()) // No need to exit here
			}
		}
	}()

	log.Infof("Initializing watcher...\n")
	for _, watchPath := range paths {
		log.Infof("Directory( %s )\n", watchPath)
		err = watcher.Watch(watchPath)
		if err != nil {
			log.Errorf("Fail to watch directory[ %s ]\n", err)
			os.Exit(2)
		}
	}

}

// 不需要监听的文件
func shouldIgnoreFile(filename string) bool {
	for _, regex := range ignoredFilesRegExps {
		r, err := regexp.Compile(regex)
		if err != nil {
			panic("Could not compile the regex: " + regex)
		}
		if r.MatchString(filename) {
			return true
		}
		continue
	}
	return false
}

// 需要监听的格式
func checkIfWatchExt(name string) bool {
	for _, s := range cfg.WatchExts {
		if strings.HasSuffix(name, s) {
			return true
		}
	}
	return false
}

// 文件修改时间 防止重复编译
func getFileModTime(path string) int64 {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		log.Errorf("Fail to open file[ %s ]\n", err)
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Errorf("Fail to get file information[ %s ]\n", err)
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

func AutoBuild(files []string) {
	state.Lock()
	defer state.Unlock()

	log.Infof("Start building...\n")

	if err := os.Chdir(curpath); err != nil {
		log.Errorf("Chdir Error: %+v\n", err)
		return
	}

	cmdName := "go"

	var err error

	args := []string{"build"}
	args = append(args, "-o", cfg.Output)
	args = append(args, cfg.BuildArgs...)
	if cfg.BuildTags != "" {
		args = append(args, "-tags", cfg.BuildTags)
	}
	args = append(args, files...)

	bcmd := exec.Command(cmdName, args...)
	bcmd.Env = append(os.Environ(), "GOGC=off")
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	log.Infof("Build Args: %s %s", cmdName, strings.Join(args, " "))
	err = bcmd.Run()

	if err != nil {
		log.Errorf("============== Build failed ===================\n")
		log.Errorf("%+v\n", err)
		return
	}
	log.Infof("Build was successful\n")
	if !cfg.DisableRun {
		Restart(cfg.Output)
	}
}

// 重启
func Restart(appname string) {
	Kill()
	go Start(appname)
}

// 关闭 （杀掉进程）
func Kill() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Kill.recover -> ", e)
		}
	}()
	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			fmt.Println("Kill -> ", err)
		}
	}
}

// 启动
func Start(appname string) {
	log.Infof("Restarting %s ...\n", appname)
	if strings.Index(appname, "./") == -1 {
		appname = "./" + appname
	}

	cmd = exec.Command(appname)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = append([]string{appname}, cfg.CmdArgs...)
	cmd.Env = append(os.Environ(), cfg.Envs...)
	log.Infof("Run %s", strings.Join(cmd.Args, " "))
	go cmd.Run()

	log.Infof("%s is running...\n", appname)
	started <- true
}

// 遍历路径下的所有目录
func readAppDirectories(directory string, paths *[]string) {
	fileInfos, err := ioutil.ReadDir(directory)
	if err != nil {
		return
	}

	useDirectory := false
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), "docs") {
			continue
		}
		if strings.HasSuffix(fileInfo.Name(), "swagger") {
			continue
		}

		if isExcluded(path.Join(directory, fileInfo.Name())) {
			continue
		}

		if fileInfo.IsDir() == true && fileInfo.Name()[0] != '.' {
			readAppDirectories(directory+"/"+fileInfo.Name(), paths)
			continue
		}
		if useDirectory == true {
			continue
		}
		*paths = append(*paths, directory)
		useDirectory = true
	}
	return
}

// 过滤不需要监听的目录
func isExcluded(filePath string) bool {
	for _, p := range cfg.ExcludedPaths {
		absP, err := path.Abs(p)
		if err != nil {
			log.Errorf("err =%v", err)
			log.Errorf("Can not get absolute path of [ %s ]\n", p)
			continue
		}
		absFilePath, err := path.Abs(filePath)
		if err != nil {
			log.Errorf("Can not get absolute path of [ %s ]\n", filePath)
			break
		}
		if strings.HasPrefix(absFilePath, absP) {
			log.Infof("Excluding from watching [ %s ]\n", filePath)
			return true
		}
	}
	return false
}
