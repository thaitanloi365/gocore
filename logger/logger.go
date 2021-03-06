package logger

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/thaitanloi365/gocore/logger/notifier"
)

// Logger instance
type Logger struct {
	context    context.Context
	cancelFunc context.CancelFunc
	config     *Config

	mutex sync.RWMutex

	queue chan *logTask

	writer     Writer
	fileWriter Writer

	debugStr      string
	debugColorStr string

	infoStr      string
	infoColorStr string

	warnStr      string
	warnColorStr string

	errStr      string
	errColorStr string

	notifier *notifier.SlackNotifier
}

// Config log config
type Config struct {
	BufferedSize int
	Colorful     bool
	TimeLocation *time.Location
	DateFormat   string
	Prefix       string

	Writer                Writer
	WriteFileExceptLevels []LogLevel

	Notifier *notifier.SlackNotifier
}

// New new writter
func New(config *Config) *Logger {
	var bufferedSize = 10
	var dateFormat = "2006-01-02 15:04:05 Z07:00"
	timeLocation, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	var defaultConfig = config
	if defaultConfig == nil {
		defaultConfig = &Config{
			Prefix:       "",
			BufferedSize: bufferedSize,
			DateFormat:   dateFormat,
			TimeLocation: timeLocation,
			Colorful:     true,
			Notifier:     nil,
		}
	}

	if defaultConfig.BufferedSize == 0 {
		defaultConfig.BufferedSize = bufferedSize
	}

	if defaultConfig.DateFormat == "" {
		defaultConfig.DateFormat = dateFormat
	}

	if defaultConfig.TimeLocation == nil {
		defaultConfig.TimeLocation = timeLocation
	}

	var writer = log.New(os.Stdout, "\r\n", 0)
	var fileWriter Writer = log.New(ioutil.Discard, "", 0)

	if defaultConfig.Writer != nil {
		fileWriter = defaultConfig.Writer
	}

	var (
		debugStr      = "%s DEBUG %s "
		infoStr       = "%s INFO %s "
		warnStr       = "%s WARN %s "
		errStr        = "%s ERROR %s "
		debugColorStr = "%s " + Green("DEBUG %s\n")
		infoColorStr  = "%s " + Blue("INFO %s\n")
		warnColorStr  = "%s " + Yellow("WARN %s\n")
		errColorStr   = "%s " + Red("ERROR %s\n")
	)

	ctx, cancelFunc := context.WithCancel(context.Background())
	var logger = &Logger{
		config:        defaultConfig,
		writer:        writer,
		fileWriter:    fileWriter,
		mutex:         sync.RWMutex{},
		context:       ctx,
		cancelFunc:    cancelFunc,
		queue:         make(chan *logTask, defaultConfig.BufferedSize),
		debugStr:      debugStr,
		debugColorStr: debugColorStr,
		infoStr:       infoStr,
		infoColorStr:  infoColorStr,
		warnStr:       warnStr,
		warnColorStr:  warnColorStr,
		errStr:        errStr,
		errColorStr:   errColorStr,
		notifier:      defaultConfig.Notifier,
	}

	logger.run()

	return logger
}

// Printf debug
func (l *Logger) Printf(format string, values ...interface{}) {
	l.queue <- l.buildlog(Debug, "", valueTypeCustom, format, values...)
}

// Debug debug
func (l *Logger) Debug(values ...interface{}) {
	l.queue <- l.buildlog(Debug, l.fileWithLineNum(), valueTypeInterface, "", values...)
}

// DebugWithEchoContext wrap http
func (l *Logger) DebugWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Debug, l.fileWithLineNum(), valueTypeInterface, "", values...).withEchoContext(c)
}

// Debugf debug with format
func (l *Logger) Debugf(format string, values ...interface{}) {
	l.queue <- l.buildlog(Debug, l.fileWithLineNum(), valueTypeInterface, format, values...)
}

// DebugfWithEchoContext debug with format
func (l *Logger) DebugfWithEchoContext(c echo.Context, format string, values ...interface{}) {
	l.queue <- l.buildlog(Debug, l.fileWithLineNum(), valueTypeInterface, format, values...).withEchoContext(c)
}

