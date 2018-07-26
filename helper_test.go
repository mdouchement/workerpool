package workerpool_test

import "github.com/mdouchement/workerpool"

func init() {
	// Hide logs during tests
	workerpool.SetLogger(&nullLogger{})
}

type nullLogger struct {
}

func (l *nullLogger) Print(v ...interface{}) {
}

func (l *nullLogger) Printf(s string, v ...interface{}) {
}

func (l *nullLogger) Println(v ...interface{}) {
}

func (l *nullLogger) Fatal(v ...interface{}) {
}

func (l *nullLogger) Fatalf(s string, v ...interface{}) {
}

func (l *nullLogger) Fatalln(v ...interface{}) {
}

func (l *nullLogger) Panic(v ...interface{}) {
}

func (l *nullLogger) Panicf(s string, v ...interface{}) {
}

func (l *nullLogger) Panicln(v ...interface{}) {
}
