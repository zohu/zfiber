package zlog

import (
	"context"
	"encoding"
	"fmt"
	"github.com/zohu/zfiber/zutil"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unicode"
)

const (
	ansiReset  = "\033[0m"
	ansiDebug  = "\033[90m"
	ansiInfo   = "\033[32m"
	ansiWarn   = "\033[33m"
	ansiError  = "\033[31m"
	ansiTime   = "\033[37m"
	ansiSource = "\033[34;4m"
)

var (
	defaultLevel      = slog.LevelInfo
	defaultTimeFormat = time.DateTime
)

type Options struct {
	AddSource   bool
	SkipCallers int
	Level       slog.Leveler
	ReplaceAttr func(groups []string, attr slog.Attr) slog.Attr
	TimeFormat  string
	NoColor     bool
}

func NewHandler(w io.Writer, opts *Options) slog.Handler {
	h := &handler{
		w:          w,
		level:      defaultLevel,
		timeFormat: defaultTimeFormat,
	}
	if opts == nil {
		return h
	}

	h.addSource = opts.AddSource
	h.skipCallers = opts.SkipCallers
	if opts.Level != nil {
		h.level = opts.Level
	}
	h.replaceAttr = opts.ReplaceAttr
	if opts.TimeFormat != "" {
		h.timeFormat = opts.TimeFormat
	}
	h.noColor = opts.NoColor
	return h
}

// handler implements a [slog.Handler].
type handler struct {
	attrsPrefix string
	groupPrefix string
	groups      []string

	mu sync.Mutex
	w  io.Writer

	addSource   bool
	skipCallers int
	level       slog.Leveler
	replaceAttr func([]string, slog.Attr) slog.Attr
	timeFormat  string
	noColor     bool
}

func (h *handler) clone() *handler {
	return &handler{
		attrsPrefix: h.attrsPrefix,
		groupPrefix: h.groupPrefix,
		groups:      h.groups,
		w:           h.w,
		addSource:   h.addSource,
		skipCallers: h.skipCallers,
		level:       h.level,
		replaceAttr: h.replaceAttr,
		timeFormat:  h.timeFormat,
		noColor:     h.noColor,
	}
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}
func (h *handler) Handle(_ context.Context, r slog.Record) error {
	buf := zutil.NewBuffer()
	defer buf.Free()

	rep := h.replaceAttr

	// time
	if !r.Time.IsZero() {
		h.color(buf, ansiTime)
		val := r.Time.Round(0) // strip monotonic to match Attr behavior
		if rep == nil {
			*buf = r.Time.AppendFormat(*buf, h.timeFormat)
			_ = buf.WriteByte(' ')
		} else if a := rep(nil /* groups */, slog.Time(slog.TimeKey, val)); a.Key != "" {
			if a.Value.Kind() == slog.KindTime {
				*buf = a.Value.Time().AppendFormat(*buf, h.timeFormat)
			} else {
				h.appendValue(buf, a.Value, false)
			}
			_ = buf.WriteByte(' ')
		}
		h.colorEnd(buf)
	}

	// level
	h.colorLevel(buf, r.Level)
	if rep == nil {
		h.appendLevel(buf, r.Level)
		_ = buf.WriteByte(' ')
	} else if a := rep(nil /* groups */, slog.Any(slog.LevelKey, r.Level)); a.Key != "" {
		h.appendValue(buf, a.Value, false)
		_ = buf.WriteByte(' ')
	}
	h.colorEnd(buf)

	// source
	if h.addSource {
		pcs := make([]uintptr, 16)
		n := runtime.Callers(5+h.skipCallers, pcs)
		fs := runtime.CallersFrames(pcs[:n])
		f, _ := fs.Next()
		if f.File != "" {
			src := &slog.Source{
				Function: f.Function,
				File:     f.File,
				Line:     f.Line,
			}
			h.color(buf, ansiSource)
			if rep == nil {
				h.appendSource(buf, src)
			} else if a := rep(nil /* groups */, slog.Any(slog.SourceKey, src)); a.Key != "" {
				h.appendValue(buf, a.Value, false)
			}
			h.colorEnd(buf)
			_ = buf.WriteByte(' ')
		}
	}

	// message
	h.colorLevel(buf, r.Level)
	if rep == nil {
		_, _ = buf.WriteString(r.Message)
	} else if a := rep(nil /* groups */, slog.String(slog.MessageKey, r.Message)); a.Key != "" {
		h.appendValue(buf, a.Value, false)
	}
	h.colorEnd(buf)
	_ = buf.WriteByte(' ')

	// handler attributes
	if len(h.attrsPrefix) > 0 {
		_, _ = buf.WriteString(h.attrsPrefix)
	}

	// attributes
	r.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(buf, attr, h.groupPrefix, h.groups)
		return true
	})

	if len(*buf) == 0 {
		return nil
	}
	(*buf)[len(*buf)-1] = '\n' // replace last space with newline

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.w.Write(*buf)
	return err
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := h.clone()

	buf := zutil.NewBuffer()
	defer buf.Free()

	// write attributes to buffer
	for _, attr := range attrs {
		h.appendAttr(buf, attr, h.groupPrefix, h.groups)
	}
	h2.attrsPrefix = h.attrsPrefix + string(*buf)
	return h2
}
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groupPrefix += name + "."
	h2.groups = append(h2.groups, name)
	return h2
}
func (h *handler) appendLevel(buf *zutil.Buffer, level slog.Level) {
	switch {
	case level < slog.LevelInfo:
		_, _ = buf.WriteString("DBG")
		appendLevelDelta(buf, level-slog.LevelDebug)
	case level < slog.LevelWarn:
		_, _ = buf.WriteString("INF")
		appendLevelDelta(buf, level-slog.LevelInfo)
	case level < slog.LevelError:
		_, _ = buf.WriteString("WRN")
		appendLevelDelta(buf, level-slog.LevelWarn)
	default:
		_, _ = buf.WriteString("ERR")
		appendLevelDelta(buf, level-slog.LevelError)
	}
}
func appendLevelDelta(buf *zutil.Buffer, delta slog.Level) {
	if delta == 0 {
		return
	} else if delta > 0 {
		_ = buf.WriteByte('+')
	}
	*buf = strconv.AppendInt(*buf, int64(delta), 10)
}
func (h *handler) appendSource(buf *zutil.Buffer, src *slog.Source) {
	dir, file := filepath.Split(src.File)
	_, _ = buf.WriteString(filepath.Join(filepath.Base(dir), file))
	_ = buf.WriteByte(':')
	_, _ = buf.WriteString(strconv.Itoa(src.Line))
}

