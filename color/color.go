// Package color provides color output utilities using charmbracelet/colorprofile.
package color

// NOTE:
// This color package was generated with Mistral devstral-2 AI,
// as unfortunately I was feeling lazy.

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// Writer wraps an io.Writer and carries color configuration.
type Writer struct {
	writer     io.Writer
	noColor    bool
	errHandler func(error) // Optional error handler for write failures
}

// NewWriter creates a new Writer that wraps the given writer.
func NewWriter(w io.Writer, noColor bool) *Writer {
	return &Writer{
		writer:     w,
		noColor:    noColor,
		errHandler: nil, // Default: ignore errors (maintains backward compatibility)
	}
}

// NewWriterWithErrorHandler creates a new Writer with an error handler.
// Example usage:
//
//	writer := color.NewWriterWithErrorHandler(os.Stdout, false, func(err error) {
//		log.Printf("Write error: %v", err)
//	})
func NewWriterWithErrorHandler(w io.Writer, noColor bool, handler func(error)) *Writer {
	return &Writer{
		writer:     w,
		noColor:    noColor,
		errHandler: handler,
	}
}

// SetErrorHandler sets the error handler for this Writer.
func (w *Writer) SetErrorHandler(handler func(error)) {
	w.errHandler = handler
}

// Writer returns the underlying io.Writer.
func (w *Writer) Writer() io.Writer {
	return w.writer
}

// IsTerminal checks if the writer is a terminal that supports colors.
func (w *Writer) IsTerminal() bool {
	// Check for forced no-color mode
	if w.noColor || os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if the writer is a file (os.Stdout, os.Stderr, or a regular file)
	file, ok := w.writer.(*os.File)
	if !ok {
		return false
	}

	// Use colorprofile to detect if the terminal supports color
	profile := colorprofile.Detect(file, os.Environ())
	return profile != colorprofile.Unknown && profile != colorprofile.Ascii
}

// NewColorProfileWriter creates a new color profile writer that wraps the given writer.
func NewColorProfileWriter(w io.Writer) io.Writer {
	return colorprofile.NewWriter(w, os.Environ())
}

// PrintColor prints colored text to the writer.
func (w *Writer) PrintColor(text string, fg ansi.Attr) {
	if !w.IsTerminal() {
		w.write([]byte(text))
		return
	}
	style := ansi.NewStyle(fg)
	colored := style.String() + text + ansi.ResetStyle
	w.write([]byte(colored))
}

// PrintColorf prints formatted colored text to the writer.
func (w *Writer) PrintColorf(format string, fg ansi.Attr, args ...any) {
	text := fmt.Sprintf(format, args...)
	if !w.IsTerminal() {
		w.write([]byte(text))
		return
	}
	style := ansi.NewStyle(fg)
	colored := style.String() + text + ansi.ResetStyle
	w.write([]byte(colored))
}

// PrintColorBold prints bold colored text to the writer.
func (w *Writer) PrintColorBold(text string, fg ansi.Attr) {
	if !w.IsTerminal() {
		w.write([]byte(text))
		return
	}
	style := ansi.NewStyle(fg, ansi.AttrBold)
	colored := style.String() + text + ansi.ResetStyle
	w.write([]byte(colored))
}

// PrintColorBoldf prints formatted bold colored text to the writer.
func (w *Writer) PrintColorBoldf(format string, fg ansi.Attr, args ...any) {
	text := fmt.Sprintf(format, args...)
	if !w.IsTerminal() {
		w.write([]byte(text))
		return
	}
	style := ansi.NewStyle(fg, ansi.AttrBold)
	colored := style.String() + text + ansi.ResetStyle
	w.write([]byte(colored))
}

// Color attributes for different types of output.
const (
	ColorFileName = ansi.AttrGreenForegroundColor
	ColorPath     = ansi.AttrCyanForegroundColor
	ColorCount    = ansi.AttrYellowForegroundColor
	ColorDate     = ansi.AttrMagentaForegroundColor
	ColorError    = ansi.AttrRedForegroundColor
	ColorHeader   = ansi.AttrBlueForegroundColor
)

// write handles writing with optional error handling.
func (w *Writer) write(data []byte) {
	if _, err := w.writer.Write(data); err != nil {
		if w.errHandler != nil {
			w.errHandler(err)
		}
		// If no error handler, ignore the error (backward compatibility)
	}
}
