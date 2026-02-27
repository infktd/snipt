#include "mainthread_darwin.h"
#include <dispatch/dispatch.h>

extern void goMainThreadCallback(void);

void dispatchOnMainQueue(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		goMainThreadCallback();
	});
}