func (h *handler) appendAttr(buf *zutil.Buffer, attr slog.Attr, groupsPrefix string, groups []string) {
	attr.Value = attr.Value.Resolve()
	if rep := h.replaceAttr; rep != nil && attr.Value.Kind() != slog.KindGroup {
		attr = rep(groups, attr)
		attr.Value = attr.Value.Resolve()
	}

	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" {
			groupsPrefix += attr.Key + "."
			groups = append(groups, attr.Key)
		}
		for _, groupAttr := range attr.Value.Group() {
			h.appendAttr(buf, groupAttr, groupsPrefix, groups)
		}
		return
	}

	h.appendKey(buf, attr.Key, groupsPrefix)
	h.appendValue(buf, attr.Value, true)
	_ = buf.WriteByte(' ')
}

func (h *handler) appendKey(buf *zutil.Buffer, key, groups string) {
	appendString(buf, groups+key, true)
	_ = buf.WriteByte('=')
}

func (h *handler) appendValue(buf *zutil.Buffer, v slog.Value, quote bool) {
	switch v.Kind() {
	case slog.KindString:
		appendString(buf, v.String(), quote)
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		appendString(buf, v.Duration().String(), quote)
	case slog.KindTime:
		appendString(buf, v.Time().String(), quote)
	case slog.KindAny:
		switch cv := v.Any().(type) {
		case slog.Level:
			h.appendLevel(buf, cv)
		case encoding.TextMarshaler:
			data, err := cv.MarshalText()
			if err != nil {
				break
			}
			appendString(buf, string(data), quote)
		case *slog.Source:
			h.appendSource(buf, cv)
		default:
			appendString(buf, fmt.Sprintf("%+v", v.Any()), quote)
		}
	default:
	}
}

func appendString(buf *zutil.Buffer, s string, quote bool) {
	if quote && needsQuoting(s) {
		*buf = strconv.AppendQuote(*buf, s)
	} else {
		_, _ = buf.WriteString(s)
	}
}

func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		if unicode.IsSpace(r) || r == '"' || r == '=' || !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}

func (h *handler) color(buf *zutil.Buffer, ansi string) {
	if h.noColor {
		return
	}
	_, _ = buf.WriteString(ansi)
}
func (h *handler) colorLevel(buf *zutil.Buffer, level slog.Level) {
	if h.noColor {
		return
	}
	switch level {
	case slog.LevelDebug:
		_, _ = buf.WriteString(ansiDebug)
	case slog.LevelInfo:
		_, _ = buf.WriteString(ansiInfo)
	case slog.LevelWarn:
		_, _ = buf.WriteString(ansiWarn)
	case slog.LevelError:
		_, _ = buf.WriteString(ansiError)
	default:
		_, _ = buf.WriteString(ansiDebug)
	}
}
func (h *handler) colorEnd(buf *zutil.Buffer) {
	if h.noColor {
		return
	}
	_, _ = buf.WriteString(ansiReset)
}
