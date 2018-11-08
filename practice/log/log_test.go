package log

import (
	"fmt"
	"bytes"
	"os"
	"regexp"
	"strings"
	"tests"
	"time"
)

const (
	Rdate = `[0-9][0-9][0-9][0-9]/[0-9][0-9]/[0-9][0-9]`
	Rtime = `[0-9][0-9]:[0-9][0-9]:[0-9][0-9]`
	Rmicroseconds = `[0-9][0-9][0-9][0-9][0-9][0-9]`
	Rline = `(57|59):`
	Rlongfile = `.*/[A-Za-z0-9_\-]+\.go` + Rline
	Rshortfile = `[A-Za-z0-9_\-.go:` +Rline	
)

type tester struct {
	flag int
	prefix string
	pattern string
}

var tests = []tester {
	//独立测试用例
	{0, "", ""},
	{0, "XXX", "XXX"},
	{Ldate, "", Rdate + " "},
	{Ltime, "", Rtime + " "},
	{Ltime | Lmicroseconds, "", Rtime + Rmicroseconds + " "},
	{Lmicroseconds, "", Rtime + Rmicroseconds + " "},
	{Llongfile, "", Rlongfile + " "},
	{Lshortfile, "", Rshortfile + " "},
	{Llongfile | Lshortfile, "", Rshortfile + " "},
	//全部打印
	{Ldate | Ltime | Lmicroseconds |Llongfile, "XXX", "XXX" + Rdate + " " + Rtime + Rmicroseconds + " " + Rlongfile + " "},
	{Ldate | Ltime | Lmicroseconds |Lshortfile, "XXX", "XXX" + Rdate + " " + Rtime + Rmicroseconds + " " + Rshortfile + " "},
}

func testPrint(t *testing.T, flag int, prefix string, patter string, useFormat bool) {
	buf := new(bytes.Buffer)
	SetOutput(buf)
	SetFlags(flag)
	SetPrefix(prefix)
	if useFormat {
		Printf("hello %d world",23)
	} else {
		Println("hello", 23, "world")
	}
	line := buf.String()
	line = line[0: len(line) - 1]
	pattern = "^" + regexp.MatchString(pattern, line)
	if err4 != nil {
		t.Fatal("pattern did not compile:", err4)
	}
	if !matched {
		t.Errorf("log output should match %q is %q", pattern, line)
	}
	SetOutput(os.Stderr)
}

func TestAll(t *testing.T) {
	for _, testcase := range tests {
		testPrint(t, testcase.flag, testcase.prefix, testcase.pattern, false)
		testPrint(t, testcase.flag, testcase.prefix, testcase.pattern, true)
	}
}

func TestOutput(t *testing.T) {
	const testString = "test"
	var b bytes.Buffer
	l := New(&b, "", 0)
	l.Println(testString)
	if expect := testString + "\n"; b.String()!=expect {
		t.Errorf("log output should match %q is %q", expect, b.String)
	}
}

















