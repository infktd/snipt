#include "tray_darwin.h"
#import <Cocoa/Cocoa.h>

extern void goTrayClicked(void);
extern void goMenuManage(void);
extern void goMenuSettings(void);
extern void goMenuQuit(void);

@interface SniPTTrayHandler : NSObject
@property (strong) NSStatusItem *statusItem;
@property (strong) NSMenu *contextMenu;
@end

@implementation SniPTTrayHandler

- (void)handleClick:(id)sender {
	NSEvent *event = [NSApp currentEvent];
	if (event.type == NSEventTypeRightMouseUp ||
		(event.modifierFlags & NSEventModifierFlagControl)) {
		[NSMenu popUpContextMenu:self.contextMenu
		               withEvent:event
		                 forView:self.statusItem.button];
	} else {
		goTrayClicked();
	}
}

- (void)manageAction:(id)sender  { goMenuManage(); }
- (void)settingsAction:(id)sender { goMenuSettings(); }
- (void)quitAction:(id)sender    { goMenuQuit(); }

@end

static SniPTTrayHandler *handler = nil;

void setupNativeTray(const void *iconData, int iconLen) {
	handler = [[SniPTTrayHandler alloc] init];

	NSStatusItem *item = [[NSStatusBar systemStatusBar]
		statusItemWithLength:NSSquareStatusItemLength];
	handler.statusItem = item;

	NSData *data = [NSData dataWithBytes:iconData length:iconLen];
	NSImage *image = [[NSImage alloc] initWithData:data];
	[image setTemplate:YES];
	[image setSize:NSMakeSize(18, 18)];

	item.button.image   = image;
	item.button.toolTip = @"snipt";
	item.button.target  = handler;
	item.button.action  = @selector(handleClick:);
	[item.button sendActionOn:(NSEventMaskLeftMouseUp | NSEventMaskRightMouseUp)];

	// Context menu shown on right-click / ctrl-click
	NSMenu *menu = [[NSMenu alloc] init];
	NSMenuItem *mi;

	mi = [menu addItemWithTitle:@"Manage"
	                     action:@selector(manageAction:)
	              keyEquivalent:@""];
	mi.target = handler;

	mi = [menu addItemWithTitle:@"Settings"
	                     action:@selector(settingsAction:)
	              keyEquivalent:@""];
	mi.target = handler;

	[menu addItem:[NSMenuItem separatorItem]];

	mi = [menu addItemWithTitle:@"Quit snipt"
	                     action:@selector(quitAction:)
	              keyEquivalent:@""];
	mi.target = handler;

	handler.contextMenu = menu;
}

void teardownNativeTray(void) {
	if (handler && handler.statusItem) {
		[[NSStatusBar systemStatusBar] removeStatusItem:handler.statusItem];
		handler.statusItem = nil;
	}
	handler = nil;
}
