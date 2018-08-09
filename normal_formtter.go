package logrus

import (
	"bytes"
	"fmt"
	"sync"
	"runtime"
	"strconv"
	"path"
)

// TextFormatter formats logs into text
type NormalFormatter struct {
	// Force disabling colors.
	DisableColors bool
	NeedQuote bool
	//2006-01-02 15:04:05.999
	TimeFormat string
	//%T [%L](%D):%M
	Partten  string
	parttenSplite [][]byte
	sync.Once
}

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
func  (f *NormalFormatter)LogDebug() string{
	calldeep:=8
	_, file, line, ok := runtime.Caller(calldeep)
	if !ok {

		file = "???"
		line = 0
	}
	_, filename := path.Split(file)
	str:= fmt.Sprintf("go:%d %s:%s", GetGID(),filename , strconv.Itoa(line))
	return str
}
func (f *NormalFormatter) init(entry *Entry) {
	if f.Partten=="" {
		f.Partten = `%T [%L](%D):%M`
		f.TimeFormat = `2006-01-02 15:04:05.999`
	}
}


func (f *NormalFormatter)logPartten(out *bytes.Buffer,entry *Entry) {
	if f.parttenSplite==nil {
		f.parttenSplite = bytes.Split([]byte(f.Partten), []byte{'%'})
	}
	// Iterate over the pieces, replacing known formats
	for i, piece := range f.parttenSplite {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'T':
				tt:=entry.Time.Format(f.TimeFormat)
				n,_:=out.WriteString(tt)
				pad:=len(f.TimeFormat)-n
				if pad>0 {
					for i:=0;i<pad ;i++  {
						out.WriteString(" ")
					}
				}
			case 'D':
				out.WriteString(f.LogDebug())
			case 'L':
				out.WriteString(entry.Level.String())
			case 'M':
				out.WriteString(entry.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}

	out.WriteByte('\n')

}
// Format renders a single log entry
func (f *NormalFormatter) Format(entry *Entry) ([]byte, error) {
	b:=entry.Buffer
	if b == nil {
		b = &bytes.Buffer{}
	}
	f.Do(func() { f.init(entry) })

	f.logPartten(b,entry)

	return b.Bytes(), nil
}





