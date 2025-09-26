#import <Carbon/Carbon.h>

char LastErrorMessage[256];
const CGKeyCode kVK_None = 0xFFFF;

enum {
  ModShift = 0x2,
  ModOption = 0x8,
};

typedef struct {
  UCKeyboardLayout *layout;
  int kind;
} KeyboardInfo;

typedef struct {
  CGKeyCode vk;
  UInt32 mods;
} KeyMapping;

typedef struct {
  UInt32 mask;
  CGKeyCode key;
  CGEventFlags flag;
} ModifierSet;

ModifierSet standardMods[2] = {
    {.mask = ModShift, .key = kVK_Shift, .flag = kCGEventFlagMaskShift},
    {.mask = ModOption, .key = kVK_Option, .flag = kCGEventFlagMaskAlternate},
};

/**
 *
 *
 * Maps a Unicode character to its corresponding virtual key code and modifier
 * mask for the current keyboard layout.
 *
 * @param c      The Unicode character to map.
 * @param kbInfo Keyboard layout information.
 * @return       KeyMapping struct with virtual key code and modifiers.
 */
KeyMapping CharToVKAndMods(UniChar c, KeyboardInfo kbInfo) {
  // prioritize whitespace
  switch (c) {
  case '\r':
    return (KeyMapping){.vk = kVK_None, .mods = 0};
  case '\n':
    return (KeyMapping){.vk = kVK_Return, .mods = 0};
  case '\t':
    return (KeyMapping){.vk = kVK_Tab, .mods = 0};
  case ' ':
    return (KeyMapping){.vk = kVK_Space, .mods = 0};
  }

  KeyMapping keymap = {.vk = kVK_None, .mods = 0};

  // loop through modifier bits
  for (UInt32 mods = 0; mods < (1 << 4); mods++) {
    // loop through keys
    for (int key = 0; key < 128; key++) {
      UniChar chars[4];
      UniCharCount len = 0;
      UInt32 deadKeyState = 0;
      OSStatus status = UCKeyTranslate(
          kbInfo.layout, key, kUCKeyActionDown, mods, kbInfo.kind,
          kUCKeyTranslateNoDeadKeysBit, &deadKeyState,
          sizeof(chars) / sizeof(chars[0]), &len, chars);

      if (status == noErr && len > 0 && chars[0] == c) {
        keymap.vk = key;
        keymap.mods = mods;
        return keymap;
      }
    } // for key
  }   // for mods

  return keymap; // not available
}

/**
 *
 *
 * Retrieves the current keyboard layout and type information.
 *
 * @return KeyboardInfo struct with layout pointer and keyboard kind.
 */
KeyboardInfo GetKeyboardInfo() {
  KeyboardInfo info = {0};

  TISInputSourceRef layoutRef = TISCopyCurrentKeyboardLayoutInputSource();
  CFDataRef layoutData = (CFDataRef)TISGetInputSourceProperty(
      layoutRef, kTISPropertyUnicodeKeyLayoutData);

  info.layout = (UCKeyboardLayout *)CFDataGetBytePtr(layoutData);
  info.kind = LMGetKbdType();

  if (layoutRef)
    CFRelease(layoutRef);

  return info;
}

/**
 *
 *
 * Simulates a key press or release event for a given key code and modifier
 * flags.
 *
 * @param key     Virtual key code.
 * @param flags   Modifier flags (e.g., shift, option).
 * @param keyDown True for key press, false for key release.
 * @return        1 on success, 0 on failure.
 */
int KeyAction(CGKeyCode key, CGEventFlags flags, bool keyDown) {
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyAction with down=%s key=%d flags=%llu\n",
           keyDown ? "true" : "false", key, flags);

  CGEventRef event = CGEventCreateKeyboardEvent(NULL, key, keyDown);
  if (!event)
    return 0;

  CGEventSetFlags(event, flags);
  CGEventPost(kCGHIDEventTap, event);
  CFRelease(event);

  return 1;
}

/**
 *
 *
 * Checks if a specific key is currently pressed down.
 *
 * @param key Virtual key code.
 * @return    1 if key is down, 0 otherwise.
 */
int KeyIsDown(CGKeyCode key) {
  return CGEventSourceKeyState(kCGEventSourceStateHIDSystemState, key) ? 1 : 0;
}

/**
 *
 *
 * Simulates a key press event for a given key code and modifier flags.
 *
 * @param key   Virtual key code.
 * @param flags Modifier flags.
 * @return      1 on success, 0 on failure.
 */
