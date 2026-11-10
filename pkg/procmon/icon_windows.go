package procmon

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"sync"
	"syscall"
	"unsafe"
)

var (
	modShell32         = syscall.NewLazyDLL("shell32.dll")
	modUser32          = syscall.NewLazyDLL("user32.dll")
	modGdi32           = syscall.NewLazyDLL("gdi32.dll")
	procExtractIconExW = modShell32.NewProc("ExtractIconExW")
	procSHGetFileInfoW = modShell32.NewProc("SHGetFileInfoW")
	procGetIconInfo    = modUser32.NewProc("GetIconInfo")
	procDestroyIcon    = modUser32.NewProc("DestroyIcon")
	procGetDC          = modUser32.NewProc("GetDC")
	procReleaseDC      = modUser32.NewProc("ReleaseDC")
	procGetObjectW     = modGdi32.NewProc("GetObjectW")
	procDeleteObject   = modGdi32.NewProc("DeleteObject")
	procGetDIBits      = modGdi32.NewProc("GetDIBits")

	iconCache   = make(map[string]string)
	iconCacheMu sync.RWMutex
)

type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type BITMAP struct {
	BmType       int32
	BmWidth      int32
	BmHeight     int32
	BmWidthBytes int32
	BmPlanes     uint16
	BmBitsPixel  uint16
	BmBits       uintptr
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type SHFILEINFOW struct {
	HIcon         uintptr
	IIcon         int32
	DwAttributes  uint32
	SzDisplayName [260]uint16
	SzTypeName    [80]uint16
}

const (
	SHGFI_ICON      = 0x000000100
	SHGFI_LARGEICON = 0x000000000
)

// GetAppIcon returns a base64 Data URI for the given exe path, or empty string if not found.
func GetAppIcon(exePath string) string {
	if exePath == "" {
		return ""
	}

	iconCacheMu.RLock()
	cached, ok := iconCache[exePath]
	iconCacheMu.RUnlock()
	if ok {
		return cached
	}

	iconB64 := extractIcon(exePath)

	iconCacheMu.Lock()
	iconCache[exePath] = iconB64
	iconCacheMu.Unlock()

	return iconB64
}

func extractIcon(exePath string) string {
	p, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return ""
	}

	var hLarge, hSmall uintptr
	ret, _, _ := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(p)),
		0,
		uintptr(unsafe.Pointer(&hLarge)),
		uintptr(unsafe.Pointer(&hSmall)),
		1,
	)

	hIcon := hLarge
	if hIcon == 0 {
		hIcon = hSmall
	}

	if ret == 0 || hIcon == 0 {
		// Fallback to SHGetFileInfoW
		var sfi SHFILEINFOW
		r, _, _ := procSHGetFileInfoW.Call(
			uintptr(unsafe.Pointer(p)),
			0,
			uintptr(unsafe.Pointer(&sfi)),
			uintptr(unsafe.Sizeof(sfi)),
			SHGFI_ICON|SHGFI_LARGEICON,
		)
		if r != 0 && sfi.HIcon != 0 {
			hIcon = sfi.HIcon
		}
	}

	if hIcon == 0 {
		return ""
	}
	defer func() {
		procDestroyIcon.Call(hIcon)
		if hSmall != 0 && hSmall != hIcon {
			procDestroyIcon.Call(hSmall)
		}
		if hLarge != 0 && hLarge != hIcon {
			procDestroyIcon.Call(hLarge)
		}
	}()

	var iconInfo ICONINFO
	r, _, _ := procGetIconInfo.Call(hIcon, uintptr(unsafe.Pointer(&iconInfo)))
	if r == 0 {
		return ""
	}
	defer func() {
		if iconInfo.HbmMask != 0 {
			procDeleteObject.Call(iconInfo.HbmMask)
		}
		if iconInfo.HbmColor != 0 {
			procDeleteObject.Call(iconInfo.HbmColor)
		}
	}()

	var bm BITMAP
	r, _, _ = procGetObjectW.Call(iconInfo.HbmColor, uintptr(unsafe.Sizeof(bm)), uintptr(unsafe.Pointer(&bm)))
	if r == 0 {
		return ""
	}

	width := int(bm.BmWidth)
	height := int(bm.BmHeight)
	if width <= 0 || height <= 0 || width > 256 || height > 256 {
		return ""
	}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return ""
	}
	defer procReleaseDC.Call(0, hdc)

	var bmi BITMAPINFOHEADER
	bmi.BiSize = uint32(unsafe.Sizeof(bmi))
	bmi.BiWidth = int32(width)
	bmi.BiHeight = -int32(height) // top-down
	bmi.BiPlanes = 1
	bmi.BiBitCount = 32
	bmi.BiCompression = 0 // BI_RGB

	bufSize := width * height * 4
	rawPixels := make([]byte, bufSize)

	r, _, _ = procGetDIBits.Call(
		hdc,
		iconInfo.HbmColor,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&rawPixels[0])),
		uintptr(unsafe.Pointer(&bmi)),
		0, // DIB_RGB_COLORS
	)
	if r == 0 {
		return ""
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	hasAlpha := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			b := rawPixels[idx]
			g := rawPixels[idx+1]
			r := rawPixels[idx+2]
			a := rawPixels[idx+3]
			if a > 0 {
				hasAlpha = true
			}
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	if !hasAlpha {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				c := img.RGBAAt(x, y)
				c.A = 255
				img.SetRGBA(x, y, c)
			}
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return ""
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes())
}
