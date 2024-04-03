package flow

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/logrusorgru/aurora"
	"github.com/piaobeizu/flow/phase"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var Colorize = aurora.NewAurora(false)

// func initManager(ctx *phase.Context) error {
// 	c, ok := ctx.Context.Value(ctxConfigKey{}).(*v1beta1.Apollo)
// 	if c == nil || !ok {
// 		return fmt.Errorf("cluster config not available in context")
// 	}

// 	manager, err := phase.NewManager(c)
// 	if err != nil {
// 		return fmt.Errorf("failed to initialize phase manager: %w", err)
// 	}

// 	manager.Concurrency = ctx.Int("concurrency")
// 	manager.ConcurrentUploads = ctx.Int("concurrent-uploads")
// 	manager.DryRun = ctx.Bool("dry-run")

// 	ctx.Context = context.WithValue(ctx.Context, ctxManagerKey{}, manager)

// 	return nil
// }

// initLogging initializes the logger
func initLogging(debug bool) {
	log.SetLevel(log.TraceLevel)
	log.SetOutput(io.Discard)
	log.SetReportCaller(true)
	initScreenLogger(logLevelFromCtx(debug))
}

func logLevelFromCtx(debug bool) log.Level {
	if debug {
		return log.DebugLevel
	}
	return log.InfoLevel
}

func initScreenLogger(lvl log.Level) {
	log.AddHook(screenLoggerHook(lvl))
}

const logPath = "k0sctl/k0sctl.log"

func LogFile() (*os.File, error) {
	fn, err := xdg.SearchCacheFile(logPath)
	if err != nil {
		fn, err = xdg.CacheFile(logPath)
		if err != nil {
			return nil, err
		}
	}

	logFile, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("Failed to open log %s: %s", fn, err.Error())
	}

	_, _ = fmt.Fprintf(logFile, "time=\"%s\" level=info msg=\"###### New session ######\"\n", time.Now().Format(time.RFC822))

	return logFile, nil
}

type loghook struct {
	Writer    io.Writer
	Formatter log.Formatter

	levels []log.Level
}

func (h *loghook) SetLevel(level log.Level) {
	h.levels = []log.Level{}
	for _, l := range log.AllLevels {
		if level >= l {
			h.levels = append(h.levels, l)
		}
	}
}

func (h *loghook) Levels() []log.Level {
	return h.levels
}

func (h *loghook) Fire(entry *log.Entry) error {
	line, err := h.Formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to format log entry: %v", err)
		return err
	}
	_, err = h.Writer.Write(line)
	return err
}

func screenLoggerHook(lvl log.Level) *loghook {
	var forceColors bool
	var writer io.Writer
	if runtime.GOOS == "windows" {
		writer = ansicolor.NewAnsiColorWriter(os.Stdout)
		forceColors = true
	} else {
		writer = os.Stdout
		if fi, _ := os.Stdout.Stat(); (fi.Mode() & os.ModeCharDevice) != 0 {
			forceColors = true
		}
	}

	if forceColors {
		Colorize = aurora.NewAurora(true)
		phase.Colorize = Colorize
	}

	l := &loghook{
		Writer: writer,
		// Formatter: &log.TextFormatter{DisableTimestamp: lvl < log.DebugLevel, ForceColors: forceColors},
		Formatter: &LogFormatter{},
	}

	l.SetLevel(lvl)

	return l
}

func fileLoggerHook(logFile io.Writer) *loghook {
	l := &loghook{
		Formatter: &log.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        time.RFC822,
			DisableLevelTruncation: true,
		},
		Writer: logFile,
	}

	l.SetLevel(log.DebugLevel)

	return l
}

type LogFormatter struct{}

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

// 实现Formatter(entry *logrus.Entry) ([]byte, error)接口
func (t *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	//根据不同的level去展示颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//自定义日期格式
	// timestamp := entry.Time.Format("2006-01-02 15:04:05")
	if entry.HasCaller() {
		//自定义文件路径
		// funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		fileVal = fmt.Sprintf("%-15s", fileVal)
		level := strings.ToUpper(entry.Level.String()[:4])
		//自定义输出格式
		fmt.Fprintf(b, "\x1b[%dm[%s] %s ===> %s\x1b[0m\n", levelColor, level, fileVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] \x1b[%dm\x1b[0m ===> %s\n", entry.Level, levelColor, entry.Message)
	}
	return b.Bytes(), nil
}