// DebugJSON print pretty json
func (l *Logger) DebugJSON(values ...interface{}) {
	l.queue <- l.buildlog(Debug, l.fileWithLineNum(), valueTypeJSON, "", values...)
}

// DebugJSONWithEchoContext print pretty json
func (l *Logger) DebugJSONWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Debug, l.fileWithLineNum(), valueTypeJSON, "", values...).withEchoContext(c)
}

// Info info
func (l *Logger) Info(values ...interface{}) {
	l.queue <- l.buildlog(Info, l.fileWithLineNum(), valueTypeInterface, "", values...)
}

// InfofWithEchoContext info with format
func (l *Logger) InfofWithEchoContext(c echo.Context, format string, values ...interface{}) {
	l.queue <- l.buildlog(Info, l.fileWithLineNum(), valueTypeInterface, format, values...).withEchoContext(c)
}

// Infof info with format
func (l *Logger) Infof(format string, values ...interface{}) {
	l.queue <- l.buildlog(Info, l.fileWithLineNum(), valueTypeInterface, format, values...)
}

// InfoWithEchoContext info with format
func (l *Logger) InfoWithEchoContext(c echo.Context, format string, values ...interface{}) {
	l.queue <- l.buildlog(Info, l.fileWithLineNum(), valueTypeInterface, format, values...).withEchoContext(c)
}

// InfoJSON print pretty json
func (l *Logger) InfoJSON(values ...interface{}) {
	l.queue <- l.buildlog(Info, l.fileWithLineNum(), valueTypeJSON, "", values...)
}

// InfoJSONWithEchoContext print pretty json
func (l *Logger) InfoJSONWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Info, l.fileWithLineNum(), valueTypeJSON, "", values...).withEchoContext(c)
}

// Warn warn
func (l *Logger) Warn(values ...interface{}) {
	l.queue <- l.buildlog(Warn, l.fileWithLineNum(), valueTypeInterface, "", values...)
}

// Warn warn
func (l *Logger) WarnWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Warn, l.fileWithLineNum(), valueTypeInterface, "", values...).withEchoContext(c)
}

// Warnf info with format
func (l *Logger) Warnf(format string, values ...interface{}) {
	l.queue <- l.buildlog(Warn, l.fileWithLineNum(), valueTypeInterface, format, values...)
}

// WarnfWithEchoContext info with format
func (l *Logger) WarnfWithEchoContext(c echo.Context, format string, values ...interface{}) {
	l.queue <- l.buildlog(Warn, l.fileWithLineNum(), valueTypeInterface, format, values...).withEchoContext(c)
}

// WarnJSON print pretty json
func (l *Logger) WarnJSON(values ...interface{}) {
	l.queue <- l.buildlog(Warn, l.fileWithLineNum(), valueTypeJSON, "", values...)
}

// WarnJSONWithEchoContext print pretty json
func (l *Logger) WarnJSONWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Warn, l.fileWithLineNum(), valueTypeJSON, "", values...).withEchoContext(c)
}

// Error error
func (l *Logger) Error(values ...interface{}) {
	l.queue <- l.buildlog(Error, l.fileWithLineNum(), valueTypeInterface, "", values...)
}

// ErrorWithEchoContext error
func (l *Logger) ErrorWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Error, l.fileWithLineNum(), valueTypeInterface, "", values...).withEchoContext(c)
}

// Errorf error with format
func (l *Logger) Errorf(format string, values ...interface{}) {
	l.queue <- l.buildlog(Error, l.fileWithLineNum(), valueTypeInterface, format, values...)
}

// ErrorfWithEchoContext error with format
func (l *Logger) ErrorfWithEchoContext(c echo.Context, format string, values ...interface{}) {
	l.queue <- l.buildlog(Error, l.fileWithLineNum(), valueTypeInterface, format, values...).withEchoContext(c)
}

// ErrorJSON print pretty json
func (l *Logger) ErrorJSON(values ...interface{}) {
	l.queue <- l.buildlog(Error, l.fileWithLineNum(), valueTypeJSON, "", values...)
}

// ErrorJSONWithEchoContext print pretty json
func (l *Logger) ErrorJSONWithEchoContext(c echo.Context, values ...interface{}) {
	l.queue <- l.buildlog(Error, l.fileWithLineNum(), valueTypeJSON, "", values...).withEchoContext(c)
}

