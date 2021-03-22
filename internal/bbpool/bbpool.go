package bbpool

import (
	"bytes"
	"errors"
	"sync"
)

var pool *sync.Pool
var once sync.Once

func Get() *bytes.Buffer {
	once.Do(initPool)

	buf, ok := pool.Get().(*bytes.Buffer)
	if !ok {
		panic(errors.New("bbpool: cannot get buffer from pool"))
	}
	return buf
}

func Put(buf *bytes.Buffer) {
	once.Do(initPool)

	buf.Reset()
	pool.Put(buf)
}

func initPool() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
}
