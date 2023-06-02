package common

import (
	"fmt"
	"time"
)

func NowString() string {
	now := time.Now()
	return fmt.Sprintf("[%02v:%02v:%03v]", now.Minute(), now.Second(), now.UnixMilli()%1000)
}
