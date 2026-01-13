package utility

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "String.hpp"
// #include <windows.h>
// #undef max
// #else
// #include <codecvt>
// #include <errno.h>
// #include <iconv.h>
// #include <iostream>
// #include <locale>
// #include <limits>
// namespace OpenLoco::Utility
// static std::string_view skipDigits(std::string_view s)
// while (!s.empty() && (s[0] == '.' || s[0] == ',' || isdigit(s[0])))
// s = s.substr(1);
// orphan member: return s;
// static std::pair<int32_t, std::string_view> parseNextNumber(std::string_view s)
// int32_t num = 0;
// while (!s.empty())
// if (s[0] != '.' && s[0] != ',')
// if (!isdigit(s[0]))
// break;
// auto newResult = num * 10;
// if (newResult / 10 != num)
// return { std::numeric_limits<int32_t>::max(), skipDigits(s) };
// num = newResult + (s[0] - '0');
// if (num < newResult)
// return { std::numeric_limits<int32_t>::max(), skipDigits(s) };
// s = s.substr(1);
// return { num, s };
// // Case-insensitive logical string comparison function
// func Strlogicalcmp(s1 string, s2 string) int32
// for (;;)
// if (s2.empty())
// return !s1.empty();
// func If(s1.empty() else
// return -1;
// func If(!(isdigit(s1[0]) else
// if (toupper(s1[0]) != toupper(s2[0]))
// func Toupper(s1[0]) return
// else
// s1 = s1.substr(1);
// s2 = s2.substr(1);
// else
// // TODO: Replace with std::from_chars when all compilers support std::from_chars
// auto [n1, s1n] = parseNextNumber(s1);
// auto [n2, s2n] = parseNextNumber(s2);
// if (n1 > n2)
// orphan member: return 1;
// func If(n2 any /* n1 < */ ) else
// return -1;
// s1 = s1n;
// s2 = s2n;
// std::string toUtf8(const std::wstring_view& src)
// int srcLen = static_cast<int>(src.size());
// int sizeReq = WideCharToMultiByte(CP_UTF8, 0, src.data(), srcLen, nullptr, 0, nullptr, nullptr);
// auto result = std::string(sizeReq, 0);
// WideCharToMultiByte(CP_UTF8, 0, src.data(), srcLen, result.data(), sizeReq, nullptr, nullptr);
// orphan member: return result;
// #else
type Convert_typeX = any /* std::codecvt_utf8<wchar_t> */ 
// std::wstring_convert<convert_typeX, wchar_t> converterX;
// std::string out = converterX.to_bytes(src.data());
// std::cout << __PRETTY_FUNCTION__ << " " << out << "\n";
// orphan member: return out;
// std::wstring toUtf16([[maybe_unused]] const std::string_view& src)
// int srcLen = static_cast<int>(src.size());
// int sizeReq = MultiByteToWideChar(CP_ACP, 0, src.data(), srcLen, nullptr, 0);
// auto result = std::wstring(sizeReq, 0);
// MultiByteToWideChar(CP_ACP, 0, src.data(), srcLen, result.data(), sizeReq);
// orphan member: return result;
// #else
// perror("STUB!\n");
// return std::to_wstring(0);
