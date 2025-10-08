//go:build windows

//written by DeepSeek-R1

package main

import (
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	// Kernel32.dll
	kernel32DLL                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateFile               = kernel32DLL.NewProc("CreateFileW")
	procDeviceIoControl          = kernel32DLL.NewProc("DeviceIoControl")
	procCloseHandle              = kernel32DLL.NewProc("CloseHandle")
	procCreateThread             = kernel32DLL.NewProc("CreateThread")
	procTerminateThread          = kernel32DLL.NewProc("TerminateThread")
	procCreateToolhelp32Snapshot = kernel32DLL.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = kernel32DLL.NewProc("Process32FirstW")
	procProcess32Next            = kernel32DLL.NewProc("Process32NextW")
	procThread32First            = kernel32DLL.NewProc("Thread32First")
	procThread32Next             = kernel32DLL.NewProc("Thread32Next")
	procOpenThread               = kernel32DLL.NewProc("OpenThread")
	procOpenProcess              = kernel32DLL.NewProc("OpenProcess")
	procSuspendThread            = kernel32DLL.NewProc("SuspendThread")
	procResumeThread             = kernel32DLL.NewProc("ResumeThread")

	// NTDLL
	ntdllDLL             = syscall.NewLazyDLL("ntdll.dll")
	procNtSuspendProcess = ntdllDLL.NewProc("NtSuspendProcess")
	procNtResumeProcess  = ntdllDLL.NewProc("NtResumeProcess")
)

const (
	TH32CS_SNAPPROCESS       = 0x00000002
	TH32CS_SNAPTHREAD        = 0x00000004
	PROCESS_TERMINATE        = 0x0001
	PROCESS_SUSPEND_RESUME   = 0x0800
	THREAD_TERMINATE         = 0x0001
	THREAD_SUSPEND_RESUME    = 0x0002
	THREAD_QUERY_INFORMATION = 0x0040
	MAX_PATH                 = 260
)

type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

type THREADENTRY32 struct {
	Size           uint32
	CntUsage       uint32
	ThreadID       uint32
	OwnerProcessID uint32
	BasePriority   int32
	DeltaPriority  int32
	Flags          uint32
}

// 获取极域进程状态
func GetMythwareProcessState() (uint32, int) {
	pid := FindMythwareProcess()
	if pid == 0 {
		return 0, -1
	}

	state := GetProcessState(pid)
	return pid, state
}

// 查找极域进程
func FindMythwareProcess() uint32 {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == 0 {
		return 0
	}
	defer procCloseHandle.Call(snapshot)

	var processEntry PROCESSENTRY32
	processEntry.Size = uint32(unsafe.Sizeof(processEntry))

	result, _, _ := procProcess32First.Call(snapshot, uintptr(unsafe.Pointer(&processEntry)))
	if result == 0 {
		return 0
	}

	for {
		processName := syscall.UTF16ToString(processEntry.ExeFile[:])
		if processName == "StudentMain.exe" {
			return processEntry.ProcessID
		}

		result, _, _ := procProcess32Next.Call(snapshot, uintptr(unsafe.Pointer(&processEntry)))
		if result == 0 {
			break
		}
	}

	return 0
}

// 获取进程状态
func GetProcessState(pid uint32) int {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPTHREAD, uintptr(pid))
	if snapshot == 0 {
		return -1
	}
	defer procCloseHandle.Call(snapshot)

	var threadEntry THREADENTRY32
	threadEntry.Size = uint32(unsafe.Sizeof(threadEntry))

	result, _, _ := procThread32First.Call(snapshot, uintptr(unsafe.Pointer(&threadEntry)))
	if result == 0 {
		return -1
	}

	hasRunningThread := false
	hasSuspendedThread := false

	for {
		if threadEntry.OwnerProcessID == pid {
			hThread, _, _ := procOpenThread.Call(THREAD_SUSPEND_RESUME|THREAD_QUERY_INFORMATION, 0, uintptr(threadEntry.ThreadID))
			if hThread != 0 {
				previousSuspendCount, _, _ := procSuspendThread.Call(hThread)

				if previousSuspendCount == 0 {
					hasRunningThread = true
				} else {
					hasSuspendedThread = true
				}
				procResumeThread.Call(hThread)
				procCloseHandle.Call(hThread)
			}
		}

		result, _, _ := procThread32Next.Call(snapshot, uintptr(unsafe.Pointer(&threadEntry)))
		if result == 0 {
			break
		}
	}
	if hasRunningThread {
		return 0
	} else if hasSuspendedThread {
		return 1
	}

	return -1
}

