package platform

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #undef _CRT_SECURE_NO_WARNINGS
const _CRT_SECURE_NO_WARNINGS = 1
// #include "Platform.h"
// #include <cstdlib>
// #include <io.h>
// #include <iostream>
// #include <tuple>
// // clang-format off
// // Windows headers are quite sensitive to the include order.
// #include <shlobj.h>
// #include <windows.h>
// #include <mmsystem.h>
// #include <shellapi.h>
// // clang-format on
// #include <OpenLoco/Utility/String.hpp>
// namespace OpenLoco::Platform
var SingleInstanceMutexName = L"OpenLocoMutex" // auto
// func GetTime() uint32
// func TimeGetTime() return
// fs::path getUserDirectory()
// auto result = fs::path{};
// wchar_t path[MAX_PATH];
// if (SUCCEEDED(SHGetFolderPathW(nullptr, CSIDL_APPDATA | CSIDL_FLAG_CREATE, nullptr, 0, path)))
// result = fs::path(path) / "OpenLoco";
// orphan member: return result;
// static std::wstring SHGetPathFromIDListLongPath(LPCITEMIDLIST pidl)
// std::wstring pszPath(MAX_PATH, 0);
// while (!SHGetPathFromIDListW(pidl, &pszPath[0]))
// if (pszPath.size() >= SHRT_MAX)
// // Clearly not succeeding at all, bail
// return std::wstring();
// pszPath.resize(pszPath.size() * 2);
// auto nullBytePos = pszPath.find(L'\0');
// if (nullBytePos != std::wstring::npos)
// pszPath.resize(nullBytePos);
// orphan member: return pszPath;
// fs::path promptDirectory(const std::string& title, void* hwnd)
// fs::path result;
// // Initialize COM and get a pointer to the shell memory allocator
// orphan member: LPMALLOC lpMalloc;
// if (SUCCEEDED(CoInitializeEx(nullptr, COINIT_APARTMENTTHREADED)) && SUCCEEDED(SHGetMalloc(&lpMalloc)))
// auto titleW = Utility::toUtf16(title);
// orphan member: BROWSEINFOW bi{};
// bi.lpszTitle = titleW.c_str();
// bi.ulFlags = BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE | BIF_NONEWFOLDERBUTTON;
// LPITEMIDLIST pidl = SHBrowseForFolderW(&bi);
// if (pidl != nullptr)
// result = fs::path(SHGetPathFromIDListLongPath(pidl).c_str());
// CoTaskMemFree(pidl);
// else
// std::cerr << "Error opening directory browse window";
// CoUninitialize();
// // SHBrowseForFolderW might minimize the main window,
// // so make sure that it's visible again.
// if (hwnd != nullptr)
// ShowWindow(static_cast<HWND>(hwnd), SW_RESTORE);
// orphan member: return result;
// static fs::path WIN32_GetModuleFileNameW(HMODULE hModule)
// uint32_t wExePathCapacity = MAX_PATH;
// std::unique_ptr<wchar_t[]> wExePath;
// orphan member: uint32_t size;
// do
// wExePathCapacity *= 2;
// wExePath = std::make_unique<wchar_t[]>(wExePathCapacity);
// size = GetModuleFileNameW(hModule, wExePath.get(), wExePathCapacity);
// } while (size >= wExePathCapacity);
// return fs::path(wExePath.get());
// fs::path getCurrentExecutablePath()
// func WIN32_GetModuleFileNameW(nullptr) return
// std::vector<fs::path> getDrives()
// char drive[4] = { 'A', ':', '\\', '\0' };
// std::vector<fs::path> drives;
// auto driveMask = GetLogicalDrives();
// for (auto i = 0; i < 26; i++)
// if (driveMask & (1 << i))
// drive[0] = 'A' + i;
// std::error_code ec;
// // Remove unaccessable drives (if error it is unaccessable)
// if (fs::is_directory(drive, ec))
// drives.push_back(drive);
// orphan member: return drives;
// std::string getEnvironmentVariable(const std::string& name)
// auto result = std::getenv(name.c_str());
// return result == nullptr ? std::string() : result;
// func IsRunningInWine() bool
// HMODULE ntdllMod = GetModuleHandleW(L"ntdll.dll");
// if (ntdllMod && GetProcAddress(ntdllMod, "wine_get_version"))
// orphan member: return true;
// orphan member: return false;
// func IsStdOutRedirected() bool
// // isatty returns a nonzero value if the descriptor is associated with a character device. Otherwise, isatty returns 0.
// func Isatty(_fileno(stdout) return
// func HasTerminalVT100SupportImpl() bool
// // See https://no-color.org/ for reference.
// const auto noColorEnvVar = getEnvironmentVariable("NO_COLOR");
// if (!noColorEnvVar.empty())
// orphan member: return false;
// const auto ntdllHandle = GetModuleHandleW(L"ntdll.dll");
// if (ntdllHandle == nullptr)
// orphan member: return false;
type RtlGetVersionFn = LONG(WINAPI*)(PRTL_OSVERSIONINFOW)
// auto RtlGetVersionFp = reinterpret_cast<RtlGetVersionFn>(GetProcAddress(ntdllHandle, "RtlGetVersion"));
// if (RtlGetVersionFp == nullptr)
// orphan member: return false;
// orphan member: RTL_OSVERSIONINFOW info{};
// info.dwOSVersionInfoSize = sizeof(info);
// if (RtlGetVersionFp(&info) != 0)
// orphan member: return false;
// // VT100 support was first introduced in 10.0.10586
// if (std::tie(info.dwMajorVersion, info.dwMinorVersion, info.dwBuildNumber) >= std::make_tuple(10U, 0U, 10586U))
// orphan member: return true;
// orphan member: return false;
// func HasTerminalVT100Support() bool
// static bool hasVT100Support = hasTerminalVT100SupportImpl();
// orphan member: return hasVT100Support;
// func EnableVT100TerminalMode() bool
// if (!isStdOutRedirected())
// orphan member: return false;
// if (!hasTerminalVT100Support())
// orphan member: return false;
// HANDLE hOut = GetStdHandle(STD_OUTPUT_HANDLE);
// if (hOut == INVALID_HANDLE_VALUE)
// orphan member: return false;
// DWORD dwMode = 0;
// if (!GetConsoleMode(hOut, &dwMode))
// orphan member: return false;
// dwMode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING;
// if (!SetConsoleMode(hOut, dwMode))
// orphan member: return false;
// orphan member: return true;
// std::vector<std::string> getCmdLineVector(int, const char**)
// // Discard the input as it will either be ASCII or not available
// // We need to get the Wide command line and convert it to utf8 command line parameters
// const auto cmdline = GetCommandLineW();
// orphan member: int argc{};
// auto* argw = CommandLineToArgvW(cmdline, &argc);
// if (argw == nullptr)
// std::cerr << "CommandLineToArgvW failed to convert commandline";
// // We will continue but just ignore the command line args.
// argc = 0;
// // This will hold a utf8 string of the command line
// std::vector<std::string> argvStrs;
// argvStrs.resize(argc);
// for (auto i = 0; i < argc; ++i)
// int length = WideCharToMultiByte(CP_UTF8, 0, argw[i], -1, nullptr, 0, nullptr, nullptr);
// if (length == 0)
// // Sadly can't print what the argument is that is causing the issue as it needs to be in utf8...
// std::cerr << "Failed to get cmdline argument utf8 length.";
// continue;
// // When resizing a std::string the null termination is handled implicitly. i.e. resize to true length not null terminated length.
// argvStrs[i].resize(length - 1);
// if (WideCharToMultiByte(CP_UTF8, 0, argw[i], -1, argvStrs[i].data(), length, nullptr, nullptr) == 0)
// std::cerr << "Failed to convert cmdline argument to utf8.";
// LocalFree(reinterpret_cast<HLOCAL>(argw));
// orphan member: return argvStrs;
// // 0x00407FFD
// func LockSingleInstance() bool
// // Check if operating system mutex exists
// HANDLE mutex = CreateMutexW(nullptr, FALSE, kSingleInstanceMutexName);
// if (mutex == nullptr)
// std::cerr << "unable to create mutex";
// orphan member: return true;
// func If(GetLastError() else
// // Already running
// CloseHandle(mutex);
// orphan member: return false;
// orphan member: return true;
