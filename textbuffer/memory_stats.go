package textbuffer

import (
	"fmt"
	"strings"
	"time"
)

var (
	// 程序启动时间
	startTime = time.Now()
)

// MemoryStats 包含内存使用的统计信息
type MemoryStats struct {
	CurrentUsage   uint64
	PeakUsage      uint64
	TotalAllocated uint64
	Allocations    uint64
	Deallocations  uint64
	UptimeSeconds  uint64
}

// String 返回格式化的统计信息
func (ms MemoryStats) String() string {
	var sb strings.Builder
	sb.WriteString("Memory Stats:\n")
	sb.WriteString(fmt.Sprintf("      Current Usage: %d bytes\n", ms.CurrentUsage))
	sb.WriteString(fmt.Sprintf("      Peak Usage: %d bytes\n", ms.PeakUsage))
	sb.WriteString(fmt.Sprintf("      Total Allocated: %d bytes\n", ms.TotalAllocated))

	if ms.Allocations > 0 {
		sb.WriteString(fmt.Sprintf("      Allocations: %d\n", ms.Allocations))
	}

	if ms.Deallocations > 0 {
		sb.WriteString(fmt.Sprintf("      Deallocations: %d\n", ms.Deallocations))
	}

	if ms.UptimeSeconds > 0 {
		sb.WriteString(fmt.Sprintf("      Uptime: %d seconds\n", ms.UptimeSeconds))
	}

	return sb.String()
}
