// SPDX-License-Identifier: Unlicense OR MIT

package explorer

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"gioui.org/app"
	"gioui.org/io/event"
	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

var (
	// https://docs.microsoft.com/en-us/windows/win32/api/commdlg/
	_Dialog32 = windows.NewLazySystemDLL("comdlg32.dll")

	_GetSaveFileName = _Dialog32.NewProc("GetSaveFileNameW")
	_GetOpenFileName = _Dialog32.NewProc("GetOpenFileNameW")

	_fileOpenDialogCLSID = ole.NewGUID("{DC1C5A9C-E88A-4DDE-A5A1-60F82A20AEF7}")
	_fileOpenDialogIID   = ole.NewGUID("{D57C7288-D4AD-4768-BE02-9D969532D960}")

	// https://docs.microsoft.com/en-us/windows/win32/api/commdlg/ns-commdlg-openfilenamew
	_FlagFileMustExist    = uint32(0x00001000)
	_FlagForceShowHidden  = uint32(0x10000000)
	_FlagOverwritePrompt  = uint32(0x00000002)
	_FlagDisableLinks     = uint32(0x00100000)
	_FlagAllowMultiSelect = uint32(0x00000200)
	_FlagExplorer         = uint32(0x00080000)
	// OFN_NOCHANGEDIR. The common dialogs change the calling process's
	// working directory to the directory of the selected file unless
	// this is set, which silently breaks every relative path the
	// application resolves after the dialog closes.
	_FlagNoChangeDir         = uint32(0x00000008)
	_FOSNoChangeDir          = uint32(0x00000008)
	_FOSPickFolders          = uint32(0x00000020)
	_FOSForceFileSystem      = uint32(0x00000040)
	_FOSPathMustExist        = uint32(0x00000800)
	_CoInitApartmentThreaded = uint32(0x00000002)
	_HResultCancelled        = uintptr(0x800704C7)
	_SigdnFileSystemPath     = uint32(0x80058000)

	_FilePathLength       = uint32(65535)
	_OpenFileStructLength = uint32(unsafe.Sizeof(_OpenFileName{}))
)

const (
	_iModalWindowShow         = 3
	_iFileDialogSetOptions    = 9
	_iFileDialogGetOptions    = 10
	_iFileDialogGetResult     = 20
	_iShellItemGetDisplayName = 5
)

type (
	// _OpenFileName is defined at https://docs.microsoft.com/pt-br/windows/win32/api/commdlg/ns-commdlg-openfilenamew
	_OpenFileName struct {
		StructSize      uint32
		Owner           uintptr
		Instance        uintptr
		Filter          *uint16
		CustomFilter    *uint16
		MaxCustomFilter uint32
		FilterIndex     uint32
		File            *uint16
		MaxFile         uint32
		FileTitle       *uint16
		MaxFileTitle    uint32
		InitialDir      *uint16
		Title           *uint16
		Flags           uint32
		FileOffset      uint16
		FileExtension   uint16
		DefExt          *uint16
		CustData        uintptr
		FnHook          uintptr
		TemplateName    *uint16
		PvReserved      uintptr
		DwReserved      uint32
		FlagsEx         uint32
	}
)

type explorer struct {
	owner atomic.Uintptr
}

func newExplorer(_ *app.Window) *explorer {
	return &explorer{}
}

func (e *Explorer) listenEvents(evt event.Event) {
	if view, ok := evt.(app.Win32ViewEvent); ok {
		e.owner.Store(view.HWND)
	}
}

func (e *Explorer) exportFile(name string) (io.WriteCloser, error) {
	pathUTF16 := make([]uint16, _FilePathLength)
	copy(pathUTF16, windows.StringToUTF16(name))

	open := _OpenFileName{
		File:          &pathUTF16[0],
		MaxFile:       _FilePathLength,
		Filter:        buildFilter([]string{filepath.Ext(name)}),
		FileExtension: uint16(strings.Index(name, filepath.Ext(name))),
		Flags:         _FlagOverwritePrompt | _FlagNoChangeDir,
		StructSize:    _OpenFileStructLength,
	}

	if r, _, _ := _GetSaveFileName.Call(uintptr(unsafe.Pointer(&open))); r == 0 {
		return nil, ErrUserDecline
	}

	path := windows.UTF16ToString(pathUTF16)
	if len(path) == 0 {
		return nil, ErrUserDecline
	}

	return os.Create(path)
}

func (e *Explorer) importFile(extensions ...string) (io.ReadCloser, error) {
	pathUTF16 := make([]uint16, _FilePathLength)

	open := _OpenFileName{
		File:       &pathUTF16[0],
		MaxFile:    _FilePathLength,
		Filter:     buildFilter(extensions),
		Flags:      _FlagFileMustExist | _FlagForceShowHidden | _FlagDisableLinks | _FlagNoChangeDir,
		StructSize: _OpenFileStructLength,
	}

	if r, _, _ := _GetOpenFileName.Call(uintptr(unsafe.Pointer(&open))); r == 0 {
		return nil, ErrUserDecline
	}

	path := windows.UTF16ToString(pathUTF16)
	if len(path) == 0 {
		return nil, ErrUserDecline
	}

	return os.Open(path)
}

