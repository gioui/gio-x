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

    if (self.mode == EXPORT_MODE) {
        exportCallback((__bridge_retained CFTypeRef)url, self.id);
    } else {
        importCallback((__bridge_retained CFTypeRef)url, self.id);
    }
}
- (void)documentPickerWasCancelled:(UIDocumentPickerViewController *)controller {
    if (self.mode == EXPORT_MODE) {
        exportCallback(0, self.id);
    } else {
        importCallback(0, self.id);
    }
}
- (bool) show:(char*)text {
    if (self.mode == EXPORT_MODE) {
        return [self exportFile:text];
    } else {
        return [self importFile:text];
    }
}
- (bool) exportFile:(char*)name {
   if (@available(iOS 14, *)) {
        self.picker = [[UIDocumentPickerViewController alloc] initForExportingURLs:@[[NSURL URLWithString:@(name)]] asCopy:true];
        self.picker.delegate = self;

        [self.controller presentViewController:self.picker animated:YES completion:nil];
        return YES;
    }
    return NO;
}
- (bool) importFile:(char*)ext {
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

        self.picker = [[UIDocumentPickerViewController alloc] initForOpeningContentTypes:contentTypes asCopy:true];
        self.picker.delegate = self;

        [self.controller presentViewController:self.picker animated:YES completion:nil];
        return YES;
    }
    return NO;
}
@end

