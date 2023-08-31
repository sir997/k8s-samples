package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var sh = `
buildctl --addr=tcp://10.23.0.134:80 \
build \
--frontend dockerfile.v0 \
--local context=./dist \
--local dockerfile=./dist/%d \
--output type=image,name=docker.ddmc-inc.com/fukang/appcenter:1.0.%d,push=true
`

var content = `
FROM docker.ddmc-inc.com/onepaas-artifact/deploy:0.0.6
COPY app%d /data/appcenter/
WORKDIR /data/appcenter
ENV APPBUILD_LANGUAGE=golang APPBUILD_LANGUAGE_VERSION=1.19
`

func main() {
	wg := sync.WaitGroup{}
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func(i int) {
			defer wg.Done()
			//			if err := os.Mkdir(fmt.Sprintf("dist/%d", i+1), os.ModeDir); err != nil {
			//				panic(err)
			//			}
			//			var filename = fmt.Sprintf("dist/%d/Dockerfile", i+1)
			//			var content = fmt.Sprintf(content, i+1)
			//			err := os.WriteFile(filename, []byte(content), os.ModePerm)
			//			if err != nil {
			//				panic(err)
			//			}
			cmd := exec.Command("sh", "-c", fmt.Sprintf(sh, i+1, i+1))
			output, err := cmd.CombinedOutput()
			if err != nil {
				panic(err)
			}
			if err := os.WriteFile(fmt.Sprintf("dist/%d/%d.txt", i+1, i+1), output, os.ModePerm); err != nil {
				panic(err)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("done")
}
