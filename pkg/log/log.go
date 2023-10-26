package log

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is a wrapper around zerolog and options
type Logger struct {
	options *loggerOptions
	output  zerolog.Logger
}

// Modifies logger to use a custom directory for log file
func (l *Logger) WithDirectory(dir string) *Logger {
	l.options.directory = dir
	l.output = zerolog.New(newRotatingLogFile(dir, l.options.name)).With().Timestamp().Logger()
	return l
}

// Modifies logger to use a custom name for log file and logger
func (l *Logger) WithName(name string) *Logger {
	l.options.name = name
	l.output = zerolog.New(newRotatingLogFile(l.options.directory, name)).
		With().
		Timestamp().
		Logger()
	return l
}

// Modifies logger to also print to console (just fmt no json)
func (l *Logger) WithConsoleOutput(want bool) *Logger {
	l.options.wantConsoleOutput = want
	return l
}

// Modifies logger to use debug level
func (l *Logger) WithDebug(want bool) *Logger {
	if want {
		l.output.Level(zerolog.DebugLevel)
	}
	return l
}

// Prints a debug message to log output if level is debug
//
// May also print to os.Stdout if logger was created with WithConsoleOutput()
func (l *Logger) Debug(msg string) {
	if e := l.output.Debug(); e.Enabled() {
		e.Msg(msg)
		if l.options.wantConsoleOutput {
			fmt.Printf("[%s] %s\n", color.WhiteString(strings.ToUpper(l.options.name)), msg)
		}
	}
}

// Prints an info message to log output
//
// May also print to os.Stdout if logger was created with WithConsoleOutput()
func (l *Logger) Info(msg string) {
	l.output.Info().Msg(msg)
	if l.options.wantConsoleOutput {
		fmt.Printf("[%s] %s\n", color.BlueString(strings.ToUpper(l.options.name)), msg)
	}
}

// Prints a warning message to log output
//
// May also print to os.Stdout if logger was created with WithConsoleOutput()
func (l *Logger) Warn(msg string) {
	l.output.Warn().Msg(msg)
	if l.options.wantConsoleOutput {
		fmt.Printf("[%s] %s\n", color.YellowString(strings.ToUpper(l.options.name)), msg)
	}
}

// Prints an error message to log output
//
// May also print to os.Stderr if logger was created with WithConsoleOutput()
func (l *Logger) Error(msg string) {
	l.output.Error().Msg(msg)
	if l.options.wantConsoleOutput {
		_, err := fmt.Fprintf(
			os.Stderr,
			"[%s] %s\n",
			color.RedString(strings.ToUpper(l.options.name)),
			msg,
		)
		if err != nil {
			fmt.Println("log: error writing to stderr:", err)
		}
	}
}

// Creates a new logger with default settings:
//
// dir: "./logs"
//
// name: "app"
//
// wantConsoleOutput: false
//
// wantFileOutput: true
func NewLogger() *Logger {
	l := &Logger{
		options: &loggerOptions{
			directory:         defaultDirectory,
			name:              defaultName,
			wantConsoleOutput: defaultConsoleOutput,
			wantFileOutput:    defaultFileOutput,
		},
		output: zerolog.New(newRotatingLogFile(defaultDirectory, defaultName)).
			With().
			Timestamp().
			Logger(),
	}
	l.output.Level(zerolog.InfoLevel)
	return l
}

const (
	defaultDirectory     = "./logs"
	defaultName          = "app"
	defaultConsoleOutput = false
	defaultFileOutput    = true
)

type loggerOptions struct {
	directory         string
	name              string
	wantConsoleOutput bool
	wantFileOutput    bool
}

func newRotatingLogFile(dir string, name string) io.Writer {
	return &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s.log", dir, name),
		MaxSize:    25,
		MaxAge:     28,
		MaxBackups: 0,
		LocalTime:  true,
		Compress:   false,
	}
}
