package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <exception>
// #include <fmt/format.h>
// #include <string>
// #include <string_view>
// namespace OpenLoco::Exception
// namespace Detail
// // TODO: Make this consteval for C++20
// constexpr std::string_view sanitizePath(std::string_view path)
// constexpr std::string_view projectPath = OPENLOCO_PROJECT_PATH;
// // Removes also the first slash.
// return path.substr(projectPath.size() + 1);
// #else
// orphan member: return path;
// // Something like std::source_location except this works prior to C++20.
type SourceLocation struct {
// std::string _function;
// std::string _file;
	Line int
	// method: explicit SourceLocation(std::string_view func = __builtin_FUNCTION(), std::string_view file = Detail::sanitizePath(__builtin_FILE()), int line = __builtin_LINE())
// : _function(func)
// , _file(file)
// , _line(line)
}
// const std::string& file() const noexcept
// orphan member: return _file;
// const std::string& function() const noexcept
// orphan member: return _function;
// func Line() int
// orphan member: return _line;
// namespace Detail
// template<typename TExceptionTag>
type ExceptionBase struct {
	Location SourceLocation
// std::string _message;
func (e *ExceptionBase) ExceptionBase(SourceLocation{} SourceLocation location =) explicit {
	// : _location{ location }
	_message = fmt.format("Exception thrown at '{}' - {}:{}", _location.function(), _location.file(), _location.line())
	}
	// explicit ExceptionBase(const std::string& message, const SourceLocation& location = SourceLocation{})
	// : _location{ location }
	_message = fmt.format("Exception '{}', thrown at '{}' - {}:{}", message, _location.function(), _location.file(), _location.line())
	}
	// const char* what() const noexcept override
	return _message.c_str()
	}
}
}
type RuntimeError = any /* Detail::ExceptionBase<struct RuntimeErrorTag> */ 
type InvalidArgument = any /* Detail::ExceptionBase<struct InvalidArgumentTag> */ 
type NotImplemented = any /* Detail::ExceptionBase<struct NotImplementedTag> */ 
type InvalidOperation = any /* Detail::ExceptionBase<struct InvalidOperationTag> */ 
type BadAllocation = any /* Detail::ExceptionBase<struct BadAllocTag> */ 
type OutOfRange = any /* Detail::ExceptionBase<struct OutOfRangeTag> */ 
type OverflowError = any /* Detail::ExceptionBase<struct OverflowErrorTag> */ 
