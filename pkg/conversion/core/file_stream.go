package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "FileStream.h"
// #include "Exception.hpp"
// #include <algorithm>
// #include <stdio.h>
// namespace OpenLoco
// static FILE* fileOpen(const std::filesystem::path& path, StreamMode mode)
// if (mode == StreamMode::none)
// throw Exception::InvalidArgument("Invalid mode argument");
// FILE* fs;
// _wfopen_s(&fs, path.wstring().c_str(), mode == StreamMode::read ? L"rb" : L"wb");
// orphan member: return fs;
// #else
// func Fopen(path.u8string() return
// func FileTell(fs FILE) int
// func Ftelli64_nolock(fs) return
// #else
// func Ftello(fs) return
// func FileSeek(fs FILE, offset int, origin int) 
// _fseeki64_nolock(fs, offset, origin);
// #else
// fseeko(fs, offset, origin);
// func ReadFile(buffer , len int, fs FILE) int
// func Fread_nolock(buffer, 1, len, fs) return
// #else
// return std::fread(buffer, 1, len, fs);
// func WriteFile(buffer , len int, fs FILE) int
// func Fwrite_nolock(buffer, 1, len, fs) return
// #else
// return std::fwrite(buffer, 1, len, fs);
// func FileClose(fs FILE) 
// _fclose_nolock(fs);
// #else
// fclose(fs);
// func GetFileLength(fs FILE) int
// const auto current = fileTell(fs);
// fileSeek(fs, 0, SEEK_END);
// const auto length = fileTell(fs);
// fileSeek(fs, current, SEEK_SET);
// orphan member: return length;
// FileStream::FileStream(const std::filesystem::path& path, StreamMode mode)
// if (!open(path, mode))
// // TODO: Make this work like fstream which is not throwing for failing to open the file.
// throw Exception::RuntimeError("Failed to open '" + path.u8string() + "' for writing");
// FileStream::~FileStream()
// close();
// bool FileStream::open(const std::filesystem::path& path, StreamMode mode)
// close();
// _file = fileOpen(path, mode);
// if (_file == nullptr)
// orphan member: return false;
// // Increase the buffer size to 1MiB.
// std::setvbuf(_file, nullptr, _IOFBF, 1024 * 1024);
// // Get the length if we are reading an existing file.
// if (mode == StreamMode::read)
// _length = getFileLength(_file);
// _offset = 0;
// _mode = mode;
// orphan member: return true;
// bool FileStream::isOpen() const noexcept
// return _file != nullptr;
// void FileStream::close()
// if (_file == nullptr)
// return;
// _mode = StreamMode::none;
// _length = 0;
// _offset = 0;
// fileClose(_file);
// _file = nullptr;
// StreamMode FileStream::getMode() const noexcept
// orphan member: return _mode;
// size_t FileStream::getLength() const noexcept
// orphan member: return _length;
// size_t FileStream::getPosition() const noexcept
// orphan member: return _offset;
// void FileStream::setPosition(size_t position)
// if (_mode == StreamMode::none)
// throw Exception::InvalidOperation("Invalid mode");
// position = std::min(_length, static_cast<size_t>(position));
// fileSeek(_file, position, SEEK_SET);
// _offset = position;
// void FileStream::read(void* buffer, size_t len)
// if (_mode != StreamMode::read)
// throw Exception::InvalidOperation("Can not read");
// const auto bytesRead = readFile(buffer, len, _file);
// if (bytesRead != len)
// throw Exception::RuntimeError("Failed to read data");
// _offset += bytesRead;
// void FileStream::write(const void* buffer, size_t len)
// if (_mode != StreamMode::write)
// throw Exception::InvalidOperation("Can not write");
// if (len == 0)
// return;
// const auto bytesWriten = writeFile(buffer, len, _file);
// if (bytesWriten != len)
// throw Exception::RuntimeError("Failed to write data");
// _offset += bytesWriten;
// _length = std::max(_length, _offset);
