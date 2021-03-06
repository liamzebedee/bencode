package bencode

import (
	"errors"
	"fmt"
	"io"

	"strconv"
)

func nextValue(l *lexer) (interface{}, error) {
	var next token
	switch next = l.peekToken(); next.typ {
	case intType:
		next = l.nextToken() //consume token
		n, err := strconv.Atoi(next.val)
		if err != nil {
			return nil, err
		}
		return n, nil
	case stringType:
		next = l.nextToken() //consume token
		return next.val, nil
	case listStartType:
		return consumeList(l)
	case dictStartType:
		return consumeDict(l)
	case eofType:
		return nil, io.EOF
	case errorType:
		return nil, next
	}

	return nil, fmt.Errorf("Unknown type: %s", next.typ)
}

func consumeDict(l *lexer) (map[string]interface{}, error) {
	head := l.nextToken()
	if head.typ != dictStartType {
		return nil, fmt.Errorf("Can't consume dict. Found: %s", head.typ)
	}
	ret := make(map[string]interface{})

	for {
		key := l.nextToken()
		switch key.typ {
		case dictEndType:
			return ret, nil
		case eofType:
			return nil, errors.New("Unexpected EOF")
		case errorType:
			return nil, key
		}

		switch l.peekToken().typ {
		case eofType:
			return nil, errors.New("Unexpected EOF")
		case errorType:
			return nil, l.nextToken() //consume the token
		case dictEndType:
			return nil, errors.New("Unexpected Dict End")
		case listEndType:
			return nil, errors.New("Unexpected List End")
		}

		val, err := nextValue(l)
		if err != nil {
			return nil, err
		}
		ret[key.val] = val
	}

	panic("unreachable")
}

func consumeList(l *lexer) ([]interface{}, error) {
	head := l.nextToken()
	if head.typ != listStartType {
		return nil, fmt.Errorf("Can't consume list. Found: %s", head.typ)
	}

	ret := make([]interface{}, 0)
	for {
		switch next := l.peekToken(); next.typ {
		case eofType:
			return nil, errors.New("Unexpected EOF")
		case errorType:
			return nil, next
		case dictEndType:
			return nil, errors.New("Unexpected Dict End")
		case listEndType:
			//consume it
			l.nextToken()
			return ret, nil
		}

		val, err := nextValue(l)
		if err != nil {
			return nil, err
		}
		ret = append(ret, val)
	}

	panic("unreachable")
}