func (e *Explorer) readFile(uri string) (io.ReadCloser, error) {
	return os.Open(uri)
}

func (e *Explorer) importFiles(extensions ...string) ([]io.ReadCloser, error) {
	pathUTF16 := make([]uint16, _FilePathLength)

	open := _OpenFileName{
		File:       &pathUTF16[0],
		MaxFile:    _FilePathLength,
		Filter:     buildFilter(extensions),
		Flags:      _FlagFileMustExist | _FlagForceShowHidden | _FlagDisableLinks | _FlagAllowMultiSelect | _FlagExplorer | _FlagNoChangeDir,
		StructSize: _OpenFileStructLength,
	}

	if r, _, _ := _GetOpenFileName.Call(uintptr(unsafe.Pointer(&open))); r == 0 {
		return nil, ErrUserDecline
	}

	paths := make([]string, 0)
	currentPath := make([]uint16, 0)
	for _, char := range pathUTF16 {
		if char == 0 {
			if len(currentPath) > 0 {
				paths = append(paths, windows.UTF16ToString(currentPath))
				currentPath = currentPath[:0]
			}
		} else {
			currentPath = append(currentPath, char)
		}
	}

	if len(paths) == 0 {
		return nil, ErrUserDecline
	}
	filePaths := paths
	if len(paths) > 1 {
		dir := paths[0]
		filePaths = make([]string, len(paths)-1)
		for i, file := range paths[1:] {
			filePaths[i] = filepath.Join(dir, file)
		}
	}

	files := make([]io.ReadCloser, len(filePaths))
	for i, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			for _, file := range files {
				if file != nil {
					file.Close()
				}
			}
			return nil, err
		}
		files[i] = file
	}

	return files, nil
}

func (e *Explorer) chooseFolder() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := windows.CoInitializeEx(0, _CoInitApartmentThreaded)
	// S_FALSE still requires a matching CoUninitialize call.
	if err != nil && err != syscall.Errno(1) {
		return "", err
	}
	defer windows.CoUninitialize()

	dialog, err := ole.CreateInstance(_fileOpenDialogCLSID, _fileOpenDialogIID)
	if err != nil {
		return "", err
	}
	defer dialog.Release()

	var options uint32
	if _, err := fileDialogCall(dialog, _iFileDialogGetOptions, uintptr(unsafe.Pointer(&options))); err != nil {
		return "", err
	}
	options |= _FOSNoChangeDir | _FOSPickFolders | _FOSForceFileSystem | _FOSPathMustExist
	if _, err := fileDialogCall(dialog, _iFileDialogSetOptions, uintptr(options)); err != nil {
		return "", err
	}

	hr, err := fileDialogCall(dialog, _iModalWindowShow, e.owner.Load())
	if hr == _HResultCancelled {
		return "", ErrUserDecline
	}
	if err != nil {
		return "", err
	}

	var item *ole.IUnknown
	if _, err := fileDialogCall(dialog, _iFileDialogGetResult, uintptr(unsafe.Pointer(&item))); err != nil {
		return "", err
	}
	if item == nil {
		return "", ErrNotAvailable
	}
	defer item.Release()

	var path *uint16
	if _, err := fileDialogCall(item, _iShellItemGetDisplayName, uintptr(_SigdnFileSystemPath), uintptr(unsafe.Pointer(&path))); err != nil {
		return "", err
	}
	if path == nil {
		return "", ErrNotAvailable
	}
	defer windows.CoTaskMemFree(unsafe.Pointer(path))
	return windows.UTF16PtrToString(path), nil
}

func fileDialogCall(obj *ole.IUnknown, index int, args ...uintptr) (uintptr, error) {
	if obj == nil || obj.RawVTable == nil {
		return 0, ole.NewError(ole.E_POINTER)
	}
	vtable := (*[32]uintptr)(unsafe.Pointer(obj.RawVTable))
	callArgs := make([]uintptr, len(args)+1)
	callArgs[0] = uintptr(unsafe.Pointer(obj))
	copy(callArgs[1:], args)
	hr, _, _ := syscall.SyscallN(vtable[index], callArgs...)
	if int32(hr) < 0 {
		return hr, ole.NewError(hr)
	}
	return hr, nil
}

func buildFilter(extensions []string) *uint16 {
	if len(extensions) <= 0 {
		return nil
	}

	patterns := make([]string, len(extensions))
	for i, extension := range extensions {
		if !strings.HasPrefix(extension, "*") {
			extension = "*" + extension
		}
		patterns[i] = extension
	}
	filter := strings.ToUpper(strings.Join(patterns, ";"))

	f := utf16.Encode([]rune(filter + "\x00" + filter + "\x00\x00"))
	return &f[0]
}
