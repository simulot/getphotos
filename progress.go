package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func PleaseWait(msg string) (done func()) {
	marks := []string{".  ", ".. ", "..."}
	c := make(chan interface{})
	go func() {
		ticks := time.NewTicker(250 * time.Millisecond)
		i := 0
		for {
			select {
			case <-c:
				fmt.Print("\r", msg, strings.Repeat(" ", len(marks[i]))+"\n")
				return
			case <-ticks.C:
				fmt.Print("\r", msg, marks[i])
			}
			i++
			i = i % len(marks)
		}
	}()
	isClosed := false
	return func() {
		if !isClosed {
			close(c)
			isClosed = true
		}
	}
}
func Progression(msg string) (setprogression func(float32), done func()) {
	p := float32(0)

	c := make(chan interface{})
	go func() {
		ticks := time.NewTicker(250 * time.Millisecond)
		for {
			select {
			case <-c:
				fmt.Print("\r", msg, fmt.Sprintf(" %3.f%%", p), "\n")
				return
			case <-ticks.C:
				fmt.Print("\r", msg, fmt.Sprintf(" %3.f%%", p))
			}
		}
	}()
	isClosed := false
	done = func() {
		if !isClosed {
			close(c)
			isClosed = true
		}
	}
	setprogression = func(newp float32) {
		p = newp * 100.0
	}
	return
}

type progressionWriter struct {
	io.Writer
	length         int
	written        int
	setProgression func(p float32)
}

func NewProgressionWriter(w io.Writer, length int, pf func(float32)) *progressionWriter {
	return &progressionWriter{
		Writer:         w,
		length:         length,
		setProgression: pf,
	}
}

func (pw *progressionWriter) Write(b []byte) (n int, err error) {
	n, err = pw.Writer.Write(b)
	pw.written += n
	pw.setProgression(float32(pw.written) / float32(pw.length))
	return
}
