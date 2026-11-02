package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

//go:embed assets/logo.png
var logoPNG []byte

const (
	windowClass = "CasaOSWindowsLauncher"
	buttonClass = "CasaOSLauncherButton"

	wmAppState   = wmApp + 1
	wmAppTray    = wmApp + 2
	wmAppShow    = wmApp + 3
	wmAppQuit    = wmApp + 4
	wmAppDestroy = wmApp + 5

	timerCopied = 1

	idToggle = 100 + iota
	idOpen
	idCopy
	idLinkLocal
	idLinkLAN
	idAutoDocker
	idData
	idLogs
	idGetDocker
	idMenuShow
	idMenuOpen
	idMenuToggle
	idMenuQuit

	windowW = 420
	windowH = 604
)

type buttonKind int

const (
	kindPrimary buttonKind = iota
	kindSecondary
	kindLink
	kindText
	kindCheck
)

type button struct {
	hwnd    uintptr
	id      int
	kind    buttonKind
	text    string
	bg      uint32
	danger  bool
	checked bool
	hover   bool
	pressed bool
	focused bool
	enabled bool
	visible bool
}

type palette struct {
	bg, card, border, text, muted                uint32
	accent, accentHover, danger, dangerHover     uint32
	disabled, disabledText, link, ok, warn, fail uint32
	secondaryHover, off                          uint32
}

var (
	lightPalette = palette{
		bg: 0xF4F6F9, card: 0xFFFFFF, border: 0xE1E5EB, text: 0x1B2330, muted: 0x667085,
		accent: 0x1E72F5, accentHover: 0x175FD1, danger: 0xD93A42, dangerHover: 0xB92F36,
		disabled: 0xD5DAE1, disabledText: 0x7B8494, link: 0x1763D6, ok: 0x17A34A, warn: 0xD98A04, fail: 0xD93A42,
		secondaryHover: 0xEEF1F5, off: 0x98A2B3,
	}
	darkPalette = palette{
		bg: 0x17191D, card: 0x22252B, border: 0x343841, text: 0xF1F3F6, muted: 0xA0A8B5,
		accent: 0x2F7FF7, accentHover: 0x4A91FA, danger: 0xE5484D, dangerHover: 0xEE5D62,
		disabled: 0x383C45, disabledText: 0x8A92A0, link: 0x6AA6FF, ok: 0x30C463, warn: 0xE8A317, fail: 0xF0585D,
		secondaryHover: 0x2D3138, off: 0x7C8493,
	}
)

// app is the state of the window. It is only touched by the thread of the window.
type app struct {
	hwnd     uintptr
	instance uintptr
	dpi      int32
	colors   palette
	dark     bool

	server *Server
	paths  Paths

	buttons map[uintptr]*button
	byID    map[int]*button

	fontTitle, fontBody, fontSmall, fontButton, fontSemi uintptr

	logo       uintptr
	logoPixels []byte

	iconLarge, iconSmall uintptr
	trayAdded            bool

	dockerState DockerState
	copied      bool
	quitting    bool

	screenshot string
}

var ui *app

func (a *app) dp(v int32) int32 { return v * a.dpi / 96 }

func systemIsDark() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	v, _, err := key.GetIntegerValue("AppsUseLightTheme")
	return err == nil && v == 0
}

func (a *app) createFonts() {
	for _, f := range []uintptr{a.fontTitle, a.fontBody, a.fontSmall, a.fontButton, a.fontSemi} {
		if f != 0 {
			procDeleteObject.Call(f) // nolint: errcheck
		}
	}
	a.fontTitle = createFont("Segoe UI Semibold", a.dp(30), 600)
	a.fontBody = createFont("Segoe UI", a.dp(14), 400)
	a.fontSmall = createFont("Segoe UI", a.dp(12), 400)
	a.fontButton = createFont("Segoe UI Semibold", a.dp(16), 600)
	a.fontSemi = createFont("Segoe UI Semibold", a.dp(14), 600)
}

