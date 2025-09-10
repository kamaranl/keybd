#pragma once
#ifndef KEY_EV_H
#define KEY_EV_H

#if !defined(IS_DARWIN) && defined(__APPLE__) && defined(__MACH__)
#define IS_DARWIN 1
#endif

#if !defined(IS_WINDOWS) && (defined(WIN32) || defined(_WIN32) || \
                             defined(__WIN32__) || defined(__WINDOWS__))
#define IS_WINDOWS 1
#endif

#if IS_DARWIN
#include <ApplicationServices/ApplicationServices.h>
#include <stdbool.h>

void SendKey(int key, bool shift);
#endif

#endif
