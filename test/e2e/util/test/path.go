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
	log.Printf("pwd: %s", pwd)
	if err != nil {
		panic(err)
	}
	rootDir := path.Clean(fmt.Sprintf("%s/../../..", pwd))

	if _, err := os.Stat(path.Join(rootDir, ".github")); err != nil {
		panic(fmt.Errorf("unexpected directory structure: %s should be the root dir (pwd: %s) but encountered error %w", rootDir, pwd, err))
	}
	var result = rootDir
	for _, segment := range strings.Split(pathFromRoot, "/") {
		result = path.Join(result, segment)
	}
	return result
}