func (a *app) loadLogo() {
	src, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		log.Println("logo:", err)
		return
	}
	// RGBA is premultiplied, as the format below; only the order of the channels differs
	rgba := image.NewRGBA(src.Bounds())
	draw.Draw(rgba, src.Bounds(), src, src.Bounds().Min, draw.Src)

	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	a.logoPixels = make([]byte, len(rgba.Pix))
	for i := 0; i+3 < len(rgba.Pix); i += 4 {
		a.logoPixels[i+0] = rgba.Pix[i+2]
		a.logoPixels[i+1] = rgba.Pix[i+1]
		a.logoPixels[i+2] = rgba.Pix[i+0]
		a.logoPixels[i+3] = rgba.Pix[i+3]
	}
	procGdipCreateBitmapFromScan0.Call(uintptr(w), uintptr(h), uintptr(rgba.Stride), pixelFormat32PARGB, // nolint: errcheck
		uintptr(unsafe.Pointer(&a.logoPixels[0])), uintptr(unsafe.Pointer(&a.logo)))
}

func (a *app) loadIcons() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	extract := shell32.NewProc("ExtractIconExW")
	extract.Call(ptr(exe), 0, uintptr(unsafe.Pointer(&a.iconLarge)), uintptr(unsafe.Pointer(&a.iconSmall)), 1) // nolint: errcheck
}

// -------------------------------------------------- creation

func runUI(server *Server, paths Paths, autostart, hidden bool, screenshot string) {
	a := &app{
		server:     server,
		paths:      paths,
		buttons:    map[uintptr]*button{},
		byID:       map[int]*button{},
		screenshot: screenshot,
	}
	ui = a

	a.dark = systemIsDark()
	a.colors = lightPalette
	if a.dark {
		a.colors = darkPalette
	}

	a.instance, _, _ = kernel32.NewProc("GetModuleHandleW").Call(0)

	startGdiplus()
	a.loadLogo()
	a.loadIcons()

	cursor, _, _ := procLoadCursorW.Call(0, idcArrow)

	mainClass := wndClassEx{
		Size:      uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:   windows.NewCallback(mainProc),
		Instance:  a.instance,
		Icon:      a.iconLarge,
		IconSm:    a.iconSmall,
		Cursor:    cursor,
		ClassName: utf16(windowClass),
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&mainClass))) // nolint: errcheck

	hand, _, _ := procLoadCursorW.Call(0, idcHand)
	childClass := wndClassEx{
		Size:      uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:   windows.NewCallback(buttonProc),
		Instance:  a.instance,
		Cursor:    hand,
		ClassName: utf16(buttonClass),
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&childClass))) // nolint: errcheck

	dpi, _, _ := procGetDpiForSystem.Call()
	a.dpi = int32(dpi)
	if a.dpi == 0 {
		a.dpi = 96
	}

	style := uintptr(wsOverlapped | wsCaption | wsSysMenu | wsMinimizeBox | wsClipChild)
	w, h := a.windowSize(style)

	sw, _, _ := procGetSystemMetrics.Call(smCxScreen)
	sh, _, _ := procGetSystemMetrics.Call(smCyScreen)

	a.hwnd, _, _ = procCreateWindowExW.Call(0, ptr(windowClass), ptr("CasaOS"), style,
		(sw-uintptr(w))/2, (sh-uintptr(h))/2, uintptr(w), uintptr(h), 0, 0, a.instance, 0)
	if a.hwnd == 0 {
		log.Println("cannot create the window")
		return
	}

	if d, _, _ := procGetDpiForWindow.Call(a.hwnd); d != 0 {
		a.dpi = int32(d)
	}
	a.createFonts()

	if a.dark {
		on := int32(1)
		procDwmSetWindowAttribute.Call(a.hwnd, dwmwaDarkMode, uintptr(unsafe.Pointer(&on)), 4) // nolint: errcheck
	}

	a.addButton(idToggle, kindPrimary, tr("btnStart"))
	a.addButton(idLinkLocal, kindLink, "")
	a.addButton(idLinkLAN, kindLink, "")
	a.addButton(idOpen, kindPrimary, tr("btnOpen"))
	a.addButton(idCopy, kindSecondary, tr("btnCopy"))
	a.addButton(idAutoDocker, kindCheck, tr("autoDocker"))
	a.addButton(idGetDocker, kindText, tr("linkDocker"))
	a.addButton(idData, kindText, tr("linkData"))
	a.addButton(idLogs, kindText, tr("linkLogs"))
	a.layout()

	a.byID[idAutoDocker].checked = loadSettings(paths).AutoDocker

	a.dockerState = server.docker.State()
	a.addTray()
	a.refresh()

	if !hidden {
		procShowWindow.Call(a.hwnd, swShow)     // nolint: errcheck
		procUpdateWindow.Call(a.hwnd)           // nolint: errcheck
		procSetForegroundWindow.Call(a.hwnd)    // nolint: errcheck
	}

	go a.watchDocker()

	if autostart {
		server.Start()
	}

	if screenshot != "" {
		go func() {
			time.Sleep(1500 * time.Millisecond)
			procPostMessageW.Call(a.hwnd, wmApp+9, 0, 0) // nolint: errcheck
		}()
	}

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if d, _, _ := procIsDialogMessageW.Call(a.hwnd, uintptr(unsafe.Pointer(&m))); d != 0 {
			continue
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m))) // nolint: errcheck
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m))) // nolint: errcheck
	}
}

