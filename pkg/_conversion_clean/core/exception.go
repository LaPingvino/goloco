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
// constexpr string_view sanitizePath(string_view path)
// constexpr string_view projectPath = OPENLOCO_PROJECT_PATH;
// // Removes also the first slash.
// return path.substr(projectPath.size() + 1);
// #else
// orphan member: return path;
// // Something like std::source_location except this works prior to C++20.
type SourceLocation struct {
// string _function;
// string _file;
	Line int
	// method: explicit SourceLocation(string_view func = __builtin_FUNCTION(), string_view file = Detail::sanitizePath(__builtin_FILE()), int line = __builtin_LINE())
// : _function(func)
// , _file(file)
// , _line(line)
}
// const string& file() const noexcept
// orphan member: return _file;
// const string& function() const noexcept
// orphan member: return _function;
// func Line() int
// orphan member: return _line;
// namespace Detail
// template<typename TExceptionTag>
type ExceptionBase struct {
	Location SourceLocation
// string _message;
func (e *ExceptionBase) ExceptionBase(SourceLocation{} SourceLocation location =) explicit {
	// : _location{ location }
// SKIPPED CONSTRUCTOR: 	_message = fmt.format("Exception thrown at '{}' - {}:{}", _location.function(), _location.file(), _location.line())
	// explicit ExceptionBase(const string& message, const SourceLocation& location = SourceLocation{})
	// : _location{ location }
// SKIPPED CONSTRUCTOR: 	_message = fmt.format("Exception '{}', thrown at '{}' - {}:{}", message, _location.function(), _location.file(), _location.line())
	// const char* what() const noexcept override
// SKIPPED CONSTRUCTOR: 	return _message.c_str()
// SKIPPED C++ SYNTAX: type RuntimeError = any /* Detail::ExceptionBase<struct RuntimeErrorTag> */ 
// SKIPPED C++ SYNTAX: type InvalidArgument = any /* Detail::ExceptionBase<struct InvalidArgumentTag> */ 
// SKIPPED C++ SYNTAX: type NotImplemented = any /* Detail::ExceptionBase<struct NotImplementedTag> */ 
// SKIPPED C++ SYNTAX: type InvalidOperation = any /* Detail::ExceptionBase<struct InvalidOperationTag> */ 
// SKIPPED C++ SYNTAX: type BadAllocation = any /* Detail::ExceptionBase<struct BadAllocTag> */ 
// SKIPPED C++ SYNTAX: type OutOfRange = any /* Detail::ExceptionBase<struct OutOfRangeTag> */ 
// SKIPPED C++ SYNTAX: type OverflowError = any /* Detail::ExceptionBase<struct OverflowErrorTag> */ 