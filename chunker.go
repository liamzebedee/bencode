package bencode

import (
	"bufio"
	"errors"
	"io"

	"strconv"
)

type chunker struct {
	r    *bufio.Reader
	errd bool
}

func newChunker(r io.Reader) *chunker {
	return &chunker{r: bufio.NewReader(r)}
}

func (c *chunker) nextValue() (string, error) {
	//peek a byte and figure out
	b, err := c.r.Peek(1)
	if err != nil {
		return "", err
	}
	switch b[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return c.nextString()
	case 'i':
		return c.nextInt()
	case 'l':
		return c.nextList()
	case 'd':
		return c.nextDict()
	}

	return "", errors.New("Unexpected delimiter")
}

func (c *chunker) nextString() (string, error) {
	//read until the :
	num, err := c.r.ReadString(':')
	if err != nil {
		return "", err
	}

	n, err := strconv.ParseInt(num[:len(num)-1], 10, 64)
	if err != nil {
		return "", err
	}

	buf := make([]byte, n)
	var p int64
	for p != n {
		nr, err := c.r.Read(buf[p:])
		if err != nil {
			return "", err
		}
		p += int64(nr)
	}

	return num + string(buf), nil
}

func (c *chunker) nextInt() (string, error) {
	bs, err := c.r.Peek(1)
	if err != nil {
		return "", err
	}

	if bs[0] != 'i' {
		return "", errors.New("Attempted to read a non-int value from nextInt")
	}

	val, err := c.r.ReadString('e')
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *chunker) nextList() (string, error) {
	//read off the beginning delimiter
	b, err := c.r.ReadByte()
	if err != nil {
		return "", err
	}
	if b != 'l' {
		return "", errors.New("Attempted to read a non-list value from nextList")
	}

	buf := []byte{b}
	for {
		bs, err := c.r.Peek(1)
		if err != nil {
			return "", err
		}
		//peek an e
		if bs[0] == 'e' {
			//consume it
			_, err := c.r.ReadByte()
			if err != nil {
				return "", err
			}
			buf = append(buf, 'e')

			break
		}

		nv, err := c.nextValue()
		if err != nil {
			return "", err
		}

		buf = append(buf, []byte(nv)...)
	}

	return string(buf), nil
}

func (c *chunker) nextDict() (string, error) {
	b, err := c.r.ReadByte()
	if err != nil {
		return "", err
	}
	if b != 'd' {
		return "", errors.New("Attempted to read a non-dict value from nextDict")
	}

	buf := []byte{b}
	for {
		bs, err := c.r.Peek(1)
		if err != nil {
			return "", err
		}

		if bs[0] == 'e' {
			//consume it
			_, err := c.r.ReadByte()
			if err != nil {
				return "", err
			}
			buf = append(buf, 'e')

			break
		}

		switch bs[0] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return "", errors.New("Key is not a string")
		}

		ns, err := c.nextString()
		if err != nil {
			return "", err
		}

		buf = append(buf, []byte(ns)...)

		nv, err := c.nextValue()
		if err != nil {
			return "", err
		}

		buf = append(buf, []byte(nv)...)
	}

	return string(buf), nil
}
