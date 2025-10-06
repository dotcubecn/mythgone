//go:build windows

package main

import (
	"log"
	"syscall"
	"time"
	"unsafe"

	"github.com/tailscale/walk"
	. "github.com/tailscale/walk/declarative"
	"github.com/tailscale/win"
)

const (
	// 窗口显示亲和性常量
	WindowDisplayAffinityNone               = 0x00000000
	WindowDisplayAffinityMonitor            = 0x00000001
	WindowDisplayAffinityExcludeFromCapture = 0x00000011

	// 窗口层级常量
	WindowHandleTop        = 0
	WindowHandleTopMost    = ^uintptr(0) // -1
	WindowHandleNotTopMost = ^uintptr(1) // -2
	WindowHandleBottom     = 1

	// Windows 消息常量
	MessageCommand       = 0x0111
	ButtonClick          = 0x00F5
	WM_WINDOWPOSCHANGING = 0x0046
	WM_MOUSEACTIVATE     = 0x0021
	WM_NCACTIVATE        = 0x0086
	WM_ACTIVATE          = 0x0006

	// 窗口样式常量
	GWL_WNDPROC = ^uintptr(3)  // -4
	GWL_EXSTYLE = ^uintptr(19) // -20

	WS_EX_NOACTIVATE = 0x08000000
	MA_NOACTIVATE    = 0x0003

	// Windows 钩子类型
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
	WH_KEYBOARD    = 2
)

var (
	// User32.dll
	user32DLL                    = syscall.NewLazyDLL("user32.dll")
	procSetWindowDisplayAffinity = user32DLL.NewProc("SetWindowDisplayAffinity")
	procSetWindowPos             = user32DLL.NewProc("SetWindowPos")
	procBringWindowToTop         = user32DLL.NewProc("BringWindowToTop")
	procFindWindow               = user32DLL.NewProc("FindWindowW")
	procPostMessage              = user32DLL.NewProc("PostMessageW")
	procSetWindowLong            = user32DLL.NewProc("SetWindowLongW")
	procCallWindowProc           = user32DLL.NewProc("CallWindowProcW")
	procGetWindowLong            = user32DLL.NewProc("GetWindowLongW")
	procSetWinEventHook          = user32DLL.NewProc("SetWinEventHook")
	procUnhookWinEvent           = user32DLL.NewProc("UnhookWinEvent")
	procSetParent                = user32DLL.NewProc("SetParent")
	procShowWindow               = user32DLL.NewProc("ShowWindow")
	procCreateWindowEx           = user32DLL.NewProc("CreateWindowExW")
	procDestroyWindow            = user32DLL.NewProc("DestroyWindow")
	procMoveWindow               = user32DLL.NewProc("MoveWindow")
	procSetWindowsHookEx         = user32DLL.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx      = user32DLL.NewProc("UnhookWindowsHookEx")
	procClipCursor               = user32DLL.NewProc("ClipCursor")

	// Kernel32.dll
	kernel32DLL         = syscall.NewLazyDLL("kernel32.dll")
	procCreateFile      = kernel32DLL.NewProc("CreateFileW")
	procDeviceIoControl = kernel32DLL.NewProc("DeviceIoControl")
	procCloseHandle     = kernel32DLL.NewProc("CloseHandle")
	procCreateThread    = kernel32DLL.NewProc("CreateThread")
	procTerminateThread = kernel32DLL.NewProc("TerminateThread")

	// 广播窗口相关变量
	originalBroadcastWindowProc uintptr
	broadcastHookEnabled        bool
	broadcastWindowHandle       syscall.Handle
	eventHook                   syscall.Handle

	// 黑屏窗口相关变量
	blackScreenMinimizeEnabled bool
	blackScreenQuitChannel     chan struct{}
	blackScreenParentWindow    syscall.Handle
	embeddedBlackScreenWindow  syscall.Handle

	// 解锁相关变量
	mouseUnlockEnabled    bool
	keyboardUnlockEnabled bool
	mouseHookThread       syscall.Handle
	keyboardHookThread    syscall.Handle
	mouseHookQuit         chan struct{}
	keyboardHookQuit      chan struct{}
	mouseHookHandle       syscall.Handle
	keyboardHookHandle    syscall.Handle

	// Walk 控件
	mainWindow                  *walk.MainWindow
	preventCaptureCheckbox      *walk.CheckBox
	topmostCheckbox             *walk.CheckBox
	bottomBroadcastCheckbox     *walk.CheckBox
	blackScreenMinimizeCheckbox *walk.CheckBox
	mouseUnlockCheckbox         *walk.CheckBox
	keyboardUnlockCheckbox      *walk.CheckBox
	broadcastButton             *walk.PushButton
	githubLink                  *walk.LinkLabel
)

