package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Gfx.h"
// #include <cstdint>
// #include <string>
// namespace OpenLoco::Ui::TextInput
const TextboxPadding int16 = 4

type InputSession struct {
	// string buffer;    // 0x011369A0
	CursorPosition int
	XOffset        int16
	CursorFrame    uint8
	InputLenLimit  uint32
	// InputSession() = default;
	// InputSession(const string initialString, uint32 inputSize)
	// buffer = initialString;
	// cursorPosition = buffer.length();
	// cursorFrame = 0;
	// xOffset = 0;
	// inputLenLimit = inputSize;
}

// func HandleInput(charCode uint32, keyCode uint32) bool
// func NeedsReoffsetting(containerWidth int16) bool
// func CalculateTextOffset(containerWidth int16)
// func ClearInput()
// func SanitizeInput()