func (a *app) windowSize(style uintptr) (int32, int32) {
	r := rect{0, 0, a.dp(windowW), a.dp(windowH)}
	procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&r)), style, 0, 0, uintptr(a.dpi)) // nolint: errcheck
	return r.Right - r.Left, r.Bottom - r.Top
}

func (a *app) addButton(id int, kind buttonKind, text string) {
	h, _, _ := procCreateWindowExW.Call(0, ptr(buttonClass), ptr(text), wsChild|wsVisible|wsTabStop,
		0, 0, 10, 10, a.hwnd, uintptr(id), a.instance, 0)
	b := &button{hwnd: h, id: id, kind: kind, text: text, enabled: true, visible: true, bg: a.colors.bg}
	a.buttons[h] = b
	a.byID[id] = b
}

func (a *app) place(id int, x, y, w, h int32) {
	procMoveWindow.Call(a.byID[id].hwnd, uintptr(a.dp(x)), uintptr(a.dp(y)), uintptr(a.dp(w)), uintptr(a.dp(h)), 1) // nolint: errcheck
}

const (
	marginX    = 28
	contentW   = windowW - 2*marginX
	yLogo      = 22
	logoSize   = 68
	yTitle     = 94
	ySubtitle  = 134
	yStatus    = 164
	yToggle    = 202
	toggleH    = 54
	yHint      = 262
	hintH      = 40
	yCard      = 310
	cardH      = 170
	yDocker    = 492
	yCheck     = 516
	yFooter    = 556
	cardInnerX = marginX + 16
	cardInnerW = contentW - 32
)

func (a *app) layout() {
	a.place(idToggle, marginX, yToggle, contentW, toggleH)

	a.place(idLinkLocal, cardInnerX, yCard+30, cardInnerW, 24)
	a.place(idLinkLAN, cardInnerX, yCard+78, cardInnerW, 24)

	half := int32(cardInnerW-8) / 2
	a.place(idOpen, cardInnerX, yCard+118, half, 38)
	a.place(idCopy, cardInnerX+half+8, yCard+118, half, 38)

	a.place(idAutoDocker, marginX, yCheck, contentW-92, 26)
	a.place(idGetDocker, marginX+contentW-110, yCheck, 110, 26)

	a.place(idData, marginX, yFooter, 110, 26)
	a.place(idLogs, marginX+114, yFooter, 60, 26)

	a.byID[idLinkLocal].bg = a.colors.card
	a.byID[idLinkLAN].bg = a.colors.card
	a.byID[idOpen].bg = a.colors.card
	a.byID[idCopy].bg = a.colors.card
}

// -------------------------------------------------- state

func (a *app) setVisible(id int, visible bool) {
	b := a.byID[id]
	if b.visible == visible {
		return
	}
	b.visible = visible
	cmd := uintptr(swHide)
	if visible {
		cmd = swShow
	}
	procShowWindow.Call(b.hwnd, cmd) // nolint: errcheck
}

func (a *app) setEnabled(id int, enabled bool) {
	b := a.byID[id]
	if b.enabled == enabled {
		return
	}
	b.enabled = enabled
	v := uintptr(0)
	if enabled {
		v = 1
	}
	procEnableWindow.Call(b.hwnd, v) // nolint: errcheck
}

func (a *app) setText(id int, text string) {
	b := a.byID[id]
	if b.text != text {
		b.text = text
		procSetWindowTextW.Call(b.hwnd, ptr(text)) // nolint: errcheck
	}
}