// 窗口显示亲和性设置
func SetWindowDisplayAffinity(windowHandle syscall.Handle, affinity uint32) bool {
	ret, _, _ := procSetWindowDisplayAffinity.Call(
		uintptr(windowHandle),
		uintptr(affinity),
	)
	return ret != 0
}

// 窗口置顶状态设置
func SetWindowTopmost(windowHandle syscall.Handle, isTopmost bool) bool {
	insertAfterHandle := uintptr(WindowHandleNotTopMost)
	if isTopmost {
		insertAfterHandle = WindowHandleTopMost
	}

	ret, _, _ := procSetWindowPos.Call(
		uintptr(windowHandle),
		insertAfterHandle,
		0, 0, 0, 0,
		0x0001|0x0002, // SWP_NOSIZE | SWP_NOMOVE
	)
	return ret != 0
}

// 窗口置底设置
func SetWindowBottom(windowHandle syscall.Handle) bool {
	ret, _, _ := procSetWindowPos.Call(
		uintptr(windowHandle),
		WindowHandleBottom,
		0, 0, 0, 0,
		0x0001|0x0002, // SWP_NOSIZE | SWP_NOMOVE
	)
	return ret != 0
}

// 将窗口置于前台
func BringWindowToFront(windowHandle syscall.Handle) bool {
	ret, _, _ := procBringWindowToTop.Call(uintptr(windowHandle))
	return ret != 0
}

// 通过标题查找窗口
func FindWindowByTitle(className, windowTitle string) syscall.Handle {
	var classPtr, titlePtr uintptr
	if className != "" {
		classNamePtr, _ := syscall.UTF16PtrFromString(className)
		classPtr = uintptr(unsafe.Pointer(classNamePtr))
	}
	if windowTitle != "" {
		windowTitlePtr, _ := syscall.UTF16PtrFromString(windowTitle)
		titlePtr = uintptr(unsafe.Pointer(windowTitlePtr))
	}

	ret, _, _ := procFindWindow.Call(classPtr, titlePtr)
	return syscall.Handle(ret)
}

// 发送窗口消息
func SendWindowMessage(windowHandle syscall.Handle, msg uint32, wParam, lParam uintptr) bool {
	ret, _, _ := procPostMessage.Call(
		uintptr(windowHandle),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret != 0
}

// 设置窗口无激活属性
func SetWindowNoActivate(windowHandle syscall.Handle, noActivate bool) bool {
	style, _, _ := procGetWindowLong.Call(
		uintptr(windowHandle),
		uintptr(GWL_EXSTYLE),
	)

	if noActivate {
		style |= WS_EX_NOACTIVATE
	} else {
		style &^= WS_EX_NOACTIVATE
	}

	ret, _, _ := procSetWindowLong.Call(
		uintptr(windowHandle),
		uintptr(GWL_EXSTYLE),
		style,
	)
	return ret != 0
}

// 设置窗口父子关系
func SetParent(childWindow, parentWindow syscall.Handle) bool {
	ret, _, _ := procSetParent.Call(
		uintptr(childWindow),
		uintptr(parentWindow),
	)
	return ret != 0
}

// 显示窗口
func ShowWindow(windowHandle syscall.Handle, cmdShow int) bool {
	ret, _, _ := procShowWindow.Call(
		uintptr(windowHandle),
		uintptr(cmdShow),
	)
	return ret != 0
}

// 移动窗口
func MoveWindow(windowHandle syscall.Handle, x, y, width, height int, repaint bool) bool {
	repaintFlag := 0
	if repaint {
		repaintFlag = 1
	}
	ret, _, _ := procMoveWindow.Call(
		uintptr(windowHandle),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(repaintFlag),
	)
	return ret != 0
}

// 创建黑屏父窗口
func CreateBlackScreenParentWindow() syscall.Handle {
	const (
		WS_OVERLAPPEDWINDOW = 0x00CF0000
		WS_MINIMIZE         = 0x20000000
		WS_EX_TOOLWINDOW    = 0x00000080
	)

	windowNamePtr, _ := syscall.UTF16PtrFromString("黑屏劫持")
	classNamePtr, _ := syscall.UTF16PtrFromString("Static")

	hwnd, _, _ := procCreateWindowEx.Call(
		uintptr(WS_EX_TOOLWINDOW),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windowNamePtr)),
		uintptr(WS_OVERLAPPEDWINDOW|WS_MINIMIZE),
		0, 0, 100, 100,
		0, 0, 0, 0,
	)

	if hwnd != 0 {
		MoveWindow(syscall.Handle(hwnd), -10000, -10000, 100, 100, false)
	}

	return syscall.Handle(hwnd)
}

