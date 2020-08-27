package macos

//#cgo LDFLAGS: -framework Foundation -framework UserNotifications

/*
#cgo CFLAGS: -x objective-c -fno-objc-arc -fmodules
#pragma clang diagnostic ignored "-Wformat-security"

#import <stdlib.h>

@import Foundation;
@import UserNotifications;

@interface UNDelegate : NSObject <UNUserNotificationCenterDelegate>
{ }
- (void)userNotificationCenter:(UNUserNotificationCenter *)center didReceiveNotificationResponse:(UNNotificationResponse *)response withCompletionHandler:(void (^)(void))completionHandler;
- (void)userNotificationCenter:(UNUserNotificationCenter *)center willPresentNotification:(UNNotification *)notification withCompletionHandler:(void (^)(UNNotificationPresentationOptions options))completionHandler;
- (void)userNotificationCenter:(UNUserNotificationCenter *)center openSettingsForNotification:(UNNotification *)notification;
@end

@implementation UNDelegate
- (void)userNotificationCenter:(UNUserNotificationCenter *)center didReceiveNotificationResponse:(UNNotificationResponse *)response withCompletionHandler:(void (^)(void))completionHandler {
	NSLog(@"didReceiveNotificationResponse");
}

- (void)userNotificationCenter:(UNUserNotificationCenter *)center willPresentNotification:(UNNotification *)notification withCompletionHandler:(void (^)(UNNotificationPresentationOptions options))completionHandler {
	NSLog(@"willPresentNotification");
}

- (void)userNotificationCenter:(UNUserNotificationCenter *)center openSettingsForNotification:(UNNotification *)notification {
	NSLog(@"openSettingsForNotification");
}
@end

UNUserNotificationCenter *nc;
BOOL enabled;
UNDelegate *del;

void setup() {
	@autoreleasepool {
		NSLog(@"Getting application bundle");
		enabled = NO;
		NSBundle *main = [NSBundle mainBundle];
		if (main.bundleIdentifier == nil) {
			NSLog(@"No app bundle.");
			return;
		}
		NSLog(@"Bundle ID: %@", main.bundleIdentifier);
		NSLog(@"Getting notification center");
		nc = [UNUserNotificationCenter currentNotificationCenter];
		del = [[UNDelegate alloc] init];
		nc.delegate = del;
		NSLog(@"Requesting authorization");
		[nc requestAuthorizationWithOptions: UNAuthorizationOptionBadge | UNAuthorizationOptionSound | UNAuthorizationOptionAlert | UNAuthorizationOptionCriticalAlert completionHandler: ^(BOOL granted, NSError *error){
			NSLog(@"Granted = %s", granted?"true":"false");
			NSLog(@"Error = %@", error);
			enabled = granted;
		}];
	}
}

NSString*
notify(char *id, char *title, char *content) {
	if (enabled != YES) {
		return nil;
	}
	NSString *ret;
	@autoreleasepool {
		NSLog(@"Getting notification center");
		nc = [UNUserNotificationCenter currentNotificationCenter];
		del = [[UNDelegate alloc] init];
		nc.delegate = del;
		[nc getNotificationSettingsWithCompletionHandler: ^(UNNotificationSettings *settings) {
			NSLog(@"Settings: %@", settings);
		}];
		NSLog(@"Creating notification");
		UNMutableNotificationContent *note = [[UNMutableNotificationContent alloc] init];
		note.title = [[NSString alloc] initWithUTF8String: title];
		note.body = [[NSString alloc] initWithUTF8String: content];
		NSString *identifier = [[NSUUID UUID] UUIDString];

		NSLog(@"Creating request");
		UNNotificationRequest *req = [UNNotificationRequest requestWithIdentifier:identifier content: note trigger:nil];
		ret = req.identifier;
		[ret retain];
		NSLog(@"Adding notification request");
		[nc addNotificationRequest:req withCompletionHandler: ^(NSError *error) {
			NSLog(@"added notification. Error: %@", error);
		}];
	}
	return ret;
}

void
cancel(void *nid) {
	[nc removePendingNotificationRequestsWithIdentifiers: @[(NSString*)nid]];
	[nc removeDeliveredNotificationsWithIdentifiers: @[(NSString*)nid]];
}

*/
import "C"

import (
	"runtime"
	"unsafe"
)

func init() {
	runtime.LockOSThread()
	C.setup()
}

type NotificationChannel struct {
	id *C.char
}

func NewNotificationChannel(id string) NotificationChannel {
	return NotificationChannel{ id: C.CString(id) }
}

func (c NotificationChannel) Send(title, text string) (*Notification, error) {
	return notify(c.id, title, text), nil
}

type Notification C.NSString

func notify(cid *C.char, title, content string) *Notification {
	ct := C.CString(title)
	defer C.free(unsafe.Pointer(ct))
	cc := C.CString(content)
	defer C.free(unsafe.Pointer(cc))

	id := C.notify(cid, ct, cc)
	return (*Notification)(id)
}

func (n *Notification) Cancel() error {
	C.cancel(unsafe.Pointer(n))
	return nil
}
