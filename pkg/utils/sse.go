package utils

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// SSEEvent 表示一个SSE事件
type SSEEvent struct {
	Event string
	Data  string
	ID    string
	Retry int
}

// SSEReader 是一个SSE事件流解析器
type SSEReader struct {
	reader    *bufio.Reader
	delimiter []byte
	event     SSEEvent
	buffer    []byte
}

// NewSSEReader 创建一个新的SSE读取器
func NewSSEReader(reader io.Reader) *SSEReader {
	return &SSEReader{
		reader:    bufio.NewReader(reader),
		delimiter: []byte{'\n', '\n'}, // SSE事件之间使用两个换行符分隔
		event:     SSEEvent{},
	}
}

// ReadEvent 读取SSE流中的下一个事件
func (r *SSEReader) ReadEvent() (*SSEEvent, error) {
	r.event = SSEEvent{} // 重置事件对象

	for {
		line, err := r.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// 如果读取到末尾但缓冲区中还有数据，处理最后一个事件
				if len(r.buffer) > 0 && !bytes.Equal(r.buffer, []byte("\n")) {
					r.buffer = append(r.buffer, '\n')
					event := r.processBuffer()
					r.buffer = nil
					if event.Data != "" || event.Event != "" {
						return &event, nil
					}
				}
			}
			return nil, err
		}

		// 将读取的行添加到缓冲区
		r.buffer = append(r.buffer, line...)

		// 检查是否有完整的事件（即双换行符）
		if bytes.HasSuffix(r.buffer, r.delimiter) || (len(line) == 1 && line[0] == '\n' && len(r.buffer) > 1 && r.buffer[len(r.buffer)-2] == '\n') {
			event := r.processBuffer()
			r.buffer = nil
			if event.Data != "" || event.Event != "" {
				return &event, nil
			}
		}
	}
}

// processBuffer 处理缓冲区中的数据，解析SSE事件
func (r *SSEReader) processBuffer() SSEEvent {
	lines := bytes.Split(r.buffer, []byte{'\n'})
	event := SSEEvent{}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// 处理注释行
		if line[0] == ':' {
			continue
		}

		// 分割字段和值
		parts := bytes.SplitN(line, []byte{':'}, 2)
		if len(parts) != 2 {
			continue
		}

		field := string(parts[0])
		value := string(parts[1])
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:] // 去除第一个空格
		}

		// 根据字段类型设置事件属性
		switch field {
		case "event":
			event.Event = value
		case "data":
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += value
		case "id":
			event.ID = value
		case "retry":
			// retry字段通常是一个整数，但我们这里简化处理
			event.Retry = 3000 // 默认3秒
		}
	}

	return event
}

// ParseSSEData 用于解析JSON格式的SSE数据
func ParseSSEData(data string) string {
	// 一些API会在data前面加上"data: "，需要去除
	if strings.HasPrefix(data, "data: ") {
		data = strings.TrimPrefix(data, "data: ")
	}
	return data
}
