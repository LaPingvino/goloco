package platform

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Platform.h"
// #include <cstdlib>
// #include <cstring>
// #include <fcntl.h>
// #include <iostream>
// #include <pwd.h>
// #include <time.h>
// #include <linux/limits.h>
// #include <sys/types.h>
// #include <unistd.h>
// #include <sys/sysctl.h>
// #include <sys/types.h>
// #include <unistd.h>
// namespace OpenLoco::Platform
var SingleInstanceMutexName = "OpenLoco.lock" // auto
// func GetTime() uint32
type Timespec struct {
// clock_gettime(CLOCK_REALTIME, &spec);
// return (spec.tv_sec * 1000) + spec.tv_nsec / 1000000;
}
// std::vector<fs::path> getDrives()
// return {};
// std::string getEnvironmentVariable(const std::string& name)
// auto result = getenv(name.c_str());
// return result == nullptr ? std::string() : result;
// static fs::path getHomeDirectory()
// auto pw = getpwuid(getuid());
// if (pw != nullptr)
// return pw->pw_dir;
// else
// func GetEnvironmentVariable("HOME") return
// fs::path getUserDirectory()
// auto path = fs::path(getEnvironmentVariable("XDG_CONFIG_HOME"));
// if (path.empty())
// path = getHomeDirectory();
// if (path.empty())
// path = "/";
// else
// path = path / fs::path(".config");
// return path / fs::path("OpenLoco");
// fs::path getCurrentExecutablePath()
// char exePath[PATH_MAX] = { 0 };
// auto bytesRead = readlink("/proc/self/exe", exePath, sizeof(exePath));
// if (bytesRead == -1)
// fprintf(stderr, "failed to read /proc/self/exe");
// #elif defined(__FreeBSD__)
// const int32_t mib[] = { CTL_KERN, KERN_PROC, KERN_PROC_PATHNAME, -1 };
// auto exeLen = sizeof(exePath);
// if (sysctl(mib, 4, exePath, &exeLen, nullptr, 0) == -1)
// fprintf(stderr, "failed to get process path");
// #elif defined(__OpenBSD__)
// // There is no way to get the path name of a running executable.
// // If you are not using the port or package, you may have to change this line!
// strlcpy(exePath, "/usr/local/bin/", sizeof(exePath));
// #else
// #error "Platform does not support full path exe retrieval"
// orphan member: return exePath;
// fs::path promptDirectory([[maybe_unused]] const std::string& Title, [[maybe_unused]] void* hwnd)
// std::string input;
// std::cout << "Type your Locomotion path: ";
// std::cin >> input;
// orphan member: return input;
// func IsRunningInWine() bool
// orphan member: return false;
// func IsStdOutRedirected() bool
// // isatty returns a nonzero value if the descriptor is associated with a character device. Otherwise, isatty returns 0.
// func Isatty(fileno(stdout) return
// func HasTerminalVT100SupportImpl() bool
// // See https://no-color.org/ for reference.
// const auto noColorEnvVar = getEnvironmentVariable("NO_COLOR");
// if (!noColorEnvVar.empty())
// orphan member: return false;
// const auto termEnvVar = getEnvironmentVariable("TERM");
// if (termEnvVar.empty())
// orphan member: return false;
// return termEnvVar.compare("xterm") == 0
// || termEnvVar.compare("xterm-256color") == 0
// || termEnvVar.compare("rxvt-unicode-256color") == 0;
// func HasTerminalVT100Support() bool
// static bool hasVT100Support = hasTerminalVT100SupportImpl();
// orphan member: return hasVT100Support;
// func EnableVT100TerminalMode() bool
// if (!isStdOutRedirected())
// orphan member: return false;
// if (!hasTerminalVT100Support())
// orphan member: return false;
// orphan member: return true;
// std::vector<std::string> getCmdLineVector(int argc, const char** argv)
// std::vector<std::string> argvStrs;
// argvStrs.resize(argc);
// for (auto i = 0; i < argc; ++i)
// argvStrs[i] = argv[i];
// orphan member: return argvStrs;
// func LockSingleInstance() bool
// // We will never close this file manually. The operating system will
// // take care of that, because flock keeps the lock as long as the
// // file is open and closes it automatically on file close.
// // This is intentional.
// int32_t pidFile = open(kSingleInstanceMutexName, O_CREAT | O_RDWR, 0666);
// if (pidFile == -1)
// std::cerr << "Cannot open lock file for writing.";
// orphan member: return false;
type Flock struct {
// lock.l_start = 0;
// lock.l_len = 0;
// lock.l_type = F_WRLCK;
// lock.l_whence = SEEK_SET;
// if (fcntl(pidFile, F_SETLK, &lock) == -1)
// if (errno == EWOULDBLOCK)
// std::cerr << "Another OpenLoco session has been found running.";
	False return
}
// std::cerr << "flock returned an uncatched errno: " << errno;
// orphan member: return false;
// orphan member: return true;
