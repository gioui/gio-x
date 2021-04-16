package theme

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"unsafe"
)

var (
	_RegistryPersonalize = `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`

	_User32 = windows.NewLazySystemDLL("user32")

	// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfow
	_SystemParameters = _User32.NewProc("SystemParametersInfoW")
)

const (
	// _GetClientAreaAnimation is the_SPI_GETCLIENTAREAANIMATION.
	_GetClientAreaAnimation = 0x1042
)

func isDark() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, _RegistryPersonalize, registry.QUERY_VALUE)
	if err != nil {
		return false, ErrNotAvailableAPI
	}
	defer k.Close()

	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false, ErrNotAvailableAPI
	}

	return v == 0, nil
}

func isReducedMotion() (bool, error) {
	disabled := true
	r, _, _ := _SystemParameters.Call(uintptr(_GetClientAreaAnimation), 0, uintptr(unsafe.Pointer(&disabled)), 0)
	if r == 0 {
		return false, ErrNotAvailableAPI
	}

	return !disabled, nil
}