// 检查极域是否在运行
func IsMythwareRunning() bool {
	pid, state := GetMythwareProcessState()
	return pid != 0 && state == 0
}

// 检查极域是否暂停
func IsMythwareSuspended() bool {
	_, state := GetMythwareProcessState()
	return state == 1
}

// 结束极域进程
func KillMythwareProcess() bool {
	pid := FindMythwareProcess()
	if pid == 0 {
		return false
	}
	return KillProcessByThreads(pid)
}

func KillProcessByThreads(pid uint32) bool {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPTHREAD, uintptr(pid))
	if snapshot == 0 {
		return false
	}
	defer procCloseHandle.Call(snapshot)

	var threadEntry THREADENTRY32
	threadEntry.Size = uint32(unsafe.Sizeof(threadEntry))

	result, _, _ := procThread32First.Call(snapshot, uintptr(unsafe.Pointer(&threadEntry)))
	if result == 0 {
		return false
	}

	success := false

	for {
		if threadEntry.OwnerProcessID == pid {
			hThread, _, _ := procOpenThread.Call(THREAD_TERMINATE, 0, uintptr(threadEntry.ThreadID))
			if hThread != 0 {
				if ret, _, _ := procTerminateThread.Call(hThread, 0); ret != 0 {
					success = true
				}
				procCloseHandle.Call(hThread)
			}
		}

		result, _, _ := procThread32Next.Call(snapshot, uintptr(unsafe.Pointer(&threadEntry)))
		if result == 0 {
			break
		}
	}

	return success
}

// 暂停极域进程
func SuspendMythwareProcess() bool {
	pid := FindMythwareProcess()
	if pid == 0 {
		return false
	}

	hProcess, _, _ := procOpenProcess.Call(PROCESS_SUSPEND_RESUME, 0, uintptr(pid))
	if hProcess == 0 {
		return false
	}
	defer procCloseHandle.Call(hProcess)

	ret, _, _ := procNtSuspendProcess.Call(hProcess)
	return ret == 0
}

// 恢复极域进程
func ResumeMythwareProcess() bool {
	pid := FindMythwareProcess()
	if pid == 0 {
		return false
	}

	hProcess, _, _ := procOpenProcess.Call(PROCESS_SUSPEND_RESUME, 0, uintptr(pid))
	if hProcess == 0 {
		return false
	}
	defer procCloseHandle.Call(hProcess)

	ret, _, _ := procNtResumeProcess.Call(hProcess)
	return ret == 0
}

// 启动极域进程
func StartMythwareProcess() bool {
	path, err := GetMythwareInstallPath()
	if err != nil {
		log.Printf("获取极域安装路径失败: %v", err)
		return false
	}

	exePath := path + "\\StudentMain.exe"

	exePathPtr, _ := syscall.UTF16PtrFromString(exePath)

	var startupInfo windows.StartupInfo
	var processInfo windows.ProcessInformation

	startupInfo.Cb = uint32(unsafe.Sizeof(startupInfo))

	err = windows.CreateProcess(
		exePathPtr,
		nil,
		nil,
		nil,
		false,
		0,
		nil,
		nil,
		&startupInfo,
		&processInfo,
	)

	if err != nil {
		log.Println(err)
		return false
	}

	windows.CloseHandle(processInfo.Thread)
	windows.CloseHandle(processInfo.Process)

	return true
}

// 获取极域安装路径
func GetMythwareInstallPath() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\WOW6432Node\TopDomain\e-Learning Class Standard\1.00`,
		registry.READ)
	if err != nil {
		key, err = registry.OpenKey(registry.LOCAL_MACHINE,
			`SOFTWARE\TopDomain\e-Learning Class Standard\1.00`,
			registry.READ)
		if err != nil {
			return "", err
		}
	}
	defer key.Close()

	targetDir, _, err := key.GetStringValue("TargetDirectory")
	if err != nil {
		return "", err
	}

	return targetDir, nil
}
