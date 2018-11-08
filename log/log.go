package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LUTC //时区
	LstdFlags = Ldate | Ltime
)

type Logger struct {
	mu sync.Mutex	//用于临时调整输出终端用，默认是输出到stdErr
	prefix string	//前缀标记，将来可以作为扩展用
	flag int //标记位，涉及日期时间时区，以及文件名或全路径名称
	out io.Writer //输出终端
	buf []byte	//输出内容
}

func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{out:out, prefix:prefix, flag:flag}
}



/*------------------------------------------------*/

// 修改放在外部，通过接口调整输出目标，开闭原则，通过新增代码来实现实现变动
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out // 一个Logger对象在生命周期内任意特定时刻，只有一个输出终端
}

//默认情况下都是输出到Stderr，输出格式日期和时间（时分秒），没有要求毫秒


// 给定指定整数和输出字符串宽度和目标Buf,将类似2015的整数，转化为字符串"2015".
// 注意小写代表内部使用, 另外当wid < 0 时，只要数字提取完毕就退出循环，前面不补0
func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid >1 {
		wid-- //控制循环关键3
		q := i/10 //整除取结果
		b[bp] = byte('0' + i - q*10) //在低位存余数,wid更多时补0
		bp-- //控制循环关键1
		i = q //控制循环关键2
	}
	// i < 10
	b[bp] = byte('0' + i) //存入最后一个余数，这里是最重要细节
	*buf = append(*buf, b[bp:]...) //注意格式,因为添加的是字符串，所以需要...
}

//这个函数的提炼是非常关键的，最核心的参数齐备以后，基于这个函数，方便外面调用就好写很多
func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	*buf = append(*buf, l.prefix...) //如果设置有前缀
	if l.flag&(Ldate | Ltime | Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 { //输出形如 “2015/06/13 " 
			year,month,day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, month, 2)
			*buf = append(*buf, '/')
			itoa(buf, day,2)
			*buf =append(*buf, ' ')
		}
	} 
	if l.flag&(Llongfile | Lshortfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i:= len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1) // 前面不要补0
		*buf = append(*buf, ": "...) //注意有分界符号 空格；
	}
}

// 二级提炼封装,调用内部formatHeader,然后核心在runtime.Caller(depth)调用，s作为补充
func (l *Logger) Output(calldepth int, s string) error {
	//早点获取时间点信息
	var now time.Time
	if l.flag&(Ldate | Ltime | Lmicroseconds) != 0 {
		now = time.Now()
	}

	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.flag&(Lshortfile| Llongfile) {
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf,now, file, line)
	l.buf = append(l.buf,s...)
	if len(s) == 0 || s[len(s) - 1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_,err := l.our.Write(l.buf)
	return err
}


// 在外层最常用的调用模式，主要是考虑入参合出参的界定，需要工程经验，业务场景很熟悉

func (l *Logger) Printf(format string, v...interface{}){
	l.Output(2, fmt.Sprintf(format, v ...))
}

func (l *Logger) Println(v ...interface{}) {
	l.Output(2,fmt.Sprintln(v...))
}

func (l *Logger) Fatal(v ...interface{}){
	l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string,v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(2, s)
	panic(s)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(2, s)
	panic(s)
}

func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(2, s)
	panic(s)
}

//注意命名，没有Get，然后是加了复数s，表示组合
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

/*------------------------------------------------*/
//特例化的接口，指定了输出目标, os.Stderr是标准出错输出的地方
var std = New(os.Stderr, "", LstdFlags)

func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

func Flags() int {
	return std.Flags()
}
 
func SetFlags(flag int) {
	std.SetFlags(flag)
}

func Prefix() int {
	return std.Prefix()
}
 
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}

/*------------------------------------------------*/
func Print(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...))
}

func Println(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}
 
func Printf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}

func Fatal(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}


func Fatalln(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func Fatalf(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}


func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(2, s)
	panic(s)
}


func Panicln(format string, v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(2, s)
	panic(s)
}

func Panicf(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(2, s)
	panic(s)
}

func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s)
}
 



















