package compiler

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type CompileError struct {
	File     string
	Msg      string
	OffStart int
	OffEnd   int
	Line     int
	ColStart int
	ColEnd   int
	LineText string
	Cause    error
}

func (e *CompileError) Error() string {
	return formatCompileError(e)
}

func compileSource(src string, filename string) (prog *Program, err *CompileError) {
	defer func() {
		if r := recover(); r != nil {
			err = normalizeCompileError(r, src, filename)
			prog = nil
		}
	}()

	read := enumerate_string(src)
	// segments := get_segments(read)
	// metas := get_metaexpressions(segments)
	// ast := build_ast(metas)
	next := enumerate_tokens(read)
	ast := build_ast_from_tokens(next)

	com := New_Compiler()
	for _, tok := range ast.toks {
		com.process(tok)
	}
	com.finish()

	prog = buildProgram(com)
	return prog, nil
}

func buildProgram(com *Compiler) *Program {
	prog := &Program{}
	prog.funcs = make([]Function, 0, len(com.funcs))
	for _, f := range com.funcs {
		prog.funcs = append(prog.funcs, *f)
	}
	return prog
}

func normalizeCompileError(r interface{}, src string, filename string) *CompileError {
	var e *CompileError
	switch v := r.(type) {
	case *CompileError:
		e = v
	case CompileError:
		e = &v
	case error:
		e = &CompileError{Msg: v.Error(), Cause: v}
	default:
		e = &CompileError{Msg: fmt.Sprintf("%v", v)}
	}

	if e.File == "" {
		e.File = filename
	}
	fillCompileErrorSource(e, src)
	return e
}

func fillCompileErrorSource(e *CompileError, src string) {
	if e == nil || src == "" {
		return
	}

	if e.LineText != "" {
		return
	}

	if e.OffStart >= 0 {
		lineStart, lineEnd, lineNo, col := locateOffsetRune(src, e.OffStart)
		if e.Line == 0 {
			e.Line = lineNo
		}
		if e.ColStart == 0 {
			e.ColStart = col
		}
		if e.OffEnd > e.OffStart && e.ColEnd == 0 {
			_, _, _, colEnd := locateOffsetRune(src, e.OffEnd-1)
			e.ColEnd = colEnd
		} else if e.ColEnd == 0 {
			e.ColEnd = e.ColStart
		}
		e.LineText = strings.TrimRight(src[lineStart:lineEnd], "\r\n")
		return
	}

	if e.Line > 0 {
		lineStart, lineEnd := lineByNumber(src, e.Line)
		if lineStart >= 0 {
			e.LineText = strings.TrimRight(src[lineStart:lineEnd], "\r\n")
		}
	}
}

func locateOffsetRune(src string, off int) (lineStartByte int, lineEndByte int, lineNo int, col int) {
	if off < 0 {
		off = 0
	}

	runePos := 0
	lineStartByte = 0
	lineNo = 1
	col = 1
	for idx, r := range src {
		if runePos >= off {
			break
		}
		if r == '\n' {
			lineNo++
			lineStartByte = idx + utf8.RuneLen(r)
			col = 1
		} else {
			col++
		}
		runePos++
	}

	lineEndByte = strings.IndexByte(src[lineStartByte:], '\n')
	if lineEndByte == -1 {
		lineEndByte = len(src)
	} else {
		lineEndByte = lineStartByte + lineEndByte
	}
	return lineStartByte, lineEndByte, lineNo, col
}

func lineByNumber(src string, lineNo int) (lineStart int, lineEnd int) {
	if lineNo <= 0 {
		return -1, -1
	}

	curr := 1
	lineStart = 0
	for idx, r := range src {
		if curr == lineNo {
			break
		}
		if r == '\n' {
			curr++
			lineStart = idx + utf8.RuneLen(r)
		}
	}
	if curr != lineNo {
		return -1, -1
	}

	lineEnd = strings.IndexByte(src[lineStart:], '\n')
	if lineEnd == -1 {
		lineEnd = len(src)
	} else {
		lineEnd = lineStart + lineEnd
	}
	return lineStart, lineEnd
}

func formatCompileError(e *CompileError) string {
	if e == nil {
		return ""
	}

	msg := e.Msg
	if msg == "" && e.Cause != nil {
		msg = e.Cause.Error()
	}
	if msg == "" {
		msg = "compile error"
	}

	file := e.File
	if file == "" {
		file = "<source>"
	}

	if e.Line <= 0 {
		return fmt.Sprintf("%s: %s", file, msg)
	}

	colEnd := e.ColEnd
	if colEnd < e.ColStart {
		colEnd = e.ColStart
	}

	header := fmt.Sprintf("%s:%d:%d-%d: %s", file, e.Line, e.ColStart, colEnd, msg)
	if e.LineText == "" || e.ColStart <= 0 {
		return header
	}

	caretLen := colEnd - e.ColStart + 1
	if caretLen < 1 {
		caretLen = 1
	}
	caret := strings.Repeat("^", caretLen)
	pad := strings.Repeat(" ", e.ColStart-1)
	return fmt.Sprintf("%s\n  %s\n  %s%s", header, e.LineText, pad, caret)
}
