package locale

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	_Kernel32 = windows.NewLazySystemDLL("kernel32")

	// https://docs.microsoft.com/en-us/windows/win32/api/winnls/nf-winnls-getuserdefaultlocalename
	_DefaultUserLang = _Kernel32.NewProc("GetUserDefaultLocaleName")

	// https://docs.microsoft.com/en-us/windows/win32/api/winnls/nf-winnls-getsystemdefaultlocalename
	_DefaultSystemLang = _Kernel32.NewProc("GetSystemDefaultLocaleName")

	// _LocaleNameMaxSize is the "LOCALE_NAME_MAX_LENGTH".
	// https://docs.microsoft.com/en-us/windows/win32/intl/locale-name-constants
	_LocaleNameMaxSize = 85
)

func getLanguage() string {
	lang := make([]uint16, _LocaleNameMaxSize)

	r, _, _ := _DefaultUserLang.Call(uintptr(unsafe.Pointer(&lang[0])), uintptr(_LocaleNameMaxSize))
	if r == 0 {
		_DefaultSystemLang.Call(uintptr(unsafe.Pointer(&lang[0])), uintptr(_LocaleNameMaxSize))
	}

	return windows.UTF16ToString(lang)
}
