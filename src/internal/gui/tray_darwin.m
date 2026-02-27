#include "tray_darwin.h"
#import <Cocoa/Cocoa.h>

extern void goTrayClicked(void);
extern void goMenuManage(void);
extern void goMenuSettings(void);
extern void goMenuCheckForUpdates(void);
extern void goMenuQuit(void);

static const void *trayIconDataPtr = NULL;
static int trayIconDataLen = 0;

@interface SniPTTrayHandler : NSObject
@property (strong) NSStatusItem *statusItem;
@property (strong) NSMenu *contextMenu;
@property (copy)   NSString *appVersion;
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

- (void)aboutAction:(id)sender {
	NSPanel *panel = [[NSPanel alloc]
		initWithContentRect:NSMakeRect(0, 0, 260, 200)
		          styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
		            backing:NSBackingStoreBuffered
		              defer:YES];
	panel.title = @"About snipt";
	panel.releasedWhenClosed = NO;

	NSView *content = panel.contentView;

	// Icon — centered
	NSData *iconData = [NSData dataWithBytes:trayIconDataPtr length:trayIconDataLen];
	NSImage *icon = [[NSImage alloc] initWithData:iconData];
	[icon setTemplate:NO];
	NSImageView *iv = [NSImageView imageViewWithImage:icon];
	iv.frame = NSMakeRect(98, 130, 64, 64);
	[content addSubview:iv];

	// Title
	NSTextField *title = [NSTextField labelWithString:@"snipt"];
	title.font = [NSFont boldSystemFontOfSize:16];
	title.alignment = NSTextAlignmentCenter;
	title.frame = NSMakeRect(0, 100, 260, 24);
	[content addSubview:title];

	// Version
	NSTextField *ver = [NSTextField labelWithString:
		[NSString stringWithFormat:@"Version %@", self.appVersion ?: @"dev"]];
	ver.font = [NSFont systemFontOfSize:12];
	ver.alignment = NSTextAlignmentCenter;
	ver.textColor = [NSColor secondaryLabelColor];
	ver.frame = NSMakeRect(0, 78, 260, 18);
	[content addSubview:ver];

	// Description
	NSTextField *desc = [NSTextField labelWithString:
		@"A snippet manager for your\nterminal and desktop."];
	desc.font = [NSFont systemFontOfSize:12];
	desc.alignment = NSTextAlignmentCenter;
	desc.textColor = [NSColor secondaryLabelColor];
	desc.maximumNumberOfLines = 2;
	desc.frame = NSMakeRect(0, 42, 260, 32);
	[content addSubview:desc];

	// OK button
	NSButton *ok = [NSButton buttonWithTitle:@"OK" target:panel action:@selector(close)];
	ok.frame = NSMakeRect(80, 8, 100, 28);
	ok.keyEquivalent = @"\r";
	[content addSubview:ok];

	[panel center];
	[NSApp activateIgnoringOtherApps:YES];
	[panel makeKeyAndOrderFront:nil];
}

- (void)checkForUpdatesAction:(id)sender { goMenuCheckForUpdates(); }
- (void)manageAction:(id)sender  { goMenuManage(); }
- (void)settingsAction:(id)sender { goMenuSettings(); }
- (void)quitAction:(id)sender    { goMenuQuit(); }

@end

static SniPTTrayHandler *handler = nil;

void setupNativeTray(const void *iconData, int iconLen, const char *version) {
	trayIconDataPtr = iconData;
	trayIconDataLen = iconLen;
	handler = [[SniPTTrayHandler alloc] init];
	handler.appVersion = version ? [NSString stringWithUTF8String:version] : @"dev";

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

	mi = [menu addItemWithTitle:@"About snipt"
	                     action:@selector(aboutAction:)
	              keyEquivalent:@""];
	mi.target = handler;

	mi = [menu addItemWithTitle:@"Check for Updates..."
	                     action:@selector(checkForUpdatesAction:)
	              keyEquivalent:@""];
	mi.target = handler;

	[menu addItem:[NSMenuItem separatorItem]];

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

// injectAppMenuItems inserts Check for Updates and Settings into
// the macOS app menu bar (the first submenu Wails creates).
void injectAppMenuItems(void) {
	if (!handler) return;

	NSMenu *mainMenu = [NSApp mainMenu];
	if (mainMenu.numberOfItems == 0) return;

	NSMenu *appMenu = [mainMenu.itemArray[0] submenu];
	if (!appMenu) return;

	NSInteger idx = 0;
	NSMenuItem *mi;

	mi = [[NSMenuItem alloc] initWithTitle:@"About snipt"
	                                action:@selector(aboutAction:)
	                         keyEquivalent:@""];
	mi.target = handler;
	[appMenu insertItem:mi atIndex:idx];

	mi = [[NSMenuItem alloc] initWithTitle:@"Check for Updates..."
	                                action:@selector(checkForUpdatesAction:)
	                         keyEquivalent:@""];
	mi.target = handler;
	[appMenu insertItem:mi atIndex:idx + 1];

	[appMenu insertItem:[NSMenuItem separatorItem] atIndex:idx + 2];

	mi = [[NSMenuItem alloc] initWithTitle:@"Settings"
	                                action:@selector(settingsAction:)
	                         keyEquivalent:@","];
	mi.target = handler;
	mi.keyEquivalentModifierMask = NSEventModifierFlagCommand;
	[appMenu insertItem:mi atIndex:idx + 3];

	[appMenu insertItem:[NSMenuItem separatorItem] atIndex:idx + 4];
}