func (l *Logger) run() {
	go l.cleanup()

	go func(ctx context.Context, queue chan *logTask) {
		for {
			select {
			case <-ctx.Done():
				return

			case data := <-queue:
				var format = l.infoStr
				var formatColor = l.infoColorStr
				var extraFormat = data.format
				var extraPrettyFormat = data.format
				switch data.logLevel {
				case Debug:
					format = l.debugStr
					formatColor = l.debugColorStr
				case Error:
					format = l.errStr
					formatColor = l.errColorStr
				case Warn:
					format = l.warnStr
					formatColor = l.warnColorStr
				}

				var separator = " "
				switch data.valueType {
				case valueTypeJSON:
					separator = "\n"
				}

				if extraPrettyFormat == "" {
					for i := 0; i < len(data.values); i++ {
						extraPrettyFormat = "%v" + separator + extraPrettyFormat
					}
				}
				if extraFormat == "" {
					for i := 0; i < len(data.values); i++ {
						extraFormat = "%v" + " " + extraFormat
					}
				}

				var fullFormatColor = formatColor + extraPrettyFormat
				var fullFormat = format + extraFormat

				if data.requestInfo != nil {
					fullFormatColor = data.formatRequestInfo() + "\n" + fullFormatColor
					fullFormat = data.formatRequestInfo() + " " + fullFormat
				}

				switch data.valueType {
				case valueTypeCustom:
					l.writer.Printf(data.format, data.values...)
					if l.ignoreWriteFile(data.logLevel) == false {

						l.writer.Printf(data.format, data.values...)

						if l.notifier != nil {
							var titleFormat = format
							if data.requestInfo != nil {
								titleFormat = data.formatRequestInfo() + "\n" + titleFormat
							}

							l.notifier.Send(fmt.Sprintf(titleFormat, data.time), fmt.Sprintf(data.format, data.values...))
						}
					}

				case valueTypeJSON:
					var prettyValues = []interface{}{}
					var values = []interface{}{}
					for _, value := range data.values {
						values = append(values, ToJSONString(value))
						prettyValues = append(prettyValues, ToPrettyJSONString(value))
					}
					l.writer.Printf(fullFormatColor, append([]interface{}{data.time, data.caller}, prettyValues...)...)
					if l.ignoreWriteFile(data.logLevel) == false {

						l.fileWriter.Printf(fullFormat, append([]interface{}{data.time, data.caller}, values...)...)

						if l.notifier != nil {
							var titleFormat = format
							if data.requestInfo != nil {
								titleFormat = data.formatRequestInfo() + "\n" + titleFormat
							}

							l.notifier.Send(fmt.Sprintf(titleFormat, data.time, data.caller), fmt.Sprintf(extraFormat, prettyValues...))
						}
					}
				default:
					l.writer.Printf(fullFormatColor, append([]interface{}{data.time, data.caller}, data.values...)...)
					if l.ignoreWriteFile(data.logLevel) == false {
						l.fileWriter.Printf(fullFormat, append([]interface{}{data.time, data.caller}, data.values...)...)
						if l.notifier != nil {
							var titleFormat = format
							if data.requestInfo != nil {
								titleFormat = data.formatRequestInfo() + "\n" + titleFormat
							}

							l.notifier.Send(fmt.Sprintf(titleFormat, data.time, data.caller), fmt.Sprintf(extraFormat, data.values...))
						}
					}

				}

				break
			}
		}
	}(l.context, l.queue)
}

func (l *Logger) cleanup() {
	<-l.context.Done()

	// Lock the destinations
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Cleanup the destinations
	close(l.queue)

}

func (l *Logger) buildlog(logtype LogLevel, caller string, valueType valueType, format string, values ...interface{}) (newlog *logTask) {
	newlog = &logTask{
		logger:    l,
		logLevel:  logtype,
		time:      time.Now().Format(l.config.DateFormat),
		format:    format,
		values:    values,
		caller:    caller,
		valueType: valueType,
	}

	return newlog
}

func (l *Logger) fileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		if ok {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ""
}

func (l *Logger) ignoreWriteFile(level LogLevel) bool {
	for _, lv := range l.config.WriteFileExceptLevels {
		if lv == level {
			return true
		}
	}

	return false

}
