package global

import (
	"os"
	"path/filepath"
	"sync"
)

func init() {
	Init()
}

var RootDir string

var once = new(sync.Once)

func Init() {
	once.Do(func() { // 只能运行一次
		inferRootDir()
		initConfig()
	})
}

func inferRootDir() {
	cwd, err := os.Getwd() // 获取当前工作目录
	if err != nil {
		panic(err)
	}
	var infer func(d string) string
	infer = func(d string) string {
		if exists(d + "/template") {
			return d
		}
		return infer(filepath.Dir(d))
	}
	RootDir = infer(cwd)
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
