//go:build darwin && !ios

package theme

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#cgo LDFLAGS: -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

long isDarkMode() {
    @autoreleasepool {
        NSString *appleInterfaceStyle = [[NSUserDefaults standardUserDefaults] stringForKey:@"AppleInterfaceStyle"];
		if (appleInterfaceStyle == nil) {
			return -1;
		}
		if ([appleInterfaceStyle isEqualToString:@"Dark"]) {
			return 1;
		}
        return 0;
    }
}

long isReducedMotion() {
    @autoreleasepool {
        NSNumber *reduceMotion = [[NSUserDefaults standardUserDefaults] objectForKey:@"reduceMotion"];
		if (reduceMotion == nil) {
			return -1;
		}
		if ([reduceMotion boolValue]) {
			return 1;
		}
        return 0;
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
