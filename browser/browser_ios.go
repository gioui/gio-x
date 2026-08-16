//go:build ios
// +build ios

package browser

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc
extern static CFTypeRef newNSString(unichar *chars, NSUInteger length);
extern static void openUrl(CFTypeRef str);
*/

import "C"
import (
	"unicode/utf16"
	"unsafe"
)

func stringToNSString(str string) C.CFTypeRef {
	u16 := utf16.Encode([]rune(str))
	var chars *C.unichar
	if len(u16) > 0 {
		chars = (*C.unichar)(unsafe.Pointer(&u16[0]))
	}
	return C.newNSString(chars, C.NSUInteger(len(u16)))
}

func OpenUrl(url string) error {
	cstr := stringToNSString(url)
	defer C.CFRelease(cstr)
	C.openUrl(cstr)
	return nil
}