func (a *app) refresh() {
	state, _, local, lan := a.server.Snapshot()

	toggle := a.byID[idToggle]
	toggle.danger = state == StateRunning

	switch state {
	case StateStopped, StateError:
		a.setText(idToggle, tr("btnStart"))
		a.setEnabled(idToggle, true)
	case StateStarting:
		a.setText(idToggle, tr("btnStarting"))
		a.setEnabled(idToggle, false)
	case StateRunning:
		a.setText(idToggle, tr("btnStop"))
		a.setEnabled(idToggle, true)
	case StateStopping:
		a.setText(idToggle, tr("btnStopping"))
		a.setEnabled(idToggle, false)
	}

	running := state == StateRunning
	a.setVisible(idLinkLocal, running)
	a.setVisible(idLinkLAN, running && lan != "")
	a.setText(idLinkLocal, local)
	a.setText(idLinkLAN, lan)
	a.setEnabled(idOpen, running)
	a.setEnabled(idCopy, running)

	if a.copied {
		a.setText(idCopy, tr("btnCopied"))
	} else {
		a.setText(idCopy, tr("btnCopy"))
	}

	a.setVisible(idGetDocker, a.dockerState == DockerNotInstalled)
	a.setVisible(idAutoDocker, a.dockerState != DockerNotInstalled)

	a.updateTray(state)

	procInvalidateRect.Call(a.hwnd, 0, 0) // nolint: errcheck
	for _, b := range a.buttons {
		procInvalidateRect.Call(b.hwnd, 0, 0) // nolint: errcheck
	}
}

func (a *app) watchDocker() {
	for {
		time.Sleep(3 * time.Second)
		state := a.server.docker.State()
		if state != a.dockerState {
			a.dockerState = state
			procPostMessageW.Call(a.hwnd, wmAppState, 0, 0) // nolint: errcheck
		}
	}
}

// -------------------------------------------------- tray

func (a *app) trayData() notifyIconData {
	d := notifyIconData{
		Size:            uint32(unsafe.Sizeof(notifyIconData{})),
		Hwnd:            a.hwnd,
		ID:              1,
		Flags:           nifMessage | nifIcon | nifTip,
		CallbackMessage: wmAppTray,
		Icon:            a.iconSmall,
	}
	return d
}

func setUTF16(dst []uint16, s string) {
	u, err := windows.UTF16FromString(s)
	if err != nil {
		return
	}
	if len(u) > len(dst) {
		u = u[:len(dst)]
		u[len(u)-1] = 0
	}
	copy(dst, u)
}

func (a *app) addTray() {
	d := a.trayData()
	setUTF16(d.Tip[:], "CasaOS")
	if r, _, _ := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&d))); r != 0 {
		a.trayAdded = true
	}
}

func (a *app) stateText(state ServerState) string {
	switch state {
	case StateStarting:
		return tr("stateStarting")
	case StateRunning:
		return tr("stateRunning")
	case StateStopping:
		return tr("stateStopping")
	case StateError:
		return tr("stateError")
	}
	return tr("stateStopped")
}

func (a *app) updateTray(state ServerState) {
	if !a.trayAdded {
		return
	}
	d := a.trayData()
	setUTF16(d.Tip[:], "CasaOS – "+a.stateText(state))
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&d))) // nolint: errcheck
}

func (a *app) removeTray() {
	if !a.trayAdded {
		return
	}
	d := a.trayData()
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&d))) // nolint: errcheck
	a.trayAdded = false
}

func (a *app) balloon(title, text string) {
	if !a.trayAdded {
		return
	}
	d := a.trayData()
	d.Flags = nifInfo
	d.InfoFlags = niifInfo
	setUTF16(d.InfoTitle[:], title)
	setUTF16(d.Info[:], text)
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&d))) // nolint: errcheck
}

func (a *app) showMenu() {
	state, _, _, _ := a.server.Snapshot()

	menu, _, _ := procCreatePopupMenu.Call()
	defer procDestroyMenu.Call(menu) // nolint: errcheck

	procAppendMenuW.Call(menu, mfString|mfDefault, idMenuShow, ptr(tr("menuShow"))) // nolint: errcheck

	openFlags := uintptr(mfString)
	if state != StateRunning {
		openFlags |= mfGrayed
	}
	procAppendMenuW.Call(menu, openFlags, idMenuOpen, ptr(tr("menuOpen"))) // nolint: errcheck
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)                          // nolint: errcheck

	switch state {
	case StateRunning:
		procAppendMenuW.Call(menu, mfString, idMenuToggle, ptr(tr("menuStop"))) // nolint: errcheck
	case StateStopped, StateError:
		procAppendMenuW.Call(menu, mfString, idMenuToggle, ptr(tr("menuStart"))) // nolint: errcheck
	default:
		procAppendMenuW.Call(menu, mfString|mfGrayed, idMenuToggle, ptr(a.stateText(state))) // nolint: errcheck
	}

	procAppendMenuW.Call(menu, mfSeparator, 0, 0)                     // nolint: errcheck
	procAppendMenuW.Call(menu, mfString, idMenuQuit, ptr(tr("menuQuit"))) // nolint: errcheck

	var p point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p))) // nolint: errcheck
	procSetForegroundWindow.Call(a.hwnd)               // nolint: errcheck
	cmd, _, _ := procTrackPopupMenu.Call(menu, tpmRightButton|tpmReturnCmd, uintptr(p.X), uintptr(p.Y), 0, a.hwnd, 0)
	procPostMessageW.Call(a.hwnd, wmNull, 0, 0) // nolint: errcheck

	if cmd != 0 {
		a.command(int(cmd))
	}
}

