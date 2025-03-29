package zlog

import (
	"bufio"
	"io"
	"runtime"
	"strings"
)

// SafeWriter
// @Description: 可能会有超长日志的情况下使用，只支持info
// @return *io.PipeWriter
func SafeWriter(ops *Options, w ...io.Writer) *io.PipeWriter {
	reader, writer := io.Pipe()
	go scan(NewZLogger(ops, w...), reader)
	runtime.SetFinalizer(writer, writerFinalizer)
	return writer
}

func scan(logger *Logger, reader *io.PipeReader) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(scanLinesOrGiveLong)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) != "" {
			logger.Info(text)
		}
	}
	_ = reader.Close()
}

const maxTokenLength = bufio.MaxScanTokenSize / 2

func scanLinesOrGiveLong(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanLines(data, atEOF)
	if advance > 0 || token != nil || err != nil {
		return
	}
	if len(data) < maxTokenLength {
		return
	}
	return maxTokenLength, data[0:maxTokenLength], nil
}

func writerFinalizer(writer *io.PipeWriter) {
	_ = writer.Close()
}
