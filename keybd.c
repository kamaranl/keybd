#include "keybd.h"

#include <ApplicationServices/ApplicationServices.h>

void SendKey(int key, bool shift)
{
    int ms = 1000;
    CGKeyCode keyShift = (CGKeyCode)56;
    CGEventTapLocation tap = kCGHIDEventTap;
    CGEventRef keyDown, keyUp, shiftDown, shiftUp;

    keyDown = CGEventCreateKeyboardEvent(NULL, key, true);
    keyUp = CGEventCreateKeyboardEvent(NULL, key, false);

    if (shift)
    {
        shiftDown = CGEventCreateKeyboardEvent(NULL, keyShift, true);
        shiftUp = CGEventCreateKeyboardEvent(NULL, keyShift, false);
        CGEventSetFlags(keyDown, kCGEventFlagMaskShift);
        CGEventSetFlags(keyUp, kCGEventFlagMaskShift);

        usleep(10 * ms);
        CGEventPost(tap, shiftDown);
        CFRelease(shiftDown);
        usleep(10 * ms);
    }

    CGEventPost(tap, keyDown);
    CFRelease(keyDown);
    usleep(50 * ms);
    CGEventPost(tap, keyUp);
    CFRelease(keyUp);

    if (shift)
    {
        CGEventPost(tap, shiftUp);
        CFRelease(shiftUp);
        usleep(10 * ms);
    }
}