// -------------------------------------------------- actions

func (a *app) showWindow() {
	procShowWindow.Call(a.hwnd, swShow) // nolint: errcheck
	if iconic, _, _ := procIsIconic.Call(a.hwnd); iconic != 0 {
		procShowWindow.Call(a.hwnd, swRestore) // nolint: errcheck
	}
	procSetForegroundWindow.Call(a.hwnd) // nolint: errcheck
}

func (a *app) quit() {
	if a.quitting {
		return
	}
	a.quitting = true
	procShowWindow.Call(a.hwnd, swHide) // nolint: errcheck
	a.server.Stop(func() {
		procPostMessageW.Call(a.hwnd, wmAppDestroy, 0, 0) // nolint: errcheck
	})
}

func (a *app) command(id int) {
	state, _, local, lan := a.server.Snapshot()

	switch id {
	case idToggle, idMenuToggle:
		if state == StateRunning {
			a.server.Stop(nil)
		} else {
			a.server.Start()
		}

	case idOpen, idMenuOpen, idLinkLocal:
		if local != "" {
			openExternal(local)
		}

	case idLinkLAN:
		if lan != "" {
			openExternal(lan)
		}

	case idCopy:
		link := lan
		if link == "" {
			link = local
		}
		if setClipboard(a.hwnd, link) {
			a.copied = true
			procSetTimer.Call(a.hwnd, timerCopied, 1800, 0) // nolint: errcheck
			a.refresh()
		}

	case idAutoDocker:
		b := a.byID[idAutoDocker]
		b.checked = !b.checked
		s := loadSettings(a.paths)
		s.AutoDocker = b.checked
		saveSettings(a.paths, s)
		procInvalidateRect.Call(b.hwnd, 0, 0) // nolint: errcheck
		if b.checked && state == StateRunning {
			go a.server.docker.OnServerStarted()
		}

	case idData:
		if err := os.MkdirAll(a.paths.Data, 0o755); err == nil {
			openExternal(a.paths.Data)
		}

	case idLogs:
		if err := os.MkdirAll(a.paths.Log, 0o755); err == nil {
			openExternal(a.paths.Log)
		}

	case idGetDocker:
		openExternal("https://www.docker.com/products/docker-desktop/")

	case idMenuShow:
		a.showWindow()

	case idMenuQuit:
		a.quit()
	}
}

// -------------------------------------------------- window procedures

func mainProc(hwnd, message, wparam, lparam uintptr) uintptr {
	a := ui
	if a == nil || a.hwnd == 0 {
		r, _, _ := procDefWindowProcW.Call(hwnd, message, wparam, lparam)
		return r
	}

	switch message {
	case wmEraseBkgnd:
		return 1

	case wmPaint:
		a.paint()
		return 0

	case wmCommand:
		a.command(int(wparam & 0xffff))
		return 0

	case wmAppState:
		a.refresh()
		return 0

	case wmAppShow:
		a.showWindow()
		return 0

	case wmAppQuit:
		a.quit()
		return 0

	case wmAppDestroy:
		procDestroyWindow.Call(hwnd) // nolint: errcheck
		return 0

	case wmApp + 9:
		a.saveScreenshot()
		return 0

	case wmAppTray:
		switch uint32(lparam) & 0xffff {
		case wmLButtonUp, wmLButtonDblClk:
			a.showWindow()
		case wmRButtonUp:
			a.showMenu()
		}
		return 0

	case wmTimer:
		if wparam == timerCopied {
			procKillTimer.Call(hwnd, timerCopied) // nolint: errcheck
			a.copied = false
			a.refresh()
		}
		return 0

	case wmDpiChanged:
		a.dpi = hiword(wparam)
		a.createFonts()
		r := (*rect)(unsafe.Pointer(lparam))
		procSetWindowPos.Call(hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0x0014) // nolint: errcheck
		a.layout()
		a.refresh()
		return 0

	case wmClose:
		state, _, _, _ := a.server.Snapshot()
		if state == StateStopped || state == StateError {
			a.quit()
			return 0
		}
		// the server is on: closing the window must not turn it off
		procShowWindow.Call(hwnd, swHide) // nolint: errcheck
		if s := loadSettings(a.paths); !s.TrayHinted {
			s.TrayHinted = true
			saveSettings(a.paths, s)
			a.balloon(tr("trayTitle"), tr("trayText"))
		}
		return 0

	case wmQueryEndSession:
		return 1

	case wmEndSession:
		if wparam != 0 {
			a.server.StopNow()
		}
		return 0

	case wmDestroy:
		a.removeTray()
		procPostQuitMessage.Call(0) // nolint: errcheck
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, message, wparam, lparam)
	return r
}

