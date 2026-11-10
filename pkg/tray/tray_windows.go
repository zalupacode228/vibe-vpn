package tray

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	moduser32   = syscall.NewLazyDLL("user32.dll")
	modshell32  = syscall.NewLazyDLL("shell32.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW = moduser32.NewProc("RegisterClassExW")
	procCreateWindowExW  = moduser32.NewProc("CreateWindowExW")
	procDefWindowProcW   = moduser32.NewProc("DefWindowProcW")
	procDestroyWindow    = moduser32.NewProc("DestroyWindow")
	procGetMessageW      = moduser32.NewProc("GetMessageW")
	procTranslateMessage = moduser32.NewProc("TranslateMessage")
	procDispatchMessageW = moduser32.NewProc("DispatchMessageW")
	procPostQuitMessage  = moduser32.NewProc("PostQuitMessage")
	procSetForeground    = moduser32.NewProc("SetForegroundWindow")
	procGetCursorPos     = moduser32.NewProc("GetCursorPos")
	procCreatePopupMenu  = moduser32.NewProc("CreatePopupMenu")
	procAppendMenuW      = moduser32.NewProc("AppendMenuW")
	procTrackPopupMenu   = moduser32.NewProc("TrackPopupMenu")
	procDestroyMenu      = moduser32.NewProc("DestroyMenu")
	procLoadIconW        = moduser32.NewProc("LoadIconW")

	procShell_NotifyIconW = modshell32.NewProc("Shell_NotifyIconW")
	procExtractIconW      = modshell32.NewProc("ExtractIconW")
	procGetModuleHandleW  = modkernel32.NewProc("GetModuleHandleW")
)

const (
	WM_USER          = 0x0400
	WM_TRAYICON      = WM_USER + 100
	WM_COMMAND       = 0x0111
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_DESTROY       = 0x0002

	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	TPM_RIGHTBUTTON = 0x0002

	ID_SHOW = 1001
	ID_EXIT = 1002
)

type NOTIFYICONDATAW struct {
	cbSize            uint32
	hWnd              uintptr
	uID               uint32
	uFlags            uint32
	uCallbackMessage  uint32
	hIcon             uintptr
	szTip             [128]uint16
	dwState           uint32
	dwStateMask       uint32
	szInfo            [256]uint16
	uTimeoutOrVersion uint32
	szInfoTitle       [64]uint16
	dwInfoFlags       uint32
	guidItem          syscall.GUID
	hBalloonIcon      uintptr
}

type WNDCLASSEXW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type Tray struct {
	hWnd uintptr
	nid  NOTIFYICONDATAW
}

var (
	onShowCallback func()
	onExitCallback func()
)

func wndProc(hWnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
			if onShowCallback != nil {
				onShowCallback()
			}
		case WM_RBUTTONUP:
			var pt POINT
			procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
			hMenu, _, _ := procCreatePopupMenu.Call()

			showStr, _ := syscall.UTF16PtrFromString("Открыть VPN")
			exitStr, _ := syscall.UTF16PtrFromString("Выход")

			procAppendMenuW.Call(hMenu, MF_STRING, ID_SHOW, uintptr(unsafe.Pointer(showStr)))
			procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
			procAppendMenuW.Call(hMenu, MF_STRING, ID_EXIT, uintptr(unsafe.Pointer(exitStr)))

			procSetForeground.Call(hWnd)
			procTrackPopupMenu.Call(hMenu, TPM_RIGHTBUTTON, uintptr(pt.X), uintptr(pt.Y), 0, hWnd, 0)
			procDestroyMenu.Call(hMenu)
		}
		return 0
	case WM_COMMAND:
		cmdId := wParam & 0xFFFF
		if cmdId == ID_SHOW {
			if onShowCallback != nil {
				onShowCallback()
			}
		} else if cmdId == ID_EXIT {
			if onExitCallback != nil {
				onExitCallback()
			}
		}
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hWnd, uintptr(msg), wParam, lParam)
	return ret
}

func InitTray(tooltip string, onShow func(), onExit func()) (*Tray, error) {
	onShowCallback = onShow
	onExitCallback = onExit

	t := &Tray{}
	readyChan := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hInst, _, _ := procGetModuleHandleW.Call(0)
		className, _ := syscall.UTF16PtrFromString("VPNTrayWindowClass")

		wc := WNDCLASSEXW{
			cbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			lpfnWndProc:   syscall.NewCallback(wndProc),
			hInstance:     hInst,
			lpszClassName: className,
		}
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

		hWnd, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(className)),
			0,
			0, 0, 0, 0,
			0, // HWND_MESSAGE or 0
			0,
			hInst,
			0,
		)
		t.hWnd = hWnd

		// Extract app icon
		var hIcon uintptr
		if exePath, err := os.Executable(); err == nil {
			exePtr, _ := syscall.UTF16PtrFromString(exePath)
			res, _, _ := procExtractIconW.Call(hInst, uintptr(unsafe.Pointer(exePtr)), 0)
			if res > 1 {
				hIcon = res
			}
		}
		if hIcon == 0 {
			hIcon, _, _ = procLoadIconW.Call(0, 32512) // IDI_APPLICATION
		}

		t.nid = NOTIFYICONDATAW{
			cbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATAW{})),
			hWnd:             hWnd,
			uID:              1,
			uFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
			uCallbackMessage: WM_TRAYICON,
			hIcon:            hIcon,
		}
		tipUTF16, _ := syscall.UTF16FromString(tooltip)
		copy(t.nid.szTip[:], tipUTF16)

		procShell_NotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&t.nid)))
		close(readyChan)

		var msg MSG
		for {
			ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if ret == 0 || int32(ret) == -1 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()

	<-readyChan
	return t, nil
}

func (t *Tray) Close() {
	if t.hWnd != 0 {
		procShell_NotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&t.nid)))
		procDestroyWindow.Call(t.hWnd)
		t.hWnd = 0
	}
}
