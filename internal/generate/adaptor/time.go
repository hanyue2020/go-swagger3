package adaptor

import "time"

type UnixMicro struct{}

func (m *UnixMicro) Gen(args ...string) any {
	return time.Now().UnixMicro()
}

type UnixMilli struct{}

func (m *UnixMilli) Gen(args ...string) any {
	return time.Now().UnixMilli()
}

type UnixNano struct{}

func (m *UnixNano) Gen(args ...string) any {
	return time.Now().UnixNano()
}

type UnixSecond struct{}

func (m *UnixSecond) Gen(args ...string) any {
	return time.Now().Unix()
}
