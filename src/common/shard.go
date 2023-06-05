package common

import (
	"strconv"
	"strings"
)

type ShardedRedisDB[T any] struct {
	rdbs []*T
}

func NewShardedRedisDB[T any](rdbs ...*T) *ShardedRedisDB[T] {
	sr := &ShardedRedisDB[T]{
		rdbs: rdbs,
	}
	return sr
}

func (sr *ShardedRedisDB[T]) Select(idx int) *T {
	return sr.rdbs[idx]
}

func (sr *ShardedRedisDB[T]) ForEach(f func(*T)) {
	for _, r := range sr.rdbs {
		f(r)
	}
}

// todo: allow custom routing algo
func (sr *ShardedRedisDB[T]) Route(snowflake string) *T {
	mId := SnowflakeIDPickMachineIdFast(snowflake)
	rId, err := strconv.Atoi(mId)
	if err != nil {
		return nil
	}
	if rId-1 < 0 || rId-1 >= len(sr.rdbs) {
		return nil
	}
	return sr.rdbs[rId-1]
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
