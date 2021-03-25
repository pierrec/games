// +build none

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	short = "short.txt"
	long  = "long.txt"
)

const commitLen = 12

func main() {
	var vshort, vlong string
	vlong = gitDescribe()
	if !strings.HasPrefix(vlong, "v") {
		vlong = "v0.0.0"
	}
	commit := gitLastCommit()
	utc, err := strconv.Atoi(gitLastCommitDate())
	check(err)
	t := time.Unix(int64(utc), 0)
	d := fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	vlong = fmt.Sprintf("%s-%s-%s", vlong, d, string(commit[:commitLen]))
	i := strings.IndexByte(vlong, '-')
	vshort = vlong[:i]
	check(ioutil.WriteFile(short, []byte(vshort), 0644))
	check(ioutil.WriteFile(long, []byte(vlong), 0644))
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func trim(b []byte, err error) string {
	check(err)
	return string(bytes.TrimSuffix(b, []byte{'\n'}))
}

func gitDescribe() string {
	git := exec.Command("git", "describe", "--always", fmt.Sprintf("--abbrev=%d", commitLen))
	return trim(git.Output())
}

func gitLastCommit() string {
	git := exec.Command("git", "rev-parse", "HEAD")
	return trim(git.Output())
}

func gitLastCommitDate() string {
	git := exec.Command("git", "log", "-1", "--date=unix", "--format=%cd", "HEAD")
	return trim(git.Output())
}
