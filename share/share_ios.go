// SPDX-License-Identifier: Unlicense OR MIT

package share

/*
#cgo CFLAGS: -Werror -xobjective-c -fmodules -fobjc-arc
#include <UIKit/UIKit.h>
#include <stdint.h>

static void openShare(CFTypeRef viewController, NSArray * obj) {
	UIActivityViewController * activityController = [[UIActivityViewController alloc] initWithActivityItems:obj applicationActivities:nil];
	[(__bridge UIViewController *)viewController presentViewController:activityController animated:YES completion:nil];
}

static void shareText(CFTypeRef viewController, char * text) {
	openShare(viewController, @[@(text)]);
}

static void shareWebsite(CFTypeRef viewController, char * link) {
	openShare(viewController, @[[NSURL URLWithString:@(link)]]);
}

*/
import "C"
import (
	"gioui.org/app"
	"gioui.org/io/event"
)

type share struct {
	window         *app.Window
	viewController C.CFTypeRef
}

func newShare(w *app.Window) *share {
	return &share{
		window: w,
	}
}

func (e *Share) listenEvents(evt event.Event) {
	switch evt := evt.(type) {
	case app.ViewEvent:
		e.viewController = C.CFTypeRef(evt.ViewController)
	}
}

func (e *Share) shareShareable(shareable Shareable) error {
	switch s := shareable.(type) {
	case Text:
		go e.window.Run(func() {
			t := C.CString(s.Text)
			C.shareText(e.viewController, t)
		})
	case Website:
		go e.window.Run(func() {
			l := C.CString(s.Link)
			C.shareWebsite(e.viewController, l)
		})
	default:
		return ErrNotAvailableAction
	}

	return nil
}
