/*!
    @header KEYBD_H
    KEYBD_H implements functions that allow manipulation of the keyboard.
    @language C
    @updated 2025-11-14
    @author Kamaran Layne
*/
#ifndef KEYBD_H
#define KEYBD_H

#include <Carbon/Carbon.h>
#include <stdarg.h>
#include <string.h>

/*!
    @var LastErrorMessage
    @abstract Declaration for a method's last error message to be written to.
*/
char LastErrorMessage[256];

/*!
    @function set_LastErrorMessage
    @abstract Sets LastErrorMessage
    @param __format
        The format template string to print.
    @param ...args
        List of args to provide to __format.
*/
void set_LastErrorMessage(const char *__format, ...) {
  char prefix[64] = "Error calling";
  char message[192];
  va_list args;

  va_start(args, __format);
  vsnprintf(message, sizeof(message), __format, args);
  va_end(args);

  snprintf(LastErrorMessage, sizeof(LastErrorMessage), "%s %s\n", prefix,
           message);
}

/*!
    @const kVK_None
    @abstract An unassigned virtual key.
*/
const CGKeyCode kVK_None = 0xFFFF;

/*!
    @const kMod_Shift
    @abstract The mask for the Shift key.
*/
const UInt32 kMod_Shift = 0x2;

/*!
    @const kMod_Option
    @abstract The mask for the Option key.
*/
const UInt32 kMod_Option = 0x8;

/*!
    @typedef KeyboardLayoutInfo
    @abstract Contains the local keyboard layout and type.
    @field kbLayout Specifies the keyboard layout.
    @field kbType Specifies the keyboard type.
*/
typedef struct {
  UCKeyboardLayout *kbLayout;
  int kbType;
} KeyboardLayoutInfo;

/*!
    @typedef KeyTranslation
    @abstract Contains a virtual key code and its modifier mask.
    @field vk Specifies the virtual key code.
    @field mods Specifies the modifier mask.
*/
typedef struct {
  CGKeyCode vk;
  UInt32 mods;
} KeyTranslation;

/*!
    @typedef Modifier
    @abstract Contains a modifier's mask, virtual key code, and event flag.
    @field mask Specifies the modifier's mask.
    @field vk Specifies the modifier's virtual key code.
    @field flag Specifies the modifier's event flag.
*/
typedef struct {
  UInt32 mask;
  CGKeyCode vk;
  CGEventFlags flag;
} Modifier;

/*!
    @var StandardMods
    @abstract Modifier array of the two most common modifier keys on MacOS.
*/
Modifier StandardMods[2] = {
    {.mask = kMod_Shift, .vk = kVK_Shift, .flag = kCGEventFlagMaskShift},
    {.mask = kMod_Option, .vk = kVK_Option, .flag = kCGEventFlagMaskAlternate},
};

/*!
    @function TranslateChar
    @abstract Translates a character into its virtual key code and modifier
   mask.
    @param c
        Character to translate.
    @param kli
        Keyboard layout information.
    @returns
        KeyTranslation
*/
KeyTranslation TranslateChar(UniChar c, KeyboardLayoutInfo kli) {
  switch (c) {
  case '\r':
    return (KeyTranslation){.vk = kVK_None, .mods = 0};
  case '\n':
    return (KeyTranslation){.vk = kVK_Return, .mods = 0};
  case '\t':
    return (KeyTranslation){.vk = kVK_Tab, .mods = 0};
  case ' ':
    return (KeyTranslation){.vk = kVK_Space, .mods = 0};
  }

  set_LastErrorMessage("TranslateChar(c=%c, kli={.kbLayout=%d, .kbType=%d})", c,
                       kli.kbLayout, kli.kbType);

  KeyTranslation keymap = {.vk = kVK_None, .mods = 0};

  for (UInt32 mods = 0; mods < (1 << 4); mods++) {
    for (int key = 0; key < 128; key++) {
      UniChar chars[4];
      UniCharCount len = 0;
      UInt32 deadKeyState = 0;
      OSStatus status =
          UCKeyTranslate(kli.kbLayout, key, kUCKeyActionDown, mods, kli.kbType,
                         kUCKeyTranslateNoDeadKeysBit, &deadKeyState,
                         sizeof(chars) / sizeof(chars[0]), &len, chars);

      if (status == noErr && len > 0 && chars[0] == c) {
        keymap.vk = key;
        keymap.mods = mods;
        return keymap;
      }
    }
  }

  return keymap;
}

/*!
    @function GetKeyboardLayoutInfo
    @abstract Identifies the local keyboard layout and type.
    @result
        KeyboardLayoutInfo
*/
KeyboardLayoutInfo GetKeyboardLayoutInfo() {
  KeyboardLayoutInfo info;

  TISInputSourceRef layoutRef = TISCopyCurrentKeyboardLayoutInputSource();
  CFDataRef layoutData = (CFDataRef)TISGetInputSourceProperty(
      layoutRef, kTISPropertyUnicodeKeyLayoutData);

  info.kbLayout = (UCKeyboardLayout *)CFDataGetBytePtr(layoutData);
  info.kbType = LMGetKbdType();

  if (layoutRef)
    CFRelease(layoutRef);

  return info;
}

/*!
    @function KeyAction
    @abstract Creates and posts a key event.
    @param vk
        The virtual key code to post.
    @param flags
        The modifier event flags to post with the virtual key.
    @param keyDown
        Specifies a key-down event.
    @return
        1: Success | 0: Failure
    @var LastErrorMessage
        The last error message is populated if the call fails.
*/
int KeyAction(CGKeyCode vk, CGEventFlags flags, bool keyDown) {
  set_LastErrorMessage("KeyAction(vk=%d, flags=%llu, keyDown=%s)", vk, flags,
                       keyDown ? "true" : "false");

  CGEventRef event = CGEventCreateKeyboardEvent(NULL, vk, keyDown);
  if (!event)
    return 0;

  CGEventSetFlags(event, flags);
  CGEventPost(kCGHIDEventTap, event);
  CFRelease(event);

  return 1;
}

