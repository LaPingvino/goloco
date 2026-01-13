package platform

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Crash.h"
// #include "Platform.h"
// #include <OpenLoco/Utility/String.hpp>
// #include <ShlObj.h>
// #include <client/windows/handler/exception_handler.h>
// namespace OpenLoco::CrashHandler
// static AppInfo _appInfo;
// [[maybe_unused]] static bool onCrash(
// const wchar_t* dumpPath, const wchar_t* miniDumpId,
// [[maybe_unused]] void* context,
// [[maybe_unused]] EXCEPTION_POINTERS* exinfo,
// [[maybe_unused]] MDRawAssertionInfo* assertion,
// bool succeeded)
// if (!succeeded)
// constexpr const char* dumpFailedMessage = "Failed to create the dump. Please file an issue with OpenLoco on GitHub and "
// "provide latest save, and provide "
// "information about what you did before the crash occurred.";
// printf("%s\n", dumpFailedMessage);
// MessageBoxA(nullptr, dumpFailedMessage, "OpenLoco", MB_OK | MB_ICONERROR);
// orphan member: return succeeded;
// wchar_t dumpFilePath[MAX_PATH];
// swprintf_s(dumpFilePath, std::size(dumpFilePath), L"%s\\%s.dmp", dumpPath, miniDumpId);
// std::wstring version = Utility::toUtf16(_appInfo.version);
// wchar_t dumpFilePathNew[MAX_PATH];
// swprintf_s(
// dumpFilePathNew, std::size(dumpFilePathNew), L"%s\\%s(%s).dmp", dumpPath, miniDumpId, version.c_str());
// // Try to rename the files
// if (_wrename(dumpFilePath, dumpFilePathNew) == 0)
// wcscpy_s(dumpFilePath, dumpFilePathNew);
// // Log information to output
// wprintf(L"Dump Path: %s\n", dumpPath);
// wprintf(L"Dump File Path: %s\n", dumpFilePath);
// wprintf(L"Dump Id: %s\n", miniDumpId);
// wprintf(L"Version: %s\n", version.c_str());
// constexpr const wchar_t* MessageFormat = L"A crash has occurred and a dump was created at\n%s.\n\nPlease file an issue "
// L"with %S on GitHub and provide "
// L"the dump and most recently saved game there.\n\n\nVersion: %s\n\n";
// wchar_t message[MAX_PATH * 2];
// swprintf_s(message, MessageFormat, dumpFilePath, _appInfo.name.c_str(), version.c_str());
// MessageBoxW(nullptr, message, L"OpenLoco", MB_OK | MB_ICONERROR);
// HRESULT coInitializeResult = CoInitialize(nullptr);
// if (SUCCEEDED(coInitializeResult))
// LPITEMIDLIST pidl = ILCreateFromPathW(dumpPath);
// LPITEMIDLIST files[6];
// uint32_t numFiles = 0;
// files[numFiles++] = ILCreateFromPathW(dumpFilePath);
// if (pidl != nullptr)
// SHOpenFolderAndSelectItems(pidl, numFiles, (LPCITEMIDLIST*)files, 0);
// ILFree(pidl);
// for (uint32_t i = 0; i < numFiles; i++)
// ILFree(files[i]);
// CoUninitialize();
// orphan member: return succeeded;
// [[maybe_unused]] static std::wstring getDumpDirectory()
// auto crashDir = Platform::getUserDirectory() / "crashes";
// if (!fs::exists(crashDir))
// fs::create_directories(crashDir);
// return crashDir.wstring();
// func Init(appInfo [[maybe_unused]] AppInfo) Handle
// _appInfo = appInfo;
// const auto pipeName = Utility::toUtf16(appInfo.name + "_breakpad");
// // Path must exist and be RW!
// auto exHandler = new google_breakpad::ExceptionHandler(
// getDumpDirectory(), 0, onCrash, 0, google_breakpad::ExceptionHandler::HANDLER_ALL, MiniDumpWithDataSegs, pipeName.c_str(), 0);
// orphan member: return exHandler;
// #else
// orphan member: return nullptr;
// func Shutdown(exHandler [[maybe_unused]] Handle) 
// if (exHandler == nullptr)
// return;
// delete static_cast<google_breakpad::ExceptionHandler*>(exHandler);
