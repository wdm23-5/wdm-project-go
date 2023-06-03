package common

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// distributed id generator

const wdmEpoch int64 = 1672531200000 // time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

// modified id generator based on Twitter Snowflake
type SnowflakeGenerator struct {
	lock      SpinLock
	timestamp int64
	machineId uint16
	sequence  uint16
}

func NewSnowFlakeGenerator(machineId string) *SnowflakeGenerator {
	return &SnowflakeGenerator{
		timestamp: time.Now().UnixMilli() - wdmEpoch,
		machineId: uint16(MustS2I(machineId)),
		sequence:  0,
	}
}

func (sg *SnowflakeGenerator) Next() SnowflakeID {
	sg.lock.Lock()
	defer sg.lock.Unlock()
	timestamp := time.Now().UnixMilli() - wdmEpoch
	if timestamp < sg.timestamp {
		panic("clock moves backwards")
	}
	if timestamp == sg.timestamp {
		sg.sequence = sg.sequence + 1
		if sg.sequence == 0 {
			for timestamp <= sg.timestamp {
				timestamp = time.Now().UnixMilli() - wdmEpoch
			}
		}
	} else {
		sg.sequence = 0
	}
	sg.timestamp = timestamp
	return SnowflakeID{timestamp: sg.timestamp, machineId: sg.machineId, sequence: sg.sequence}
}

type SnowflakeID struct {
	timestamp int64
	machineId uint16
	sequence  uint16
}

var snowflakeIDError = errors.New("invalid Snowflake ID")
var snowflakeRegex = regexp.MustCompile(`t(\d+)m(\d+)s(\d+)`)

func MakeSnowflakeID(sid string) (SnowflakeID, error) {
	match := snowflakeRegex.FindStringSubmatch(sid)
	if len(match) != 4 {
		return SnowflakeID{}, snowflakeIDError
	}
	timestamp, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil || timestamp < 0 {
		return SnowflakeID{}, snowflakeIDError
	}
	machineId, err := strconv.ParseUint(match[2], 10, 16)
	if err != nil || machineId > math.MaxUint16 {
		return SnowflakeID{}, snowflakeIDError
	}
	sequence, err := strconv.ParseUint(match[3], 10, 16)
	if err != nil || sequence > math.MaxUint16 {
		return SnowflakeID{}, snowflakeIDError
	}
	return SnowflakeID{timestamp: timestamp, machineId: uint16(machineId), sequence: uint16(sequence)}, nil
}

// faster but not 100% safe
func SnowflakeIDPickMachineIdFast(sid string) string {
	sb := strings.Builder{}
	start := false
	for _, ru := range sid {
		if ru == 'm' {
			start = true
			continue
		}
		if ru == 's' {
			break
		}
		if start {
			sb.WriteRune(ru)
		}
	}
	return sb.String()
}

// since redis stores int as string, this implementation directly uses string as id
func (sid SnowflakeID) String() string {
	return fmt.Sprintf("t%vm%vs%v", sid.timestamp, sid.machineId, sid.sequence)
}

// spin lock

type SpinLock struct {
	locked atomic.Int32
}

func (sl *SpinLock) Lock() {
	for !sl.locked.CompareAndSwap(0, 1) {
		runtime.Gosched()
	}
}

func (sl *SpinLock) Unlock() {
	if val := sl.locked.Add(-1); val != 0 {
		panic("SpinLock double unlock")
	}
}
