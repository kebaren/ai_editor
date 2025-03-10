package textbuffer

import (
	"sync"
)

// MemoryPool 是一个简单的内存池实现，用于重用已分配的内存
type MemoryPool struct {
	pools     map[int]*sync.Pool
	maxSize   int
	sizeSteps []int
}

// NewMemoryPool 创建一个新的内存池
func NewMemoryPool(maxSize int) *MemoryPool {
	// 定义内存池的大小步长
	sizeSteps := []int{
		128,     // 128 bytes
		512,     // 512 bytes
		1024,    // 1 KB
		4096,    // 4 KB
		16384,   // 16 KB
		65536,   // 64 KB
		262144,  // 256 KB
		1048576, // 1 MB
		4194304, // 4 MB
	}

	pools := make(map[int]*sync.Pool)
	for _, size := range sizeSteps {
		size := size // 避免闭包问题
		pools[size] = &sync.Pool{
			New: func() interface{} {
				return make([]rune, 0, size)
			},
		}
	}

	return &MemoryPool{
		pools:     pools,
		maxSize:   maxSize,
		sizeSteps: sizeSteps,
	}
}

// GetBuffer 获取一个指定大小的缓冲区
func (mp *MemoryPool) GetBuffer(minSize int) []rune {
	if minSize > mp.maxSize {
		// 如果请求的大小超过最大值，直接分配新的内存
		return make([]rune, 0, minSize)
	}

	// 找到最接近的大小步长
	var poolSize int
	for _, size := range mp.sizeSteps {
		if size >= minSize {
			poolSize = size
			break
		}
	}

	if poolSize == 0 {
		// 如果没有找到合适的大小，使用最大的步长
		poolSize = mp.sizeSteps[len(mp.sizeSteps)-1]
	}

	// 从对应的池中获取缓冲区
	buf := mp.pools[poolSize].Get().([]rune)
	return buf[:0] // 返回长度为0的切片，保持容量
}

// PutBuffer 将缓冲区放回池中
func (mp *MemoryPool) PutBuffer(buf []rune) {
	if cap(buf) > mp.maxSize {
		// 如果缓冲区太大，不放回池中
		return
	}

	// 找到合适的池
	var poolSize int
	for _, size := range mp.sizeSteps {
		if size >= cap(buf) {
			poolSize = size
			break
		}
	}

	if poolSize > 0 {
		mp.pools[poolSize].Put(buf)
	}
}

// GetString 从字符串创建一个缓冲区
func (mp *MemoryPool) GetString(s string) []rune {
	runes := mp.GetBuffer(len(s))
	return append(runes, []rune(s)...)
}

// Stats 返回内存池的统计信息
type PoolStats struct {
	Size       int
	BufferSize int
	InUse      int
}

// GetStats 获取内存池的统计信息
func (mp *MemoryPool) GetStats() map[int]PoolStats {
	stats := make(map[int]PoolStats)
	for size := range mp.pools {
		stats[size] = PoolStats{
			Size:       size,
			BufferSize: size * 4, // 每个rune是4字节
		}
	}
	return stats
}