func buttonProc(hwnd, message, wparam, lparam uintptr) uintptr {
	a := ui
	var b *button
	if a != nil {
		b = a.buttons[hwnd]
	}
	if b == nil {
		r, _, _ := procDefWindowProcW.Call(hwnd, message, wparam, lparam)
		return r
	}

	invalidate := func() { procInvalidateRect.Call(hwnd, 0, 0) } // nolint: errcheck

	switch message {
	case wmEraseBkgnd:
		return 1

	case wmPaint:
		a.paintButton(b)
		return 0

	case wmGetDlgCode:
		return dlgcButton | dlgcUndefPush

	case wmEnable:
		invalidate()
		return 0

	case wmSetFocus:
		b.focused = true
		invalidate()
		return 0

	case wmKillFocus:
		b.focused = false
		b.pressed = false
		invalidate()
		return 0

	case wmMouseMove:
		if !b.hover {
			b.hover = true
			t := trackMouseEvent{Size: uint32(unsafe.Sizeof(trackMouseEvent{})), Flags: tmeLeave, Track: hwnd}
			procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&t))) // nolint: errcheck
			invalidate()
		}
		return 0

	case wmMouseLeave:
		b.hover = false
		invalidate()
		return 0

	case wmLButtonDown:
		procSetFocus.Call(hwnd)   // nolint: errcheck
		procSetCapture.Call(hwnd) // nolint: errcheck
		b.pressed = true
		invalidate()
		return 0

	case wmLButtonUp:
		wasPressed := b.pressed
		b.pressed = false
		procReleaseCapture.Call() // nolint: errcheck
		invalidate()
		var r rect
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r))) // nolint: errcheck
		x, y := loword(lparam), hiword(lparam)
		if wasPressed && b.enabled && x >= 0 && y >= 0 && x < r.Right && y < r.Bottom {
			procPostMessageW.Call(a.hwnd, wmCommand, uintptr(b.id), 0) // nolint: errcheck
		}
		return 0

	case wmKeyDown:
		if (wparam == vkSpace || wparam == vkReturn) && b.enabled {
			procPostMessageW.Call(a.hwnd, wmCommand, uintptr(b.id), 0) // nolint: errcheck
			return 0
		}
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, message, wparam, lparam)
	return r
}

// -------------------------------------------------- painting