// 销毁窗口
func DestroyWindow(windowHandle syscall.Handle) bool {
	ret, _, _ := procDestroyWindow.Call(uintptr(windowHandle))
	return ret != 0
}

// 广播窗口钩子过程
func broadcastWindowHook(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == WM_WINDOWPOSCHANGING {
		SetWindowBottom(syscall.Handle(hwnd))
		return 0
	}

	if msg == WM_MOUSEACTIVATE {
		return MA_NOACTIVATE
	}

	if msg == WM_NCACTIVATE || msg == WM_ACTIVATE {
		if wParam != 0 {
			return 0
		}
	}

	ret, _, _ := procCallWindowProc.Call(
		originalBroadcastWindowProc,
		hwnd,
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret
}

// 设置广播窗口钩子
func SetBroadcastWindowHook(windowHandle syscall.Handle) bool {
	originalProc, _, _ := procGetWindowLong.Call(
		uintptr(windowHandle),
		uintptr(GWL_WNDPROC),
	)
	originalBroadcastWindowProc = originalProc

	ret, _, _ := procSetWindowLong.Call(
		uintptr(windowHandle),
		uintptr(GWL_WNDPROC),
		syscall.NewCallback(broadcastWindowHook),
	)
	return ret != 0
}

// Windows事件钩子
func winEventHook(hWinEventHook syscall.Handle, event uint32, hwnd syscall.Handle, idObject, idChild int32, dwEventThread, dwmsEventTime uint32) uintptr {
	if broadcastHookEnabled && hwnd == broadcastWindowHandle {
		SetWindowBottom(hwnd)
		SetWindowNoActivate(hwnd, true)
	}
	return 0
}

// 设置广播事件钩子
func SetBroadcastEventHook() bool {
	const (
		EVENT_OBJECT_LOCATIONCHANGE = 0x800B
		WINEVENT_OUTOFCONTEXT       = 0x0000
	)

	ret, _, _ := procSetWinEventHook.Call(
		uintptr(EVENT_OBJECT_LOCATIONCHANGE),
		uintptr(EVENT_OBJECT_LOCATIONCHANGE),
		0,
		syscall.NewCallback(winEventHook),
		0,
		0,
		WINEVENT_OUTOFCONTEXT,
	)
	eventHook = syscall.Handle(ret)
	return eventHook != 0
}

// 移除广播事件钩子
func RemoveBroadcastEventHook() {
	if eventHook != 0 {
		procUnhookWinEvent.Call(uintptr(eventHook))
		eventHook = 0
	}
}

// 切换广播窗口状态
func ToggleBroadcastWindow(owner *walk.MainWindow) {
	broadcastWindowHandle = FindWindowByTitle("", "屏幕广播")
	if broadcastWindowHandle == 0 {
		dialog := walk.NewTaskDialog()
		_, err := dialog.Show(walk.TaskDialogOpts{
			Owner:         owner,
			Title:         "画栋朝飞南浦云, 珠帘暮卷西山雨.",
			Instruction:   "未找到屏幕广播窗口",
			IconSystem:    walk.TaskDialogSystemIconInformation,
			CommonButtons: win.TDCBF_OK_BUTTON,
		})
		if err != nil {
			log.Println(err)
		}
		return
	}

	wParam := uintptr((ButtonClick << 16) | 1004)
	SendWindowMessage(broadcastWindowHandle, MessageCommand, wParam, 0)
}

// 钩子过程函数
func HookProc(nCode, wParam, lParam uintptr) uintptr {
	return 0
}

// 鼠标解锁线程
func MouseHookThread() uintptr {
	for {
		select {
		case <-mouseHookQuit:
			return 0
		default:
			// 设置低级鼠标钩子
			hook, _, _ := procSetWindowsHookEx.Call(
				WH_MOUSE_LL,
				syscall.NewCallback(HookProc),
				0,
				0,
			)
			if hook != 0 {
				mouseHookHandle = syscall.Handle(hook)
			}

			// 解除鼠标锁定
			procClipCursor.Call(0)

			time.Sleep(25 * time.Millisecond)

			// 卸载钩子
			if mouseHookHandle != 0 {
				procUnhookWindowsHookEx.Call(uintptr(mouseHookHandle))
				mouseHookHandle = 0
			}
		}
	}
}

// 键盘解锁线程
func KeyboardHookThread() uintptr {
	for {
		select {
		case <-keyboardHookQuit:
			return 0
		default:
			// 设置多个键盘钩子确保覆盖
			hook1, _, _ := procSetWindowsHookEx.Call(
				WH_KEYBOARD_LL,
				syscall.NewCallback(HookProc),
				0,
				0,
			)
			hook2, _, _ := procSetWindowsHookEx.Call(
				WH_KEYBOARD_LL,
				syscall.NewCallback(HookProc),
				0,
				0,
			)
			hook3, _, _ := procSetWindowsHookEx.Call(
				WH_KEYBOARD,
				syscall.NewCallback(HookProc),
				0,
				0,
			)
			hook4, _, _ := procSetWindowsHookEx.Call(
				WH_KEYBOARD,
				syscall.NewCallback(HookProc),
				0,
				0,
			)

			if hook1 != 0 {
				keyboardHookHandle = syscall.Handle(hook1)
			}

			// 尝试解除驱动级键盘锁
			deviceNamePtr, _ := syscall.UTF16PtrFromString("\\\\.\\TDKeybd")
			device, _, _ := procCreateFile.Call(
				uintptr(unsafe.Pointer(deviceNamePtr)),
				uintptr(syscall.GENERIC_READ|syscall.GENERIC_WRITE),
				uintptr(syscall.FILE_SHARE_READ),
				0,
				uintptr(syscall.OPEN_EXISTING),
				0,
				0,
			)

			if device != 0 && device != ^uintptr(0) {
				var enable uint32 = 1
				var bytesReturned uint32
				procDeviceIoControl.Call(
					device,
					0x220000,
					uintptr(unsafe.Pointer(&enable)),
					4,
					0,
					0,
					uintptr(unsafe.Pointer(&bytesReturned)),
					0,
				)
				procCloseHandle.Call(device)
			}

			time.Sleep(30 * time.Millisecond)

			// 卸载所有钩子
			if hook1 != 0 {
				procUnhookWindowsHookEx.Call(hook1)
			}
			if hook2 != 0 {
				procUnhookWindowsHookEx.Call(hook2)
			}
			if hook3 != 0 {
				procUnhookWindowsHookEx.Call(hook3)
			}
			if hook4 != 0 {
				procUnhookWindowsHookEx.Call(hook4)
			}
			keyboardHookHandle = 0
		}
	}
}

// 启动鼠标解锁
func StartMouseUnlock() {
	mouseHookQuit = make(chan struct{})
	thread, _, _ := procCreateThread.Call(
		0,
		0,
		syscall.NewCallback(MouseHookThread),
		0,
		0,
		0,
	)
	mouseHookThread = syscall.Handle(thread)
}

// 停止鼠标解锁
func StopMouseUnlock() {
	if mouseHookQuit != nil {
		close(mouseHookQuit)
		mouseHookQuit = nil
	}
	if mouseHookHandle != 0 {
		procUnhookWindowsHookEx.Call(uintptr(mouseHookHandle))
		mouseHookHandle = 0
	}
	if mouseHookThread != 0 {
		procTerminateThread.Call(uintptr(mouseHookThread), 0)
		mouseHookThread = 0
	}
}

// 启动键盘解锁
func StartKeyboardUnlock() {
	keyboardHookQuit = make(chan struct{})
	thread, _, _ := procCreateThread.Call(
		0,
		0,
		syscall.NewCallback(KeyboardHookThread),
		0,
		0,
		0,
	)
	keyboardHookThread = syscall.Handle(thread)
}

// 停止键盘解锁
func StopKeyboardUnlock() {
	if keyboardHookQuit != nil {
		close(keyboardHookQuit)
		keyboardHookQuit = nil
	}
	if keyboardHookHandle != 0 {
		procUnhookWindowsHookEx.Call(uintptr(keyboardHookHandle))
		keyboardHookHandle = 0
	}
	if keyboardHookThread != 0 {
		procTerminateThread.Call(uintptr(keyboardHookThread), 0)
		keyboardHookThread = 0
	}
}

// 打开URL链接
func OpenURL(url string) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procShellExecute := shell32.NewProc("ShellExecuteW")
	procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(url))),
		0,
		0,
		uintptr(win.SW_SHOWNORMAL),
	)
}

