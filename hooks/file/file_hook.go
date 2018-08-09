package filelog

import (
	"fmt"
	"os"
	"github.com/lestrrat/go-strftime"
	"github.com/lq277012525/logrus"
	"time"
	"path"
)

type FileHook struct {
	pattern  string   //baseName+ "_%Y%m%d.log"
	out    *os.File
	globPattern *strftime.Strftime
	lastWrite time.Time
}
func NewFileHook(baseName string) *FileHook{
	strfobj, err := strftime.New(baseName)
	if err!=nil {
		fmt.Fprintf(os.Stderr, "Unable to read baseName, %v", err)
		return nil
	}
	rl:=  &FileHook{}
	rl.pattern = baseName
	rl.globPattern = strfobj
	rl.out=nil
	rl.lastWrite=time.Now()
	pa,_:=path.Split(baseName)
	if pa!="" {
		err:=os.MkdirAll(pa,0644)
		if err!=nil {
			fmt.Fprintln(os.Stderr, "Unable to Mkdir, path:", pa)
		}
	}
	return rl
}

func IsDiffDay(time1,time2 time.Time)bool{
	Year,Month,Day:=time1.Date()
	tm1 := time.Date(Year, Month,Day+1, 0, 0, 0, 0, time1.Location()).Unix()
	tm2 := time.Date(Year, Month,Day-1, 0, 0, 0, 0, time1.Location()).Unix()
	if time2.Unix() < tm1 && time2.Unix()>= tm2{
		return false
	}
	return true
}
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	if hook.out==nil || IsDiffDay(entry.Time,hook.lastWrite) {
		filename:=hook.globPattern.FormatString(entry.Time)
		fd,err:=os.OpenFile(filename,os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to OpenFile %s, %v\n",filename ,err)
			if  hook.out==nil {
				return err
			}
		}
		if hook.out!=nil && err==nil {
			hook.out.Close()
		}
		hook.lastWrite=entry.Time
		hook.out=fd
	}
	
	hook.out.Write(entry.FormatMsg)
	hook.out.Sync()
	return nil
}
func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
