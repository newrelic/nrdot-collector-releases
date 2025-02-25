package test

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func NewPathRelativeToRootDir(pathFromRoot string) string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	var dir = pwd
	for {
		_, err = os.Stat(path.Join(dir, ".github"))
		if err == nil {
			break
		}
		if dir == "/" {
			log.Panicf("couldn't find repo root dir starting from pwd: %s", pwd)
		}
		dir = path.Clean(fmt.Sprintf("%s/..", dir))
	}
	var result = dir
	for _, segment := range strings.Split(pathFromRoot, "/") {
		result = path.Join(result, segment)
	}
	return result
}
