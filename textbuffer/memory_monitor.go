package textbuffer

import (
	"runtime"
	"sync/atomic"
	"time"
)

// MemoryMonitor 用于监控内存使用情况
type MemoryMonitor struct {
	// 总分配的内存
	totalAllocated uint64
	// 当前使用的内存
	currentUsage uint64
	// 内存使用的峰值
	peakUsage uint64
	// 分配次数
	allocations uint64
	// 释放次数
	deallocations uint64
	// 开始监控的时间
	startTime time.Time
	// 是否正在运行
	running bool
	// 停止通道
	stopChan chan struct{}
}

// NewMemoryMonitor 创建一个新的内存监控器
func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
	}
}

// Start 开始监控内存使用情况
func (mm *MemoryMonitor) Start() {
	if mm.running {
		return
	}
	mm.running = true
	go mm.monitor()
}

// Stop 停止监控内存使用情况
func (mm *MemoryMonitor) Stop() {
	if !mm.running {
		return
	}
	mm.running = false
	mm.stopChan <- struct{}{}
}

// monitor 监控内存使用情况
func (mm *MemoryMonitor) monitor() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 更新内存使用情况
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			// 更新当前使用的内存
			atomic.StoreUint64(&mm.currentUsage, memStats.Alloc)

			// 更新峰值内存使用
			if memStats.Alloc > atomic.LoadUint64(&mm.peakUsage) {
				atomic.StoreUint64(&mm.peakUsage, memStats.Alloc)
			}

			// 更新总分配的内存
			atomic.StoreUint64(&mm.totalAllocated, memStats.TotalAlloc)

			// 更新分配和释放次数
			atomic.StoreUint64(&mm.allocations, memStats.Mallocs)
			atomic.StoreUint64(&mm.deallocations, memStats.Frees)
		case <-mm.stopChan:
			return
		}
	}
}

// TrackAllocation 跟踪内存分配
func (mm *MemoryMonitor) TrackAllocation(size int) {
	if size <= 0 {
		return
	}

	// 更新总分配的内存
	atomic.AddUint64(&mm.totalAllocated, uint64(size))
	// 更新当前使用的内存
	newUsage := atomic.AddUint64(&mm.currentUsage, uint64(size))
	// 更新峰值内存使用
	for {
		peak := atomic.LoadUint64(&mm.peakUsage)
		if newUsage <= peak || atomic.CompareAndSwapUint64(&mm.peakUsage, peak, newUsage) {
			break
		}
	}
	// 更新分配次数
	atomic.AddUint64(&mm.allocations, 1)
}

// TrackDeallocation 跟踪内存释放
func (mm *MemoryMonitor) TrackDeallocation(size int) {
	if size <= 0 {
		return
	}

	// 更新当前使用的内存
	atomic.AddUint64(&mm.currentUsage, ^uint64(size-1))
	// 更新释放次数
	atomic.AddUint64(&mm.deallocations, 1)
}

// GetStats 获取内存使用统计信息
func (mm *MemoryMonitor) GetStats() MemoryStats {
	return MemoryStats{
		CurrentUsage:   atomic.LoadUint64(&mm.currentUsage),
		PeakUsage:      atomic.LoadUint64(&mm.peakUsage),
		TotalAllocated: atomic.LoadUint64(&mm.totalAllocated),
		Allocations:    atomic.LoadUint64(&mm.allocations),
		Deallocations:  atomic.LoadUint64(&mm.deallocations),
		UptimeSeconds:  uint64(time.Since(mm.startTime).Seconds()),
	}
}
