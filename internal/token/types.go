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

var varTokens = make([]VarToken)

func (tok *Token) AsVar() VarToken {
   return varTokens[tok.Data]
}
