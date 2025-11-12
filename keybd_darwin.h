#import <Carbon/Carbon.h>

/*!
    @var LastErrorMessage
    @abstract Declaration for a method's last error message to be written to.
*/
char LastErrorMessage[256];

/*!
    @const kVK_None
    @abstract An unassigned virtual key.
*/
const CGKeyCode kVK_None = 0xFFFF;

/*!
    @const kModShift
    @abstract The mask for the Shift key.
*/
const UInt32 kModShift = 0x2;

/*!
    @const kModOption
    @abstract The mask for the Option key.
*/
const UInt32 kModOption = 0x8;

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
    {.mask = kModShift, .vk = kVK_Shift, .flag = kCGEventFlagMaskShift},
    {.mask = kModOption, .vk = kVK_Option, .flag = kCGEventFlagMaskAlternate},
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
  KeyboardLayoutInfo info = {0};

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
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyAction with down=%s key=%d flags=%llu\n",
           keyDown ? "true" : "false", vk, flags);

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
        1: Success | 0: Failure
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
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyPress with key=%d flags=%llu", vk, flags);
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
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyRelease with key=%d flags=%llu", vk, flags);
  return KeyAction(vk, flags, false);
}

/*!
    @function KeyPressAndRelease
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
int KeyPressAndRelease(CGKeyCode vk, CGEventFlags flags, int keyPressDur) {
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyPressAndRelease with key=%d flags=%llu", vk,
           flags);

  int errCount = KeyPress(vk, flags);

  if (keyPressDur > 0)
    usleep(keyPressDur);

  errCount += KeyRelease(vk, flags);

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

  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Unable to perform SetMods with flags=%llu mods=%x", *flags, mods);

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
            int tabSize) {
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

    if (c == '\t' && tabSize > 0) {
      current.vk = kVK_Space;
      numTaps = tabSize;
    }

    for (int j = 0; j < numTaps; j++)
      errCount += KeyPressAndRelease(current.vk, flags, keyPressDur);

    SetMods(&flags, current.mods, next.mods, false);

    if (i < last) {
      current = next;
      usleep(keyDelay);
    }
  }

  return errCount ? 0 : 1;
}