/*!
    @function KeyIsDown
    @abstract Retrieves the current "down" state of vk.
    @param vk
        The virtual key code to check.
    @return
        1: True | 0: False
*/
int KeyIsDown(CGKeyCode vk) {
  return CGEventSourceKeyState(kCGEventSourceStateHIDSystemState, vk) ? 1 : 0;
}

/*!
    @function KeyPress
    @abstract Posts a key press event to the system.
    @param vk
        The virtual key code to post.
    @param flags
        The modifier event flags to post with the virtual key.
    @return
        1: Success | 0: Failure
    @var LastErrorMessage
        The last error message is populated if the call fails.
*/
int KeyPress(CGKeyCode vk, CGEventFlags flags) {
  set_LastErrorMessage("KeyPress(vk=%d, flags=%llu)", vk, flags);
  return KeyAction(vk, flags, true);
}

/*!
    @function KeyRelease
    @abstract Posts a key release event to the system.
    @param vk
        The virtual key code to post.
    @param flags
        The modifier event flags to post with the virtual key.
    @return
        1: Success | 0: Failure
    @var LastErrorMessage
        The last error message is populated if the call fails.
*/
int KeyRelease(CGKeyCode vk, CGEventFlags flags) {
  set_LastErrorMessage("KeyRelease(vk=%d, flags=%llu)", vk, flags);
  return KeyAction(vk, flags, false);
}

/*!
    @function KeyTap
    @abstract Performs a key press & release with a pause in between.
    @param vk
        The virtual key code to post.
    @param flags
        The modifier event flags to post with the virtual key.
    @return
        1: Success | 0: Failure
    @var LastErrorMessage
        The last error message is populated if the call fails.
*/
int KeyTap(CGKeyCode vk, CGEventFlags flags, int keyPressDur) {
  set_LastErrorMessage("KeyTap(vk=%d, flags=%llu, keyPressDur=%d)", vk, flags,
                       keyPressDur);

  int errCount;

  if (!KeyPress(vk, flags))
    errCount++;

  if (keyPressDur > 0)
    usleep(keyPressDur);

  if (!KeyRelease(vk, flags))
    errCount++;

  return errCount ? 0 : 1;
}

/*!
    @function SetMods
    @abstract Sets a modifier's physical state and event flags.
    @param flags
        The modifier event flags to post with the virtual key.
    @param mods
        The modifier mask to set state and flags for.
    @param modsNext
        The next modifier key to set state and flags for.
    @param keyDown
        Specifies a key-down event.
    @return
        1: Success | 0: Failure
    @var LastErrorMessage
        The last error message is populated if the call fails.
*/
int SetMods(CGEventFlags *flags, UInt32 mods, UInt32 modsNext, bool keyDown) {
  set_LastErrorMessage("SetMods(*flags=%llu, mods=%x, modsNext=%x, keyDown=%s)",
                       *flags, mods, modsNext, keyDown ? "true" : "false");

  int counter = 0;

  for (int i = 0; i < 2; i++) {
    Modifier m = StandardMods[i];

    if ((mods & m.mask) != 0) {
      *flags |= m.flag;

      if (!keyDown) {
        if ((modsNext & m.mask) == 0) {
          KeyRelease(m.vk, 0);
          counter++;
        }
      } else {
        if (!KeyIsDown(m.vk)) {
          KeyPress(m.vk, *flags);
          counter++;
        }
      }
    }
  }

  return counter;
}

/*!
    @function TypeStr
    @abstract Types the provided string.
    @param str
        The string to type.
    @param modPressDur
        How long to keep a modifier key pressed before releasing.
    @param keyPressDur
        How long to keep a key pressed before releasing.
    @param keyDelay
        How long to wait after releasing a key before pressing the next key.
    @return
        1: Success | 0: Failure
*/
int TypeStr(const char *str, int modPressDur, int keyPressDur, int keyDelay,
            int tabsToSpaces, int tabSize) {
  set_LastErrorMessage("TypeStr(*str=%.100s, modPressDur=%d, keyPressDur=%d, "
                       "keyDelay=%d, tabsToSpaces=%d, tabSize=%d)",
                       str, modPressDur, keyPressDur, keyDelay, tabsToSpaces,
                       tabSize);

  int len = strlen(str);
  int last = len - 1;
  int errCount = 0;
  KeyboardLayoutInfo kbInfo = GetKeyboardLayoutInfo();
  KeyTranslation current, next;
  usleep(keyDelay);

  for (size_t i = 0; i < len; i++) {
    UniChar c = str[i];
    CGEventFlags flags = 0;

    if (i == 0)
      current = TranslateChar(c, kbInfo);

    if (i < last)
      next = TranslateChar(str[i + 1], kbInfo);
    else if (i == last)
      next = (KeyTranslation){0};

    if (SetMods(&flags, current.mods, 0, true))
      usleep(modPressDur);

    int numTaps = 1;

    if (c == '\t' && tabsToSpaces) {
      current.vk = kVK_Space;
      numTaps = tabSize;
    }

    for (int j = 0; j < numTaps; j++)
      if (!KeyTap(current.vk, flags, keyPressDur))
        errCount++;

    SetMods(&flags, current.mods, next.mods, false);

    if (i < last) {
      current = next;
      usleep(keyDelay);
    }
  }

  return errCount ? 0 : 1;
}

#endif
