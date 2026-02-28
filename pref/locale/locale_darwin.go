package locale

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>

const char* GetPreferredLanguage() {
    @autoreleasepool {
	 	// Get the user's preferred language
		NSArray *preferredLanguages = [NSLocale preferredLanguages];
		if (preferredLanguages.count > 0) {
			return [preferredLanguages.firstObject UTF8String];
		}

		// Fallback to the system locale if preferred language is not available
		NSString *systemLanguage = [[NSLocale currentLocale] localeIdentifier];
		return [systemLanguage  UTF8String];
    }
}
*/
import "C"

func getLanguage() string {
	language := C.GetPreferredLanguage()
	return C.GoString(language)
}
