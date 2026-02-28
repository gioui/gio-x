//go:build ios

package theme

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#cgo LDFLAGS: -framework UIKit
#import <Foundation/Foundation.h>
#import <UIKit/UIKit.h>

long isDarkMode() {
    @autoreleasepool {
        if (@available(iOS 13.0, *)) {
            UIUserInterfaceStyle style = [UITraitCollection currentTraitCollection].userInterfaceStyle;
			if (style == UIUserInterfaceStyleDark) {
				return 1;
			}
            return 0;
        }
		return -1;
    }
}

long isReducedMotion() {
    @autoreleasepool {
		if (@available(iOS 8.0, *)) {
        	if (UIAccessibilityIsReduceMotionEnabled()) {
				return 1;
			}
			return 0;
		}
		return -1;
    }
}
*/
import "C"

func isDark() (bool, error) {
	i := C.isDarkMode()
	if i < 0 {
		return false, ErrNotAvailableAPI
	}
	return i >= 1, nil
}

func isReducedMotion() (bool, error) {
	i := C.isReducedMotion()
	if i < 0 {
		return false, ErrNotAvailableAPI
	}
	return i >= 1, nil
}
