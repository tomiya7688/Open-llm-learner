package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const DefaultMaxMessageBytes = 16 << 20

type Decoder struct {
	scanner *bufio.Scanner
}

func NewDecoder(r io.Reader, maxMessageBytes int) *Decoder {
	if maxMessageBytes <= 0 {
		maxMessageBytes = DefaultMaxMessageBytes
	}
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), maxMessageBytes)
	return &Decoder{scanner: s}
}

func (d *Decoder) Decode(dst *Message) error {
	if !d.scanner.Scan() {
		if err := d.scanner.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	if err := json.Unmarshal(d.scanner.Bytes(), dst); err != nil {
		return fmt.Errorf("decode protocol message: %w", err)
	}
	if dst.Type == "" {
		return fmt.Errorf("decode protocol message: missing type")
	}
	return nil
}

func Encode(w io.Writer, msg Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encode protocol message: %w", err)
	}
	b = append(b, '\n')
	_, err = w.Write(b)
	return err
}
