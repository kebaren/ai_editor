package textbuffer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler 提供性能分析功能
type Profiler struct {
	// CPU分析文件
	cpuFile *os.File
	// 内存分析文件
	memFile *os.File
	// 分析开始时间
	startTime time.Time
	// 是否正在进行CPU分析
	cpuProfiling bool
	// 是否正在进行内存分析
	memProfiling bool
	// 分析结果输出目录
	outputDir string
}

// NewProfiler 创建一个新的性能分析器
func NewProfiler(outputDir string) (*Profiler, error) {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	return &Profiler{
		outputDir: outputDir,
	}, nil
}

// StartCPUProfiling 开始CPU分析
func (p *Profiler) StartCPUProfiling() error {
	if p.cpuProfiling {
		return fmt.Errorf("CPU profiling already started")
	}

	filename := filepath.Join(p.outputDir, fmt.Sprintf("cpu_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %v", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("failed to start CPU profile: %v", err)
	}

	p.cpuFile = f
	p.cpuProfiling = true
	p.startTime = time.Now()
	return nil
}

// StopCPUProfiling 停止CPU分析
func (p *Profiler) StopCPUProfiling() error {
	if !p.cpuProfiling {
		return fmt.Errorf("CPU profiling not started")
	}

	pprof.StopCPUProfile()
	if err := p.cpuFile.Close(); err != nil {
		return fmt.Errorf("failed to close CPU profile file: %v", err)
	}

	p.cpuFile = nil
	p.cpuProfiling = false
	return nil
}

// StartMemoryProfiling 开始内存分析
func (p *Profiler) StartMemoryProfiling() error {
	if p.memProfiling {
		return fmt.Errorf("memory profiling already started")
	}

	filename := filepath.Join(p.outputDir, fmt.Sprintf("mem_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %v", err)
	}

	p.memFile = f
	p.memProfiling = true
	p.startTime = time.Now()
	return nil
}

// StopMemoryProfiling 停止内存分析
func (p *Profiler) StopMemoryProfiling() error {
	if !p.memProfiling {
		return fmt.Errorf("memory profiling not started")
	}

	runtime.GC() // 获取更准确的内存使用情况
	if err := pprof.WriteHeapProfile(p.memFile); err != nil {
		return fmt.Errorf("failed to write memory profile: %v", err)
	}

	if err := p.memFile.Close(); err != nil {
		return fmt.Errorf("failed to close memory profile file: %v", err)
	}

	p.memFile = nil
	p.memProfiling = false
	return nil
}

// TakeSnapshot 创建一个内存快照
func (p *Profiler) TakeSnapshot() error {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("snapshot_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %v", err)
	}
	defer f.Close()

	runtime.GC() // 获取更准确的内存使用情况
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write snapshot: %v", err)
	}

	return nil
}

// CollectGoroutineProfile 收集goroutine分析信息
func (p *Profiler) CollectGoroutineProfile() error {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("goroutine_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %v", err)
	}
	defer f.Close()

	if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %v", err)
	}

	return nil
}

// CollectBlockProfile 收集阻塞分析信息
func (p *Profiler) CollectBlockProfile() error {
	filename := filepath.Join(p.outputDir, fmt.Sprintf("block_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create block profile file: %v", err)
	}
	defer f.Close()

	if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %v", err)
	}

	return nil
}

// GetProfileInfo 获取分析信息
func (p *Profiler) GetProfileInfo() ProfileInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return ProfileInfo{
		NumGoroutine:    runtime.NumGoroutine(),
		NumCPU:          runtime.NumCPU(),
		AllocBytes:      memStats.Alloc,
		TotalAllocBytes: memStats.TotalAlloc,
		SysBytes:        memStats.Sys,
		NumGC:           uint64(memStats.NumGC),
		CPUProfiling:    p.cpuProfiling,
		MemProfiling:    p.memProfiling,
		StartTime:       p.startTime,
	}
}

// ProfileInfo 包含分析信息
type ProfileInfo struct {
	NumGoroutine    int
	NumCPU          int
	AllocBytes      uint64
	TotalAllocBytes uint64
	SysBytes        uint64
	NumGC           uint64
	CPUProfiling    bool
	MemProfiling    bool
	StartTime       time.Time
}

// String 返回格式化的分析信息
func (pi ProfileInfo) String() string {
	return fmt.Sprintf(
		"Profile Info:\n"+
			"  Goroutines: %d\n"+
			"  CPUs: %d\n"+
			"  Current Allocation: %d bytes\n"+
			"  Total Allocation: %d bytes\n"+
			"  System Memory: %d bytes\n"+
			"  GC Cycles: %d\n"+
			"  CPU Profiling: %v\n"+
			"  Memory Profiling: %v\n"+
			"  Start Time: %v",
		pi.NumGoroutine,
		pi.NumCPU,
		pi.AllocBytes,
		pi.TotalAllocBytes,
		pi.SysBytes,
		pi.NumGC,
		pi.CPUProfiling,
		pi.MemProfiling,
		pi.StartTime,
	)
}
