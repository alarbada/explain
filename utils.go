package main

import (
	"fmt"
	"runtime"

	"github.com/sashabaranov/go-openai"
)

func wrapErr(err *error) {
	if e := *err; e != nil {
		_, file, line, _ := runtime.Caller(2)
		*err = fmt.Errorf("%v:%v %w", file, line, e)	
	}
}

type Msg = openai.ChatCompletionMessage

