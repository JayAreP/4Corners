//go:build windows
// +build windows

package gui

import (
	"syscall"
	"unsafe"
)

var (
	comdlg32         = syscall.NewLazyDLL("comdlg32.dll")
	getSaveFileNameW = comdlg32.NewProc("GetSaveFileNameW")
)

type openFilename struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

func openSaveFileDialog() string {
	var filename [syscall.MAX_PATH]uint16
	filter, _ := syscall.UTF16PtrFromString("Benchmark Files (*.dat)\x00*.dat\x00All Files (*.*)\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("Select File Location for Benchmark Device")
	defExt, _ := syscall.UTF16PtrFromString("dat")

	ofn := openFilename{
		lStructSize: uint32(unsafe.Sizeof(openFilename{})),
		lpstrFile:   &filename[0],
		nMaxFile:    syscall.MAX_PATH,
		lpstrFilter: filter,
		lpstrTitle:  title,
		lpstrDefExt: defExt,
		flags:       0x00080000 | 0x00000004, // OFN_EXPLORER | OFN_OVERWRITEPROMPT
	}

	ret, _, _ := getSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "" // User cancelled
	}

	return syscall.UTF16ToString(filename[:])
}