func (a *app) paint() {
	var ps paintStruct
	procBeginPaint.Call(a.hwnd, uintptr(unsafe.Pointer(&ps))) // nolint: errcheck
	defer procEndPaint.Call(a.hwnd, uintptr(unsafe.Pointer(&ps))) // nolint: errcheck

	var r rect
	procGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&r))) // nolint: errcheck

	c := newCanvas(ps.Hdc, r.Right, r.Bottom)
	defer c.present()

	k := a.colors
	c.clear(k.bg)

	w := r.Right

	// header
	if a.logo != 0 {
		size := a.dp(logoSize)
		c.drawImage(a.logo, (w-size)/2, a.dp(yLogo), size, size)
	}
	c.text("CasaOS", a.fontTitle, k.text, 0, a.dp(yTitle), w, a.dp(40), dtCenter|dtSingleLine|dtVCenter)
	c.text(tr("subtitle"), a.fontSmall, k.muted, 0, a.dp(ySubtitle), w, a.dp(20), dtCenter|dtSingleLine|dtVCenter)

	state, detail, _, lan := a.server.Snapshot()

	// status
	dot := k.off
	switch state {
	case StateStarting, StateStopping:
		dot = k.warn
	case StateRunning:
		dot = k.ok
	case StateError:
		dot = k.fail
	}
	label := a.stateText(state)
	labelW := c.textWidth(label, a.fontSemi)
	dotSize := a.dp(10)
	total := dotSize + a.dp(8) + labelW
	x := (w - total) / 2
	y := a.dp(yStatus)
	h := a.dp(26)
	c.fillCircle(x, y+(h-dotSize)/2, dotSize, dot)
	c.text(label, a.fontSemi, k.text, x+dotSize+a.dp(8), y, labelW+a.dp(4), h, dtLeft|dtSingleLine|dtVCenter)

	// hint under the button
	hint := detail
	if hint == "" {
		switch state {
		case StateRunning:
			hint = tr("hintRunning")
		case StateStopped:
			hint = tr("hintStopped")
		}
	}
	hintColor := k.muted
	if state == StateError {
		hintColor = k.fail
	}
	c.text(hint, a.fontSmall, hintColor, a.dp(marginX), a.dp(yHint), w-2*a.dp(marginX), a.dp(hintH), dtCenter|dtWordBreak|dtEndEllipsis)

	// card with the addresses
	cx, cy, cw, ch := a.dp(marginX), a.dp(yCard), a.dp(contentW), a.dp(cardH)
	c.fillRound(cx, cy, cw, ch, a.dp(12), k.card)
	c.strokeRound(cx, cy, cw, ch, a.dp(12), k.border, float32(a.dp(1)))

	innerX, innerW := a.dp(cardInnerX), a.dp(cardInnerW)
	if state == StateRunning {
		c.text(tr("labelLocal"), a.fontSmall, k.muted, innerX, cy+a.dp(10), innerW, a.dp(18), dtLeft|dtSingleLine|dtVCenter)
		c.text(tr("labelLAN"), a.fontSmall, k.muted, innerX, cy+a.dp(58), innerW, a.dp(18), dtLeft|dtSingleLine|dtVCenter)
		if lan == "" {
			c.text(tr("noLAN"), a.fontBody, k.muted, innerX, cy+a.dp(78), innerW, a.dp(24), dtLeft|dtSingleLine|dtVCenter)
		}
	} else {
		c.text(tr("noAddress"), a.fontBody, k.muted, innerX, cy+a.dp(14), innerW, a.dp(90), dtCenter|dtSingleLine|dtVCenter)
	}

	// docker
	dockerDot, dockerText := k.off, tr("dockerStopped")
	switch a.dockerState {
	case DockerRunning:
		dockerDot, dockerText = k.ok, tr("dockerRunning")
	case DockerStarting:
		dockerDot, dockerText = k.warn, tr("dockerStarting")
	case DockerNotInstalled:
		dockerText = tr("dockerMissing")
	}
	dy := a.dp(yDocker)
	dSize := a.dp(8)
	c.fillCircle(a.dp(marginX), dy+(a.dp(20)-dSize)/2, dSize, dockerDot)
	c.text(dockerText, a.fontSmall, k.muted, a.dp(marginX)+dSize+a.dp(8), dy, w, a.dp(20), dtLeft|dtSingleLine|dtVCenter)
}

