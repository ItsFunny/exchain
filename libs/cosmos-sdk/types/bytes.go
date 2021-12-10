package types

import (
	"fmt"
	"sync"
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
	mu          sync.Mutex
	StorageAll  uint64
	StorageTime time.Duration
}

func (s *StorageManager) UpdateTime(ts time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StorageTime += ts
	s.StorageAll++
}

func (s *StorageManager) Log(extra string) {
	fmt.Println("StorageM", extra, "StorageGetCount", s.StorageAll, "StorageGetTimes", s.StorageTime.Milliseconds())
}

func (s *StorageManager) Clean() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StorageAll = 0
	s.StorageTime = time.Duration(0)
}

var (
	MStorage = &StorageManager{}
)
