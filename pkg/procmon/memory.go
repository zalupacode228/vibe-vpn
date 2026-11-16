package procmon

import (
	"os"
	"runtime/debug"
	"syscall"
	"unsafe"
)

var (
	procGetCurrentProcess        = modKernel32.NewProc("GetCurrentProcess")
	procSetProcessWorkingSetSize = modKernel32.NewProc("SetProcessWorkingSetSize")
)

const (
	PROCESS_SET_QUOTA = 0x0100
)

// TrimMemory forces Go runtime to release unused heap memory back to Windows
// and flushes the working set of the main process and all descendant processes
// (such as WebView2 browser, GPU, renderer, and utility processes).
func TrimMemory() {
	debug.FreeOSMemory()

	curH, _, _ := procGetCurrentProcess.Call()
	if curH != 0 {
		procSetProcessWorkingSetSize.Call(curH, ^uintptr(0), ^uintptr(0))
	}

	myPID := uint32(os.Getpid())
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(uintptr(TH32CS_SNAPPROCESS), 0)
	if snapshot == uintptr(syscall.InvalidHandle) {
		return
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32W
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	type procNode struct {
		pid    uint32
		parent uint32
	}
	var allProcs []procNode

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	for ret != 0 {
		allProcs = append(allProcs, procNode{pid: entry.Th32ProcessID, parent: entry.Th32ParentProcessID})
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	}

	// Recursively collect all descendant processes (children, grandchildren, etc.)
	descendants := make(map[uint32]bool)
	queue := []uint32{myPID}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, p := range allProcs {
			if p.parent == curr && !descendants[p.pid] {
				descendants[p.pid] = true
				queue = append(queue, p.pid)
			}
		}
	}

	for pid := range descendants {
		childH, _, _ := procOpenProcess.Call(uintptr(PROCESS_SET_QUOTA|PROCESS_QUERY_LIMITED_INFORMATION), 0, uintptr(pid))
		if childH != 0 {
			procSetProcessWorkingSetSize.Call(childH, ^uintptr(0), ^uintptr(0))
			procCloseHandle.Call(childH)
		}
	}
}
