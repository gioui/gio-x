// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios
// +build darwin,!ios

#include "_cgo_export.h"
#import <Foundation/Foundation.h>
#import <Appkit/AppKit.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>

@implementation explorer_macos
+ (void) createFile:(CFTypeRef)viewRef name:(char*)name {
	NSView *view = (__bridge NSView *)viewRef;

	NSSavePanel *panel = [NSSavePanel savePanel];

    [panel setNameFieldStringValue:@(name)];
	[panel beginSheetModalForWindow:[view window] completionHandler:^(NSInteger result){
		if (result == NSModalResponseOK) {
			createCallback((char *)[[panel URL].absoluteString UTF8String]);
			return;
		}
		createCallback((char *)(""));
	}];
}
+ (void) openFile:(CFTypeRef)viewRef ext:(char*)ext {
	NSView *view = (__bridge NSView *)viewRef;

	NSOpenPanel *panel = [NSOpenPanel openPanel];

    NSMutableArray<NSString*> *exts = [[@(ext) componentsSeparatedByString:@","] mutableCopy];
    NSMutableArray<UTType*> *contentTypes = [[NSMutableArray alloc]init];

    int i;
    for (i = 0; i < [exts count]; i++) {
        id utt = [UTType typeWithFilenameExtension:exts[i]];
        if (utt != nil){
            [contentTypes addObject:utt];
        }
     }

    [(NSSavePanel*)panel setAllowedContentTypes:[NSArray arrayWithArray:contentTypes]];
	[panel beginSheetModalForWindow:[view window] completionHandler:^(NSInteger result){
		if (result == NSModalResponseOK) {
			openCallback((char *)[[panel URL].absoluteString UTF8String]);
			return;
		}
		openCallback((char *)(""));
	}];
}
@end