#include <Foundation/Foundation.h>
#include <UIKit/UIKit.h>
#include <stdint.h>

static CFTypeRef newNSString(unichar *chars, NSUInteger length) {
	@autoreleasepool {
		NSString *s = [NSString string];
		if (length > 0) {
			s = [NSString stringWithCharacters:chars length:length];
		}
		return CFBridgingRetain(s);
	}
}

static void openUrl(CFTypeRef str) {
	@autoreleasepool {
		NSString *s = (__bridge NSString *)str;
    NSURL *url = [NSURL URLWithString:s];
    UIApplication *application = [UIApplication sharedApplication];
    [application openURL:url options:@{} completionHandler:nil];
	}
}