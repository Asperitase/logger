package logs

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	Info = iota
	Warning
	Error
	Fatal
)

var tags = map[int]string{
	Info:    "[info] ",
	Warning: "[warning] ",
	Error:   "[error] ",
	Fatal:   "[fatal] ",
}

var (
	once     sync.Once
	instance *Logger
)

type color_filter_writer struct {
	Writer io.Writer
}

type Logger struct {
	Level   int
	IsColor bool
	Buffer  map[io.Writer]struct{}
	Mutex   sync.RWMutex
}

func New(level int, color bool, buffer io.Writer) *Logger {

	once.Do(func() {
		instance = &Logger{
			Level:   level,
			IsColor: color,
			Buffer:  map[io.Writer]struct{}{os.Stdout: {}},
		}
	})

	return instance
}

func (color_filter *color_filter_writer) Write(bytes []byte) (int, error) {
	remove_symbol := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean_bytes := remove_symbol.ReplaceAll(bytes, []byte(""))
	return color_filter.Writer.Write(clean_bytes)
}

func (this *Logger) AddWriter(writer io.Writer) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	filtered_writer := &color_filter_writer{Writer: writer}
	this.Buffer[filtered_writer] = struct{}{}
}

func (this *Logger) CancelWriter(writer io.Writer) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	delete(this.Buffer, writer)
}

func (this *Logger) SetLevel(level int) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.Level = level
}

func (this *Logger) IsColorAvailable(boolean bool) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.IsColor = boolean
}

func (this *Logger) get_tag(level int, color bool) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05.000000 ")

	var (
		white        = "\033[0m"
		green        = "\033[38;2;82;116;67m"
		light_green  = "\033[38;2;169;193;157m"
		yellow       = "\033[38;2;205;192;0m"
		light_yellow = "\033[38;2;254;242;120m"
		red          = "\033[38;2;188;19;26m"
		light_red    = "\033[38;2;246;126;115m"
	)

	if this.IsColor {
		switch level {
		case Info:
			return green + timestamp + light_green + tags[Info] + white
		case Warning:
			return yellow + timestamp + light_yellow + tags[Warning] + white
		case Error:
			return red + timestamp + light_red + tags[Error] + white
		case Fatal:
			return red + timestamp + light_red + tags[Fatal] + white
		}
	}

	return timestamp + tags[level]
}

func (this *Logger) format_arg(arg interface{}, spec string) string {
	switch v := arg.(type) {
	case int:
		if spec == "" {
			return fmt.Sprintf("%d", v)
		}
		return fmt.Sprintf("%"+spec+"d", v)
	case float64:
		if spec == "" {
			return fmt.Sprintf("%v", v)
		}
		return fmt.Sprintf("%"+spec, v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (this *Logger) format_message(format string, args ...interface{}) string {
	builder := strings.Builder{}
	builder.Grow(len(format) + 10*len(args))
	i := 0
	for i < len(format) {
		open_bracket := strings.IndexByte(format[i:], '{')
		if open_bracket == -1 {
			builder.WriteString(format[i:])
			break
		}
		builder.WriteString(format[i : i+open_bracket])

		close_bracket := strings.IndexByte(format[i+open_bracket:], '}')
		if close_bracket == -1 {
			builder.WriteString(format[i:])
			break
		}

		specifier := format[i+open_bracket+1 : i+open_bracket+close_bracket]
		var index int
		var format_spec string
		n, err := fmt.Sscanf(specifier, "%d:%s", &index, &format_spec)
		if err == nil && n == 2 {
			if index < len(args) {
				arg := args[index]
				builder.WriteString(this.format_arg(arg, format_spec))
			}
		} else if n, err := fmt.Sscanf(specifier, "%d", &index); err == nil && n == 1 {
			if index < len(args) {
				arg := args[index]
				builder.WriteString(this.format_arg(arg, ""))
			}
		}

		i += open_bracket + close_bracket + 1
	}

	return builder.String()
}

func (this *Logger) Log(level int, format string, args ...interface{}) {
	if level < 0 || level > Fatal {
		level = Info
	}

	if level < this.Level {
		return
	}

	message := this.format_message(format, args...)

	for writer := range this.Buffer {
		isColor := writer == os.Stdout
		writer.Write([]byte(this.get_tag(level, isColor) + message))

		if level == Fatal {
			os.Exit(0)
		}
	}
}
