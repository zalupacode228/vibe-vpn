package main

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	shell32                 = syscall.NewLazyDLL("shell32.dll")
	procIsUserAnAdmin       = shell32.NewProc("IsUserAnAdmin")
	procShellExecuteW       = shell32.NewProc("ShellExecuteW")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW        = kernel32.NewProc("CreateMutexW")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	user32                  = syscall.NewLazyDLL("user32.dll")
	procFindWindowW         = user32.NewProc("FindWindowW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")

	appMutexHandle uintptr
)

func isAdmin() bool {
	ret, _, _ := procIsUserAnAdmin.Call()
	return ret != 0
}

func activateExistingWindow() {
	wTitle, _ := syscall.UTF16PtrFromString("VPN")
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(wTitle)))
	if hwnd != 0 {
		procShowWindow.Call(hwnd, 9) // SW_RESTORE
		procSetForegroundWindow.Call(hwnd)
	}
}

func releaseSingleInstanceMutex() {
	if appMutexHandle != 0 {
		procCloseHandle.Call(appMutexHandle)
		appMutexHandle = 0
	}
}

func checkSingleInstanceMutex() bool {
	name, _ := syscall.UTF16PtrFromString("Local\\VPN_App_Single_Instance_Mutex_9a8b")
	handle, _, err := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(name)))
	const ERROR_ALREADY_EXISTS = 183
	if err == syscall.Errno(ERROR_ALREADY_EXISTS) {
		activateExistingWindow()
		return false
	}
	appMutexHandle = handle
	return true
}

func elevateProcess() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)

	// Keep existing flags
	var filtered []string
	for _, a := range os.Args[1:] {
		filtered = append(filtered, `"`+a+`"`)
	}
	args := strings.Join(filtered, " ")
	params, _ := syscall.UTF16PtrFromString(args)
	cwd, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))

	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		uintptr(unsafe.Pointer(cwd)),
		1, // SW_SHOWNORMAL
	)
	return ret > 32
}

func main() {
	if !checkSingleInstanceMutex() {
		os.Exit(0)
	}

	noElevate := false
	for _, a := range os.Args {
		if a == "--no-elevate" {
			noElevate = true
			break
		}
	}

	if !isAdmin() && !noElevate {
		if elevateProcess() {
			os.Exit(0)
		}
	}

	app := NewApp()

	wvDataPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "VPN", "webview2")
	if os.Getenv("LOCALAPPDATA") == "" {
		wvDataPath = filepath.Join(os.Getenv("APPDATA"), "VPN", "webview2")
	}

	err := wails.Run(&options.App{
		Title:            "VPN",
		Width:            1000,
		Height:           700,
		MinWidth:         380,
		MinHeight:        600,
		Frameless:         true,
		HideWindowOnClose: true,
		BackgroundColour:  &options.RGBA{R: 20, G: 18, B: 24, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnDomReady: app.domReady,
		OnShutdown: app.shutdown,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "e9f0c23a-7a54-4a57-b087-vpn-client-single",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				if app.ctx != nil {
					runtime.WindowShow(app.ctx)
					runtime.WindowUnminimise(app.ctx)
					runtime.WindowSetAlwaysOnTop(app.ctx, true)
					runtime.WindowSetAlwaysOnTop(app.ctx, false)
				}
			},
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			BackdropType:         windows.None,
			Theme:                windows.Dark,
			WebviewUserDataPath:  wvDataPath,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
