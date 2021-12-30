// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios
// +build ios

#include <UIKit/UIKit.h>
#include <stdint.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include "_cgo_export.h"

@implementation explorer_picker
- (void)documentPicker:(UIDocumentPickerViewController *)controller didPickDocumentsAtURLs:(NSArray<NSURL *> *)urls {
    NSURL *url = [urls objectAtIndex:0];

    switch (self.mode) {
    case EXPORT_MODE:
        exportCallback((__bridge_retained CFTypeRef)url, self.id);
        return;
    case IMPORT_MODE:
        importCallback((__bridge_retained CFTypeRef)url, self.id);
        return;
    }
}
- (void)documentPickerWasCancelled:(UIDocumentPickerViewController *)controller {
    switch (self.mode) {
    case EXPORT_MODE:
        exportCallback(0, self.id);
        return;
    case IMPORT_MODE:
        importCallback(0, self.id);
        return;
    }
}
@end

CFTypeRef createPicker(CFTypeRef controllerRef, int32_t id) {
	explorer_picker *e = [[explorer_picker alloc] init];
	e.controller = (__bridge UIViewController *)controllerRef;
	e.id = id;
	return (__bridge_retained CFTypeRef)e;
}

bool exportFile(CFTypeRef expl, char * name) {
   if (@available(iOS 14, *)) {
        explorer_picker *explorer = (__bridge explorer_picker *)expl;
        explorer.picker = [[UIDocumentPickerViewController alloc] initForExportingURLs:@[[NSURL URLWithString:@(name)]] asCopy:true];
        explorer.picker.delegate = explorer;
        explorer.mode = EXPORT_MODE;

        [explorer.controller presentViewController:explorer.picker animated:YES completion:nil];
        return YES;
    }
    return NO;
}

bool importFile(CFTypeRef expl, char * ext) {
  if (@available(iOS 14, *)) {
        explorer_picker *explorer = (__bridge explorer_picker *)expl;

        NSMutableArray<NSString*> *exts = [[@(ext) componentsSeparatedByString:@","] mutableCopy];
        NSMutableArray<UTType*> *contentTypes = [[NSMutableArray alloc]init];

        int i;
        for (i = 0; i < [exts count]; i++) {
            UTType *utt = [UTType typeWithFilenameExtension:exts[i]];
            if (utt != nil) {
                [contentTypes addObject:utt];
            }
        }

        explorer.picker = [[UIDocumentPickerViewController alloc] initForOpeningContentTypes:contentTypes asCopy:true];
        explorer.picker.delegate = explorer;
        explorer.mode = IMPORT_MODE;

        [explorer.controller presentViewController:explorer.picker animated:YES completion:nil];
        return YES;
    }
    return NO;
}

CFTypeRef fileWriteHandler(CFTypeRef u) {
    NSURL *url = (__bridge NSURL *)u;
    [url startAccessingSecurityScopedResource];

	NSError *err = nil;
	NSFileHandle *handler = [NSFileHandle fileHandleForWritingToURL:url error:&err];
	if (err != nil) {
		return 0;
	}
	return (__bridge_retained CFTypeRef)handler;
}

CFTypeRef fileReadHandler(CFTypeRef u) {
    NSURL *url = (__bridge NSURL *)u;
    [url startAccessingSecurityScopedResource];

	NSError *err = nil;
	NSFileHandle *handler = [NSFileHandle fileHandleForReadingFromURL:url error:&err];
	if (err != nil) {
		return 0;
	}
	return (__bridge_retained CFTypeRef)handler;
}

bool fileWrite(CFTypeRef handler, uint8_t *b, uint64_t len) {
	if (@available(iOS 13, *)) {
        NSData *data = [NSData dataWithBytes:b length:len];

        NSError *err = nil;
        return [(__bridge NSFileHandle *)handler writeData:data error:&err];
	}
	return NO;
}

uint64_t fileRead(CFTypeRef handler, uint8_t *b, uint64_t len) {
	if (@available(iOS 13, *)) {
	    NSError *err = nil;
		NSData *data = [(__bridge NSFileHandle *)handler readDataUpToLength:len error:&err];
		[data getBytes:b length:data.length];
		return data.length;
	}
	return 0;
}

void closeFile(CFTypeRef handler, CFTypeRef u) {
	[(__bridge NSURL *)u stopAccessingSecurityScopedResource];
	[(__bridge NSFileHandle *)handler closeFile];
}
