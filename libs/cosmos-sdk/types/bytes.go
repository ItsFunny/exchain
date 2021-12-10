package types

import (
	"fmt"
	"time"
)

// copy bytes
func CopyBytes(bz []byte) (ret []byte) {
	if bz == nil {
		return nil
	}
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

type StorageManager struct {
	StorageAll  uint64
	StorageTime time.Duration
}

func (s *StorageManager) UpdateTime(ts time.Duration) {
	s.StorageTime += ts
	s.StorageAll++
}

func (s *StorageManager) Log(extra string) {
	fmt.Println("StorageM", extra, "allCnt", s.StorageAll, "allTime", s.StorageTime)
}

func (s *StorageManager) Clean() {
	s.StorageAll = 0
	s.StorageTime = time.Duration(0)
}

var (
	MStorage = &StorageManager{}
)
