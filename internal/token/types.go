package token

import (
   "sync"
)

//
// Text
//

type TextToken struct {
   Text string
}

var textTokens = make([]*TextToken)
var textMu sync.Mutex

func (tok *Token) MakeText() *TextToken {
   data := &TextToken{}
   textMu.Lock()
   textTokens = append(textTokens, data)
   tok.Data = len(textTokens) - 1
   textMu.Unlock()
   return data
}

func (tok *Token) AsText() *TextToken {
   return textTokens[tok.Data]
}

//
// Var
//

type VarToken struct {
   Name string
   Scope rune
   Type rune
   Op rune
}

var varTokens = make([]*VarToken)
var varMu sync.Mutex

func (tok *Token) MakeVar() *VarToken {
   data := &VarToken{}
   varMu.Lock()
   varTokens = append(varTokens, data)
   tok.Data = len(varTokens) - 1
   varMu.Unlock()
   return data
}

func (tok *Token) AsVar() *VarToken {
   return varTokens[tok.Data]
}

const (
	VarScopeLocal     rune = 'l'
	VarScopeShared    rune = 's'
	VarScopeGlobal    rune = 'g'
	VarScopeUniversal rune = 'u'
	VarTypeString     rune = 's'
	VarTypeFloat      rune = 'n'
	VarTypeProcess    rune = 'j'
	VarTypePath       rune = 'p'
	VarTypeList       rune = 'l'
	VarTypeDict       rune = 'd'
	VarOpAssign       rune = '='
	VarOpAppend       rune = '<'
	VarOpPrepend      rune = '>'
	VarOpDelete       rune = 'd'
)

func VarScopeName(scope rune) string {
	switch scope {
	case VarScopeLocal:
		return "local"
	case VarScopeShared:
		return "shared"
	case VarScopeGlobal:
		return "global"
	case VarScopeUniversal:
		return "universal"
	}
	return "unknown"
}

func VarTypeName(kind rune) string {
	switch kind {
	case VarTypeString:
		return "string"
	case VarTypeFloat:
		return "float"
	case VarTypeProcess:
		return "process"
	case VarTypePath:
		return "path"
	case VarTypeList:
		return "list"
	case VarTypeDict:
		return "dict"
	}
	return "unknown"
}

func VarOpName(op rune) string {
	switch op {
	case VarOpAssign:
		return "assign"
	case VarOpAppend:
		return "append"
	case VarOpPrepend:
		return "prepend"
	case VarOpDelete:
		return "delete"
	}
	return "unknown"
}

//
// Keyword if
//

type KwIfToken struct {
   Subtype rune
   Else bool
}

var kwIfTokens = make([]*KwIfToken)
var kwIfMu sync.Mutex

func (tok *Token) MakeKwIf() *KwIfToken {
   data := &KwIfToken{}
   kwIfMu.Lock()
   kwIfTokens = append(kwIfTokens, data)
   tok.Data = len(kwIfTokens) - 1
   kwIfMu.Unlock()
   return data
}

func (tok *Token) AsKeywordIf() *KwIfToken {
   return kwIfTokens[tok.Data]
}

const (
	IfCondExpr rune = iota
	IfCondSuccess
	IfCondFailure
)

func IfCondName(mode rune) string {
	switch mode {
	case IfCondExpr:
		return "expr"
	case IfCondSuccess:
		return "success"
	case IfCondFailure:
		return "failure"
	}
	return "unknown"
}

//
// Keyword import
//

type KwImportToken struct {
   Path string
   Merge bool
}

var kwImportTokens = make([]*KwImportToken)
var kwImportMu sync.Mutex

func (tok *Token) MakeKwImport() *KwImportToken {
   data := &KwImportToken{}
   kwImportMu.Lock()
   kwImportTokens = append(kwImportTokens, data)
   tok.Data = len(kwImportTokens) - 1
   kwImportMu.Unlock()
   return data
}

func (tok *Token) AsKeywordImport() *KwImportToken {
   return kwImportTokens[tok.Data]
}

//
// Function
//

type FunctionToken struct {
   Name string
}

var functionTokens = make([]*functionToken)
var functionMu sync.Mutex

func (tok *Token) MakeFunction() *FunctionToken {
   data := &FunctionToken{}
   functionMu.Lock()
   functionTokens = append(functionTokens, data)
   tok.Data = len(functionTokens) - 1
   functionMu.Unlock()
   return data
}

func (tok *Token) AsFunction() *FunctionToken {
   return functionTokens[tok.Data]
}

//
// Command
//

type CommandToken struct {
   TODO bool
}

var commandTokens = make([]*CommandToken)
var commandMu sync.Mutex

func (tok *Token) MakeCommand() *CommandToken {
   data := &CommandToken{}
   kwImportMu.Lock()
   kwImportTokens = append(kwImportTokens, data)
   tok.Data = len(kwImportTokens) - 1
   kwImportMu.Unlock()
   return data
}

func (tok *Token) AsKeywordImport() *KwImportToken {
   return kwImportTokens[tok.Data]
}

//
// Extra
//

const (
	IF rune = iota
	IFZ
	IFN
	FOR
	VAR
	IMPORT
	SOURCE
)

var KEYWORDS = []string{
	"if",
	"ifz",
	"ifn",
	"for",
	"var",
	"import",
	"source",
}