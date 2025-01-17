package ntp

import (
	"github.com/beevik/ntp"
	"time"
)

const (
	DisplayFormat = "2006/01/02 15:04"
)

func IsDatetimeSynced() (bool, time.Time, error) {
	t, err := ntp.Time("ntp.aliyun.com")
	if err != nil {
		return false, time.Time{}, newError().Base(err)
	}
	if seconds := t.Sub(time.Now().UTC()).Seconds(); seconds >= 90 || seconds <= -90 {
		return false, t, nil
	}
	return true, t, nil
}
