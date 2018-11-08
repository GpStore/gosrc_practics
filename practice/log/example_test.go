package log_test //example_test文件隶属log_test package，注意这一点

import (
	"bytes"
	"fmt"
	"log"
)

func ExampleLogger() {
	var ( //这种写法很奇特，需要注意，都是初始化?
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	logger.Print("Hello, log file!")

	fmt.Print(&buf)
	// Output
	// logger: example_test.go:15: Hello, log file!
}

func ExampleLogger_Output() {
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "INFO: ", log.Lshortfile)

		infof = func(info string) {
			logger.Output(2, info)
		}
	)

	infof("Hello world!") //打印的是这一行
	fmt.Print(&buf)
	// Output:
	// INFO: example_test.go:32: Hello world!

}