func main() {
	application, err := walk.InitApp()
	if err != nil {
		log.Fatal(err)
	}

	// 创建退出通道
	quitChannel := make(chan struct{})
	broadcastQuitChannel := make(chan struct{})
	blackScreenQuitChannel = make(chan struct{})

	// 创建主窗口
	err = MainWindow{
		AssignTo: &mainWindow,
		Title:    "Mythgone",
		Icon: func() *walk.Icon {
			icon, err := walk.NewIconFromResourceId(2)
			if err != nil {
				log.Fatal(err)
			}
			return icon
		}(),
		Size:            Size{Width: 500, Height: 300},
		DisableMaximize: true,
		DisableResizing: true,
		Layout:          VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}, Spacing: 15},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2, Spacing: 10},
				Children: []Widget{
					GroupBox{
						Title:  "窗口控制",
						Layout: VBox{Margins: Margins{Left: 10, Top: 12, Right: 10, Bottom: 12}, Spacing: 12},
						Children: []Widget{
							CheckBox{
								AssignTo: &preventCaptureCheckbox,
								Text:     "禁止捕获",
								Checked:  true,
								OnCheckedChanged: func() {
									if mainWindow == nil {
										return
									}
									if preventCaptureCheckbox.Checked() {
										SetWindowDisplayAffinity(syscall.Handle(mainWindow.Handle()), WindowDisplayAffinityExcludeFromCapture)
									} else {
										SetWindowDisplayAffinity(syscall.Handle(mainWindow.Handle()), WindowDisplayAffinityNone)
									}
								},
							},
							CheckBox{
								AssignTo: &topmostCheckbox,
								Text:     "窗口置顶",
								Checked:  true,
								OnCheckedChanged: func() {
									if mainWindow != nil {
										if topmostCheckbox.Checked() {
											SetWindowTopmost(syscall.Handle(mainWindow.Handle()), true)
										} else {
											SetWindowTopmost(syscall.Handle(mainWindow.Handle()), false)
											BringWindowToFront(syscall.Handle(mainWindow.Handle()))
											mainWindow.Activate()
										}
									}
								},
							},
						},
					},
					GroupBox{
						Title:  "广播控制",
						Layout: VBox{Margins: Margins{Left: 10, Top: 12, Right: 10, Bottom: 12}, Spacing: 12},
						Children: []Widget{
							CheckBox{
								AssignTo: &bottomBroadcastCheckbox,
								Text:     "广播置底",
								Checked:  false,
								OnCheckedChanged: func() {
									if bottomBroadcastCheckbox.Checked() {
										broadcastHookEnabled = true
										if SetBroadcastEventHook() {
											go monitorBroadcastWindow(broadcastQuitChannel)
										}
									} else {
										broadcastHookEnabled = false
										RemoveBroadcastEventHook()
										close(broadcastQuitChannel)
										broadcastQuitChannel = make(chan struct{})
									}
								},
							},
							PushButton{
								AssignTo: &broadcastButton,
								Text:     "切换窗口模式",
								MinSize:  Size{Width: 40, Height: 32},
								OnClicked: func() {
									ToggleBroadcastWindow(mainWindow)
								},
							},
						},
					},
					GroupBox{
						Title:  "黑屏控制",
						Layout: VBox{Margins: Margins{Left: 10, Top: 12, Right: 10, Bottom: 12}},
						Children: []Widget{
							CheckBox{
								AssignTo: &blackScreenMinimizeCheckbox,
								Text:     "隐藏黑屏",
								Checked:  false,
								OnCheckedChanged: func() {
									if blackScreenMinimizeCheckbox.Checked() {
										blackScreenMinimizeEnabled = true
										go monitorBlackScreenWindow(blackScreenQuitChannel)
									} else {
										blackScreenMinimizeEnabled = false
										if embeddedBlackScreenWindow != 0 {
											SetParent(embeddedBlackScreenWindow, 0)
											embeddedBlackScreenWindow = 0
										}
										if blackScreenParentWindow != 0 {
											DestroyWindow(blackScreenParentWindow)
											blackScreenParentWindow = 0
										}
										close(blackScreenQuitChannel)
										blackScreenQuitChannel = make(chan struct{})
									}
								},
							},
						},
					},
					GroupBox{
						Title:  "解锁控制 (实验)",
						Layout: VBox{Margins: Margins{Left: 10, Top: 12, Right: 10, Bottom: 12}, Spacing: 12},
						Children: []Widget{
							CheckBox{
								AssignTo: &mouseUnlockCheckbox,
								Text:     "解除鼠标锁",
								Checked:  false,
								OnCheckedChanged: func() {
									if mouseUnlockCheckbox.Checked() {
										mouseUnlockEnabled = true
										StartMouseUnlock()
									} else {
										mouseUnlockEnabled = false
										StopMouseUnlock()
									}
								},
							},
							CheckBox{
								AssignTo: &keyboardUnlockCheckbox,
								Text:     "解除键盘锁",
								Checked:  false,
								OnCheckedChanged: func() {
									if keyboardUnlockCheckbox.Checked() {
										keyboardUnlockEnabled = true
										StartKeyboardUnlock()
									} else {
										keyboardUnlockEnabled = false
										StopKeyboardUnlock()
									}
								},
							},
						},
					},
				},
			},

			VSpacer{},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					LinkLabel{
						AssignTo: &githubLink,
						Text:     `由 dotcubecn 与所有贡献者开发 (<a href="https://github.com/dotcubecn/mythgone">GitHub</a>)`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							OpenURL(link.URL())
						},
					},
					HSpacer{},
				},
			},
		},
	}.Create()
	if err != nil {
		log.Fatal(err)
	}

	// 初始设置
	if preventCaptureCheckbox.Checked() {
		SetWindowDisplayAffinity(syscall.Handle(mainWindow.Handle()), WindowDisplayAffinityExcludeFromCapture)
	}

	SetWindowTopmost(syscall.Handle(mainWindow.Handle()), topmostCheckbox.Checked())

	// 置顶循环
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if topmostCheckbox != nil && topmostCheckbox.Checked() && mainWindow != nil {
					SetWindowTopmost(syscall.Handle(mainWindow.Handle()), true)
				}
			case <-quitChannel:
				return
			}
		}
	}()

	// 窗口关闭事件处理
	mainWindow.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		broadcastHookEnabled = false
		blackScreenMinimizeEnabled = false
		mouseUnlockEnabled = false
		keyboardUnlockEnabled = false
		RemoveBroadcastEventHook()
		StopMouseUnlock()
		StopKeyboardUnlock()
		if embeddedBlackScreenWindow != 0 {
			SetParent(embeddedBlackScreenWindow, 0)
		}
		if blackScreenParentWindow != 0 {
			DestroyWindow(blackScreenParentWindow)
		}
		close(quitChannel)
		close(broadcastQuitChannel)
		close(blackScreenQuitChannel)
		syscall.Exit(0)
	})

	application.Run()
}

