package main

import (
	"strings"
	"time"
)

type TimeEasyTemplate struct {
	ModeleHMS string
	Template  string
}

func (f *TimeEasyTemplate) Set(s string) error {
	f.ModeleHMS = s

	tmpl := [][]string{
		{"%Y", "2006"},
		{"%m", "01"},
		{"%d", "02"},
		{"%H", "15"},
		{"%M", "04"},
		{"%S", "05"},
	}

	for _, p := range tmpl {
		s = strings.ReplaceAll(s, p[0], p[1])
	}
	f.Template = s
	return nil
}

func (s TimeEasyTemplate) String() string {
	return s.ModeleHMS
}

func (s TimeEasyTemplate) Format(t time.Time) string {
	return t.Format(s.Template)
}
