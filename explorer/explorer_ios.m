// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios
// +build ios

#include <UIKit/UIKit.h>
#include <stdint.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include "_cgo_export.h"

@implementation explorer
- (void)documentPicker:(UIDocumentPickerViewController *)controller didPickDocumentsAtURLs:(NSArray<NSURL *> *)urls {
    NSURL * url = [urls objectAtIndex:0];
    [url startAccessingSecurityScopedResource];

    if (self.mode == 0) {
        createCallback((__bridge_retained CFTypeRef)url);
    } else {
        openCallback((__bridge_retained CFTypeRef)url);
    }
}
- (void)documentPickerWasCancelled:(UIDocumentPickerViewController *)controller {
    if (self.mode == 0) {
        createCallback(0);
    } else {
        openCallback(0);
    }
}
- (void) createFile:(CFTypeRef)viewRef name:(char*)name {
   if (@available(iOS 14, *)) {
        self.controller = (__bridge UIViewController *)viewRef;
        self.picker = [[UIDocumentPickerViewController alloc] initForExportingURLs:@[[NSURL URLWithString:@(name)]] asCopy:true];
        self.picker.delegate = self;

        [self.controller presentViewController:self.picker animated:YES completion:nil];
    }
}
- (void) openFile:(CFTypeRef)viewRef ext:(char*)ext {
    if (@available(iOS 14, *)) {
        NSMutableArray<NSString*> *exts = [[@(ext) componentsSeparatedByString:@","] mutableCopy];
        NSMutableArray<UTType*> *contentTypes = [[NSMutableArray alloc]init];

        int i;
        for (i = 0; i < [exts count]; i++) {
            id utt = UTTypePlainText;
            if (utt != nil) {
                [contentTypes addObject:utt];
            }
        }

        self.controller = (__bridge UIViewController *)viewRef;
        self.picker = [[UIDocumentPickerViewController alloc] initForOpeningContentTypes:contentTypes asCopy:true];
        self.picker.delegate = self;

        [self.controller presentViewController:self.picker animated:YES completion:nil];
    }
}
@end