// 监控广播窗口
func monitorBroadcastWindow(quit chan struct{}) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	var hookedWindow syscall.Handle

	for {
		select {
		case <-ticker.C:
			broadcastWindow := FindWindowByTitle("", "屏幕广播")
			if broadcastWindow != 0 {
				broadcastWindowHandle = broadcastWindow
				if hookedWindow == 0 {
					if SetBroadcastWindowHook(broadcastWindow) {
						hookedWindow = broadcastWindow
					}
				}
				if broadcastHookEnabled {
					SetWindowBottom(broadcastWindow)
					SetWindowNoActivate(broadcastWindow, true)
				}
			} else {
				hookedWindow = 0
				broadcastWindowHandle = 0
			}
		case <-quit:
			return
		}
	}
}

// 监控黑屏窗口
func monitorBlackScreenWindow(quit chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if blackScreenMinimizeEnabled {
				blackScreenWindow := FindWindowByTitle("", "BlackScreen Window")
				if blackScreenWindow != 0 && embeddedBlackScreenWindow == 0 {
					blackScreenParentWindow = CreateBlackScreenParentWindow()
					if blackScreenParentWindow != 0 {
						if SetParent(blackScreenWindow, blackScreenParentWindow) {
							embeddedBlackScreenWindow = blackScreenWindow
							ShowWindow(blackScreenParentWindow, 6) // SW_MINIMIZE
						}
					}
				} else if blackScreenWindow == 0 && embeddedBlackScreenWindow != 0 {
					embeddedBlackScreenWindow = 0
					if blackScreenParentWindow != 0 {
						DestroyWindow(blackScreenParentWindow)
						blackScreenParentWindow = 0
					}
				}
			}
		case <-quit:
			return
		}
	}
}