int KeyPress(CGKeyCode key, CGEventFlags flags) {
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyPress with key=%d flags=%llu", key, flags);
  return KeyAction(key, flags, true);
}

/**
 *
 *
 * Simulates a key release event for a given key code and modifier flags.
 *
 * @param key   Virtual key code.
 * @param flags Modifier flags.
 * @return      1 on success, 0 on failure.
 */
int KeyRelease(CGKeyCode key, CGEventFlags flags) {
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyRelease with key=%d flags=%llu", key, flags);
  return KeyAction(key, flags, false);
}

/**
 *
 *
 * Simulates a key press followed by a key release, with an optional delay.
 *
 * @param key         Virtual key code.
 * @param flags       Modifier flags.
 * @param keyPressDur Microseconds to hold the key down.
 * @return            1 on success, 0 on failure.
 */
int KeyPressAndRelease(CGKeyCode key, CGEventFlags flags, int keyPressDur) {
  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Error performing KeyPressAndRelease with key=%d flags=%llu", key,
           flags);

  int errCount = KeyPress(key, flags);

  if (keyPressDur > 0)
    usleep(keyPressDur);

  errCount += KeyRelease(key, flags);

  return errCount ? 0 : 1;
}

/**
 *
 *
 * Presses or releases modifier keys (Shift, Option) as needed for the current
 * and next key.
 *
 * @param flags    Pointer to event flags to update.
 * @param mods     Modifier mask for the current key.
 * @param modsNext Modifier mask for the next key.
 * @param keyDown  True to press, false to release.
 * @return         Number of modifier actions performed.
 */
int SetMods(CGEventFlags *flags, UInt32 mods, UInt32 modsNext, bool keyDown) {
  int counter = 0;

  for (int i = 0; i < 2; i++) {
    ModifierSet m = standardMods[i];

    if ((mods & m.mask) != 0) // mod mask has bits present
    {
      *flags |= m.flag; // set flags

      if (!keyDown) // key up
      {
        if ((modsNext & m.mask) ==
            0) // next key's mod mask does not have bits present
        {
          KeyRelease(m.key, 0);
          counter++;
        }
      } else {
        if (!KeyIsDown(m.key)) // key is up
        {
          KeyPress(m.key, *flags);
          counter++;
        }
      }
    }
  }

  snprintf(LastErrorMessage, sizeof(LastErrorMessage),
           "Unable to perform SetMods with flags=%llu mods=%x", *flags, mods);

  return counter;
}

/**
 *
 *
 * Types a string by simulating key presses/releases for each character,
 * handling modifiers, delays, and tab-to-space conversion.
 *
 * @param str         The string to type.
 * @param modPressDur Delay after pressing modifiers (microseconds).
 * @param keyPressDur Delay to hold each key down (microseconds).
 * @param keyDelay    Delay between key presses (microseconds).
 * @param tabSize     Number of spaces to substitute for a tab character.
 * @return            1 on success, 0 on failure.
 */
int TypeStr(const char *str, int modPressDur, int keyPressDur, int keyDelay,
            int tabSize) {
  int len = strlen(str);
  int last = len - 1;
  int errCount = 0;
  KeyboardInfo kbInfo = GetKeyboardInfo();
  KeyMapping current, next;
  usleep(keyDelay);

  // loop through characters in string
  for (size_t i = 0; i < len; i++) {
    UniChar c = str[i];
    CGEventFlags flags = 0; // reset flags

    if (i == 0)
      current = CharToVKAndMods(c, kbInfo);

    if (i < last)
      next = CharToVKAndMods(str[i + 1], kbInfo);
    else if (i == last)
      next = (KeyMapping){0};

    // press modifiers
    if (SetMods(&flags, current.mods, 0, true))
      usleep(modPressDur);

    // default taps
    int numTaps = 1;

    // to sub spaces for tabs
    if (c == '\t' && tabSize > 0) {
      current.vk = kVK_Space;
      numTaps = tabSize;
    }

    // loop through press + release
    for (int j = 0; j < numTaps; j++)
      errCount += KeyPressAndRelease(current.vk, flags, keyPressDur);

    // release or hold modifiers for next char
    SetMods(&flags, current.mods, next.mods, false);

    if (i < last) {
      current = next;
      usleep(keyDelay);
    }
  }

  return errCount ? 0 : 1;
}
