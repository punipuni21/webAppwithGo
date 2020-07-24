package trace

import (
	"fmt"
	"io"
)

// Tracer はコードないでの出来事を記録出来るオブジェクトを表すインタフェース
type Tracer interface {
	Trace(...interface{})
}

//New は
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	t.out.Write([]byte(fmt.Sprint(a...))) //io.Writerは[]byte型なので型変換
	t.out.Write([]byte("\n"))
}

type nilTracer struct{}

func (t *nilTracer) Trace(a ...interface{}) {}

func Off() Tracer {
	return &nilTracer{}
}
