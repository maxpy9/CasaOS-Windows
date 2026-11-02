package main

import (
	"math"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	gdiplus  = windows.NewLazySystemDLL("gdiplus.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	dwmapi   = windows.NewLazySystemDLL("dwmapi.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procIsDialogMessageW     = user32.NewProc("IsDialogMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procPostMessageW         = user32.NewProc("PostMessageW")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadImageW           = user32.NewProc("LoadImageW")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procFillRect             = user32.NewProc("FillRect")
	procDrawTextW            = user32.NewProc("DrawTextW")
	procSetWindowPos         = user32.NewProc("SetWindowPos")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procAdjustWindowRectEx   = user32.NewProc("AdjustWindowRectExForDpi")
	procGetDpiForWindow      = user32.NewProc("GetDpiForWindow")
	procGetDpiForSystem      = user32.NewProc("GetDpiForSystem")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	procFindWindowW          = user32.NewProc("FindWindowW")
	procIsWindowVisible      = user32.NewProc("IsWindowVisible")
	procIsIconic             = user32.NewProc("IsIconic")
	procEnableWindow         = user32.NewProc("EnableWindow")
	procIsWindowEnabled      = user32.NewProc("IsWindowEnabled")
	procSetFocus             = user32.NewProc("SetFocus")
	procGetFocus             = user32.NewProc("GetFocus")
	procSetCapture           = user32.NewProc("SetCapture")
	procReleaseCapture       = user32.NewProc("ReleaseCapture")
	procTrackMouseEvent      = user32.NewProc("TrackMouseEvent")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procCreatePopupMenu      = user32.NewProc("CreatePopupMenu")
	procAppendMenuW          = user32.NewProc("AppendMenuW")
	procTrackPopupMenu       = user32.NewProc("TrackPopupMenu")
	procDestroyMenu          = user32.NewProc("DestroyMenu")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procOpenClipboard        = user32.NewProc("OpenClipboard")
	procCloseClipboard       = user32.NewProc("CloseClipboard")
	procEmptyClipboard       = user32.NewProc("EmptyClipboard")
	procSetClipboardData     = user32.NewProc("SetClipboardData")
	procSetTimer             = user32.NewProc("SetTimer")
	procKillTimer            = user32.NewProc("KillTimer")
	procPrintWindow          = user32.NewProc("PrintWindow")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procSetProcessDpiContext = user32.NewProc("SetProcessDpiAwarenessContext")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procCreateFontW            = gdi32.NewProc("CreateFontW")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procGetDIBits              = gdi32.NewProc("GetDIBits")

	procGdiplusStartup             = gdiplus.NewProc("GdiplusStartup")
	procGdipCreateFromHDC          = gdiplus.NewProc("GdipCreateFromHDC")
	procGdipDeleteGraphics         = gdiplus.NewProc("GdipDeleteGraphics")
	procGdipSetSmoothingMode       = gdiplus.NewProc("GdipSetSmoothingMode")
	procGdipSetInterpolationMode   = gdiplus.NewProc("GdipSetInterpolationMode")
	procGdipSetPixelOffsetMode     = gdiplus.NewProc("GdipSetPixelOffsetMode")
	procGdipCreateSolidFill        = gdiplus.NewProc("GdipCreateSolidFill")
	procGdipDeleteBrush            = gdiplus.NewProc("GdipDeleteBrush")
	procGdipCreatePen1             = gdiplus.NewProc("GdipCreatePen1")
	procGdipDeletePen              = gdiplus.NewProc("GdipDeletePen")
	procGdipCreatePath             = gdiplus.NewProc("GdipCreatePath")
	procGdipDeletePath             = gdiplus.NewProc("GdipDeletePath")
	procGdipAddPathArcI            = gdiplus.NewProc("GdipAddPathArcI")
	procGdipAddPathLineI           = gdiplus.NewProc("GdipAddPathLineI")
	procGdipClosePathFigure        = gdiplus.NewProc("GdipClosePathFigure")
	procGdipFillPath               = gdiplus.NewProc("GdipFillPath")
	procGdipDrawPath               = gdiplus.NewProc("GdipDrawPath")
	procGdipFillEllipseI           = gdiplus.NewProc("GdipFillEllipseI")
	procGdipCreateBitmapFromScan0  = gdiplus.NewProc("GdipCreateBitmapFromScan0")
	procGdipDisposeImage           = gdiplus.NewProc("GdipDisposeImage")
	procGdipDrawImageRectI         = gdiplus.NewProc("GdipDrawImageRectI")
	procShellNotifyIconW           = shell32.NewProc("Shell_NotifyIconW")
	procShellExecuteW              = shell32.NewProc("ShellExecuteW")
	procDwmSetWindowAttribute      = dwmapi.NewProc("DwmSetWindowAttribute")
	procGlobalAlloc                = kernel32.NewProc("GlobalAlloc")
	procGlobalLock                 = kernel32.NewProc("GlobalLock")
	procGlobalUnlock               = kernel32.NewProc("GlobalUnlock")
	procCreateMutexW               = kernel32.NewProc("CreateMutexW")
)

const (
	wsOverlapped  = 0x00000000
	wsCaption     = 0x00C00000
	wsSysMenu     = 0x00080000
	wsMinimizeBox = 0x00020000
	wsChild       = 0x40000000
	wsVisible     = 0x10000000
	wsTabStop     = 0x00010000
	wsClipChild   = 0x02000000

	wmDestroy         = 0x0002
	wmSetFocus        = 0x0007
	wmKillFocus       = 0x0008
	wmPaint           = 0x000F
	wmClose           = 0x0010
	wmQueryEndSession = 0x0011
	wmEraseBkgnd      = 0x0014
	wmEndSession      = 0x0016
	wmCommand         = 0x0111
	wmTimer           = 0x0113
	wmMouseMove       = 0x0200
	wmLButtonDown     = 0x0201
	wmLButtonUp       = 0x0202
	wmRButtonUp       = 0x0205
	wmLButtonDblClk   = 0x0203
	wmMouseLeave      = 0x02A3
	wmKeyDown         = 0x0100
	wmKeyUp           = 0x0101
	wmGetDlgCode      = 0x0087
	wmEnable          = 0x000A
	wmDpiChanged      = 0x02E0
	wmSysCommand      = 0x0112
	wmApp             = 0x8000
	wmNull            = 0x0000

	dlgcButton        = 0x2000
	dlgcUndefPush     = 0x0020
	dlgcDefPush       = 0x0010
	vkReturn          = 0x0D
	vkSpace           = 0x20
	swHide            = 0
	swShow            = 5
	swRestore         = 9
	swShowNormal      = 1
	cwUseDefault      = 0x80000000
	idcArrow          = 32512
	idcHand           = 32649
	imageIcon         = 1
	lrDefaultColor    = 0
	smCxScreen        = 0
	smCyScreen        = 1
	smCxSmIcon        = 49
	smCySmIcon        = 50
	smCxIcon          = 11
	smCyIcon          = 12
	srcCopy           = 0x00CC0020
	transparent       = 1
	dtLeft            = 0x00
	dtCenter          = 0x01
	dtRight           = 0x02
	dtVCenter         = 0x04
	dtWordBreak       = 0x10
	dtSingleLine      = 0x20
	dtEndEllipsis     = 0x8000
	dtNoPrefix        = 0x0800
	dtCalcRect        = 0x0400
	tmeLeave          = 0x00000002
	tpmRightButton    = 0x0002
	tpmReturnCmd      = 0x0100
	mfString          = 0x0000
	mfSeparator       = 0x0800
	mfGrayed          = 0x0001
	mfDefault         = 0x1000
	nimAdd            = 0
	nimModify         = 1
	nimDelete         = 2
	nifMessage        = 1
	nifIcon           = 2
	nifTip            = 4
	nifInfo           = 0x10
	niifInfo          = 1
	gmemMoveable      = 2
	cfUnicodeText     = 13
	errorAlreadyExist = 183

	dwmwaDarkMode = 20

	dpiAwarenessPerMonitorV2 = ^uintptr(3) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 is (HANDLE)-4

	smoothingAntiAlias   = 4
	interpolationBicubic = 7
	pixelOffsetHalf      = 4
	pixelFormat32PARGB   = 0xE200B
	unitPixel            = 2
)

type point struct{ X, Y int32 }

type rect struct{ Left, Top, Right, Bottom int32 }

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type paintStruct struct {
	Hdc       uintptr
	Erase     int32
	RcPaint   rect
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}

type trackMouseEvent struct {
	Size      uint32
	Flags     uint32
	Track     uintptr
	HoverTime uint32
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	Size            uint32
	Hwnd            uintptr
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            uintptr
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        guid
	BalloonIcon     uintptr
}

type gdiplusStartupInput struct {
	Version       uint32
	Callback      uintptr
	NoThread      int32
	NoCodecs      int32
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

func utf16(s string) *uint16 {
	p, err := windows.UTF16PtrFromString(s)
	if err != nil {
		p, _ = windows.UTF16PtrFromString("")
	}
	return p
}

func ptr(s string) uintptr { return uintptr(unsafe.Pointer(utf16(s))) }

func loword(v uintptr) int32 { return int32(int16(v & 0xffff)) }
func hiword(v uintptr) int32 { return int32(int16((v >> 16) & 0xffff)) }

func rgb(c uint32) uintptr {
	// 0xRRGGBB to the COLORREF of GDI
	return uintptr(((c >> 16) & 0xff) | (c & 0xff00) | ((c & 0xff) << 16))
}

// -------------------------------------------------- drawing helpers

func startGdiplus() {
	input := gdiplusStartupInput{Version: 1}
	var token uintptr
	procGdiplusStartup.Call(uintptr(unsafe.Pointer(&token)), uintptr(unsafe.Pointer(&input)), 0) // nolint: errcheck
}

// canvas is a double buffer for a window.
type canvas struct {
	target, dc, bitmap, old uintptr
	graphics                uintptr
	w, h                    int32
}

func newCanvas(target uintptr, w, h int32) *canvas {
	c := &canvas{target: target, w: w, h: h}
	c.dc, _, _ = procCreateCompatibleDC.Call(target)
	c.bitmap, _, _ = procCreateCompatibleBitmap.Call(target, uintptr(w), uintptr(h))
	c.old, _, _ = procSelectObject.Call(c.dc, c.bitmap)
	procGdipCreateFromHDC.Call(c.dc, uintptr(unsafe.Pointer(&c.graphics)))                 // nolint: errcheck
	procGdipSetSmoothingMode.Call(c.graphics, smoothingAntiAlias)                          // nolint: errcheck
	procGdipSetInterpolationMode.Call(c.graphics, interpolationBicubic)                    // nolint: errcheck
	procGdipSetPixelOffsetMode.Call(c.graphics, pixelOffsetHalf)                           // nolint: errcheck
	procSetBkMode.Call(c.dc, transparent)                                                  // nolint: errcheck
	return c
}

func (c *canvas) present() {
	procGdipDeleteGraphics.Call(c.graphics)                                                       // nolint: errcheck
	procBitBlt.Call(c.target, 0, 0, uintptr(c.w), uintptr(c.h), c.dc, 0, 0, srcCopy)              // nolint: errcheck
	procSelectObject.Call(c.dc, c.old)                                                            // nolint: errcheck
	procDeleteObject.Call(c.bitmap)                                                               // nolint: errcheck
	procDeleteDC.Call(c.dc)                                                                       // nolint: errcheck
}

func (c *canvas) clear(color uint32) {
	brush, _, _ := procCreateSolidBrush.Call(rgb(color))
	r := rect{0, 0, c.w, c.h}
	procFillRect.Call(c.dc, uintptr(unsafe.Pointer(&r)), brush) // nolint: errcheck
	procDeleteObject.Call(brush)                                // nolint: errcheck
}

func (c *canvas) roundedPath(x, y, w, h, r int32) uintptr {
	var path uintptr
	procGdipCreatePath.Call(0, uintptr(unsafe.Pointer(&path))) // nolint: errcheck

	d := r * 2
	arc := func(ax, ay int32, start, sweep float32) {
		procGdipAddPathArcI.Call(path, uintptr(ax), uintptr(ay), uintptr(d), uintptr(d), // nolint: errcheck
			uintptr(math.Float32bits(start)), uintptr(math.Float32bits(sweep)))
	}
	arc(x, y, 180, 90)
	arc(x+w-d, y, 270, 90)
	arc(x+w-d, y+h-d, 0, 90)
	arc(x, y+h-d, 90, 90)
	procGdipClosePathFigure.Call(path) // nolint: errcheck
	return path
}

func (c *canvas) fillRound(x, y, w, h, r int32, color uint32) {
	var brush uintptr
	procGdipCreateSolidFill.Call(uintptr(0xFF000000|color), uintptr(unsafe.Pointer(&brush))) // nolint: errcheck
	path := c.roundedPath(x, y, w, h, r)
	procGdipFillPath.Call(c.graphics, brush, path) // nolint: errcheck
	procGdipDeletePath.Call(path)                  // nolint: errcheck
	procGdipDeleteBrush.Call(brush)                // nolint: errcheck
}

func (c *canvas) strokeRound(x, y, w, h, r int32, color uint32, width float32) {
	var pen uintptr
	procGdipCreatePen1.Call(uintptr(0xFF000000|color), uintptr(math.Float32bits(width)), unitPixel, uintptr(unsafe.Pointer(&pen))) // nolint: errcheck
	path := c.roundedPath(x, y, w, h, r)
	procGdipDrawPath.Call(c.graphics, pen, path) // nolint: errcheck
	procGdipDeletePath.Call(path)                // nolint: errcheck
	procGdipDeletePen.Call(pen)                  // nolint: errcheck
}

func (c *canvas) fillCircle(x, y, d int32, color uint32) {
	var brush uintptr
	procGdipCreateSolidFill.Call(uintptr(0xFF000000|color), uintptr(unsafe.Pointer(&brush))) // nolint: errcheck
	procGdipFillEllipseI.Call(c.graphics, brush, uintptr(x), uintptr(y), uintptr(d), uintptr(d)) // nolint: errcheck
	procGdipDeleteBrush.Call(brush) // nolint: errcheck
}

func (c *canvas) drawImage(image uintptr, x, y, w, h int32) {
	procGdipDrawImageRectI.Call(c.graphics, image, uintptr(x), uintptr(y), uintptr(w), uintptr(h)) // nolint: errcheck
}

func (c *canvas) text(s string, font uintptr, color uint32, x, y, w, h int32, flags uintptr) {
	old, _, _ := procSelectObject.Call(c.dc, font)
	procSetTextColor.Call(c.dc, rgb(color)) // nolint: errcheck
	r := rect{x, y, x + w, y + h}
	u, err := windows.UTF16FromString(s)
	if err == nil && len(u) > 1 {
		procDrawTextW.Call(c.dc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(&r)), flags|dtNoPrefix) // nolint: errcheck
	}
	procSelectObject.Call(c.dc, old) // nolint: errcheck
}

// textWidth returns the width of a single line of text.
func (c *canvas) textWidth(s string, font uintptr) int32 {
	old, _, _ := procSelectObject.Call(c.dc, font)
	r := rect{}
	u, err := windows.UTF16FromString(s)
	if err == nil && len(u) > 1 {
		procDrawTextW.Call(c.dc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(&r)), dtCalcRect|dtSingleLine|dtNoPrefix) // nolint: errcheck
	}
	procSelectObject.Call(c.dc, old) // nolint: errcheck
	return r.Right - r.Left
}

func createFont(name string, heightPx int32, weight int32) uintptr {
	font, _, _ := procCreateFontW.Call(
		uintptr(uint32(-heightPx)), 0, 0, 0, uintptr(weight), 0, 0, 0,
		1,          // DEFAULT_CHARSET
		0, 0,       // precision
		5,          // CLEARTYPE_QUALITY
		0,          // pitch and family
		ptr(name),
	)
	return font
}

func setClipboard(hwnd uintptr, s string) bool {
	u, err := windows.UTF16FromString(s)
	if err != nil {
		return false
	}
	if r, _, _ := procOpenClipboard.Call(hwnd); r == 0 {
		return false
	}
	defer procCloseClipboard.Call() // nolint: errcheck
	procEmptyClipboard.Call()       // nolint: errcheck

	size := uintptr(len(u) * 2)
	mem, _, _ := procGlobalAlloc.Call(gmemMoveable, size)
	if mem == 0 {
		return false
	}
	p, _, _ := procGlobalLock.Call(mem)
	if p == 0 {
		return false
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u)), u)
	procGlobalUnlock.Call(mem) // nolint: errcheck

	r, _, _ := procSetClipboardData.Call(cfUnicodeText, mem)
	return r != 0
}

func openExternal(target string) {
	procShellExecuteW.Call(0, ptr("open"), ptr(target), 0, 0, swShowNormal) // nolint: errcheck
}
