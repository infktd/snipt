#ifndef TRAY_DARWIN_H
#define TRAY_DARWIN_H

void setupNativeTray(const void *iconData, int iconLen, const char *version);
void teardownNativeTray(void);
void injectAppMenuItems(void);

#endif
