package filelog

import (
	"fmt"
	"os"
	"github.com/lestrrat/go-strftime"
	"github.com/lq277012525/logrus"
	"time"
	"path"
)

type HookCheck interface {
	CheckLog(entry *logrus.Entry) error
} 
type IFileHook struct {
	HookCheck
	out    *os.File
}
func (hook *IFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (hook *IFileHook) Fire(entry *logrus.Entry) error {
    err :=hook.HookCheck.CheckLog(entry)
	if err!=nil {
		return err
	}
	hook.out.Write(entry.FormatMsg)
	hook.out.Sync()
	return nil
}

type FileHook struct {
	IFileHook
	pattern  string   //baseName+ "_%Y%m%d.log"
	globPattern *strftime.Strftime
	lastWrite time.Time
}

type SimFileHook struct {
	Name string
	CurSize int64
	LimitSize int64
	IFileHook
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
	rl.IFileHook.HookCheck=rl
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
func (hook *FileHook) CheckLog(entry *logrus.Entry) error{
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
	return nil
}

func NewSimFileHook(Name string,limitsize int64) *SimFileHook {
	 r:=&SimFileHook{Name:Name,LimitSize:limitsize}
	 r.IFileHook.HookCheck=r
	 return r
}
func (hook *SimFileHook) CheckLog(entry *logrus.Entry) error{
	if hook.LimitSize>0 && hook.CurSize> hook.LimitSize{
		if hook.out!=nil {
			hook.out.Close()
		}
		hook.out=nil
	}
	if hook.out==nil  {
		fd,err:=os.OpenFile(hook.Name,os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to OpenFile %s, %v\n",hook.Name ,err)
			if  hook.out==nil {
				return err
			}
		}
		hook.out=fd
		hook.CurSize=0
		if hook.CurSize >0 {
			fd.Truncate(0)
		}else {
			sz,er:=fd.Stat()
			if er==nil {
				hook.CurSize=sz.Size()
			}
		}
	}
	if  hook.LimitSize>=0 {
		hook.CurSize =hook.CurSize + int64(len(entry.FormatMsg))
	}
	return nil
}

