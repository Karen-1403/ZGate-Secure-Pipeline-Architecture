package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
)

// JSONProtocol reads and writes JSON Lines (JSONL) messages over an io.Reader/Writer.
type JSONProtocol struct {
	r *bufio.Reader
	w *bufio.Writer
}

func NewJSONProtocol(r *bufio.Reader, w *bufio.Writer) *JSONProtocol {
	return &JSONProtocol{r: r, w: w}
}

// ReadMessage reads one JSON line and returns the raw bytes (without newline).
func (p *JSONProtocol) ReadMessage() ([]byte, error) {
	line, err := p.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// strip newline
	// Accept CRLF or LF
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
	}
	return line, nil
}

// WriteMessage writes the object as a JSON line and flushes.
func (p *JSONProtocol) WriteMessage(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	if _, err := p.w.Write(append(data, '\n')); err != nil {
		return err
	}
	return p.w.Flush()
}