func (a *app) paintButton(b *button) {
	var ps paintStruct
	procBeginPaint.Call(b.hwnd, uintptr(unsafe.Pointer(&ps))) // nolint: errcheck
	defer procEndPaint.Call(b.hwnd, uintptr(unsafe.Pointer(&ps))) // nolint: errcheck

	var r rect
	procGetClientRect.Call(b.hwnd, uintptr(unsafe.Pointer(&r))) // nolint: errcheck

	c := newCanvas(ps.Hdc, r.Right, r.Bottom)
	defer c.present()

	k := a.colors
	c.clear(b.bg)
	w, h := r.Right, r.Bottom
	active := b.hover || b.pressed

	switch b.kind {
	case kindPrimary:
		fill, textColor := k.accent, uint32(0xFFFFFF)
		if b.danger {
			fill = k.danger
		}
		if active && b.enabled {
			fill = k.accentHover
			if b.danger {
				fill = k.dangerHover
			}
		}
		if !b.enabled {
			fill, textColor = k.disabled, k.disabledText
		}
		radius := a.dp(10)
		c.fillRound(0, 0, w-1, h-1, radius, fill)
		if b.focused {
			c.strokeRound(a.dp(2), a.dp(2), w-1-2*a.dp(2), h-1-2*a.dp(2), radius-a.dp(2), 0xFFFFFF, float32(a.dp(1)))
		}
		font := a.fontButton
		if h < a.dp(44) {
			font = a.fontSemi
		}
		c.text(b.text, font, textColor, 0, 0, w, h, dtCenter|dtSingleLine|dtVCenter)

	case kindSecondary:
		fill, textColor := k.card, k.text
		if active && b.enabled {
			fill = k.secondaryHover
		}
		if !b.enabled {
			textColor = k.disabledText
		}
		radius := a.dp(10)
		c.fillRound(0, 0, w-1, h-1, radius, fill)
		border := k.border
		if b.focused {
			border = k.accent
		}
		c.strokeRound(0, 0, w-1, h-1, radius, border, float32(a.dp(1)))
		c.text(b.text, a.fontSemi, textColor, 0, 0, w, h, dtCenter|dtSingleLine|dtVCenter)

	case kindLink:
		c.text(b.text, a.fontSemi, k.link, 0, 0, w, h, dtLeft|dtSingleLine|dtVCenter|dtEndEllipsis)
		if active || b.focused {
			tw := c.textWidth(b.text, a.fontSemi)
			c.fillRound(0, h-a.dp(3), tw, a.dp(1), 0, k.link)
		}

	case kindText:
		color := k.muted
		if active || b.focused {
			color = k.link
		}
		c.text(b.text, a.fontSmall, color, 0, 0, w, h, dtLeft|dtSingleLine|dtVCenter)

	case kindCheck:
		box := a.dp(16)
		y := (h - box) / 2
		radius := a.dp(4)
		if b.checked {
			c.fillRound(0, y, box, box, radius, k.accent)
			// the check mark is the glyph of the system font for symbols
			c.text("", a.symbolFont(), 0xFFFFFF, 0, y, box, box, dtCenter|dtSingleLine|dtVCenter)
		} else {
			border := k.off
			if active || b.focused {
				border = k.accent
			}
			c.fillRound(0, y, box, box, radius, k.card)
			c.strokeRound(0, y, box, box, radius, border, float32(a.dp(1)))
		}
		color := k.muted
		if active || b.focused {
			color = k.text
		}
		c.text(b.text, a.fontSmall, color, box+a.dp(8), 0, w-box-a.dp(8), h, dtLeft|dtSingleLine|dtVCenter)
	}
}

var symbolFontCache = map[int32]uintptr{}

func (a *app) symbolFont() uintptr {
	if f, ok := symbolFontCache[a.dpi]; ok {
		return f
	}
	f := createFont("Segoe Fluent Icons", a.dp(11), 400)
	symbolFontCache[a.dpi] = f
	return f
}

// -------------------------------------------------- screenshot (for testing the window)

func (a *app) saveScreenshot() {
	var r rect
	procGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&r))) // nolint: errcheck
	w, h := r.Right, r.Bottom

	screen, _, _ := procGetDC.Call(a.hwnd)
	defer procReleaseDC.Call(a.hwnd, screen) // nolint: errcheck

	dc, _, _ := procCreateCompatibleDC.Call(screen)
	defer procDeleteDC.Call(dc) // nolint: errcheck
	bitmap, _, _ := procCreateCompatibleBitmap.Call(screen, uintptr(w), uintptr(h))
	defer procDeleteObject.Call(bitmap) // nolint: errcheck
	old, _, _ := procSelectObject.Call(dc, bitmap)

	const pwClientOnly, pwRenderFullContent = 1, 2
	procPrintWindow.Call(a.hwnd, dc, pwClientOnly|pwRenderFullContent) // nolint: errcheck
	procSelectObject.Call(dc, old)                                     // nolint: errcheck

	info := bitmapInfoHeader{Width: w, Height: -h, Planes: 1, BitCount: 32}
	info.Size = uint32(unsafe.Sizeof(info))
	pixels := make([]byte, int(w)*int(h)*4)
	procGetDIBits.Call(dc, bitmap, 0, uintptr(h), uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&info)), 0) // nolint: errcheck

	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for i := 0; i+3 < len(pixels); i += 4 {
		img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = pixels[i+2], pixels[i+1], pixels[i], 255
	}

	out, err := os.Create(a.screenshot)
	if err != nil {
		log.Println("screenshot:", err)
		return
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		log.Println("screenshot:", err)
	}
}
