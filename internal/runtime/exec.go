package runtime

import (
	// "fmt"
   "strings"
   "unicode"
   "unicode/utf8"
	xxisToken "github.com/Greccl/xxis/internal/token"
)

func Exec(tok *xxisToken.Token, ctx *ThreadContext) error {
	if tok == nil {
		return nil
	}

	switch tok.Typ {
	case 'P':
		return execProgram(tok, ctx)
	case 'F':
		return execFunction(tok, ctx)
	case 'B':
		for _, child := range tok.Toks {
			if err := Exec(child, ctx); err != nil {
				return err
			}
		}
	case 'K':
		return execKeyword(tok, ctx)
	case 'C', 'S':
		return execCommand(tok, ctx)
	case '&', '/':
		return nil
	}
	return nil
}

func Expand(tok *Token, ctx *ThreadContext) []string {
   var args []string
   var current strings.Builder

   flush := func(force bool) {
      if current.Len() > 0 || force {
         args = append(args, current.String())
         current.Reset()
      }
   }

   pushBuffer := func(buf []rune, split bool) {
      for _, r := range buf {
         if split && unicode.IsSpace(r) {
            flush(true)
         } else {
            current.WriteRune(r)
         }
      }
   }

   pushString := func(str string, split bool) {
      start := 0
      for i, r := range str {
         if split && unicode.IsSpace(r) {
            current.WriteString(str[start:i])
            flush(true)
            start = i + utf8.RuneLen(r)
         }
      }
      if start < len(str) {
         current.WriteString(str[start:])
      }
   }

   for _, t := range tok.Toks {
      switch t.Typ {
      case 'T':
         pushBuffer(t.Buf, true)
      case 'Q':
         str := ExpandQuoted(t, ctx)
         pushString(str, false)
      case 'S':
         // TODO
         pushString("<cmd>", true)
      case 'V':
         str := ctx.ExpandVar(string(t.Buf))
         pushString(str, true)
      }
   }
   flush(true)

   return args
}

func ExpandQuoted(tok *Token, ctx *ThreadContext) string {
   var result strings.Builder

   pushBuffer := func(buf []rune) {
      result.WriteString(string(buf))
   }

   pushString := func(str string) {
      result.WriteString(str)
   }

   for _, t := range tok.Toks {
      switch t.Typ {
      case 'T':
         pushBuffer(t.Buf)
      case 'Q':
         // TODO
         panic("quote inside quote?")
      case 'S':
         // TODO
         pushString("<cmd>")
      case 'V':
         v := ctx.ExpandVar(string(t.Buf))
         pushString(v)
      }
   }

   return result.String()
}

func execCommand(tok *Token, ctx *ThreadContext) error {
   return nil
}

func execProgram(tok *Token, ctx *ThreadContext) error {
   /*
	for i, child := range tok.Toks {
		if i == 0 {
			continue
		}
		name := functionName(child)
		if name != "" {
			ctx.Module.Funcs[name] = child
		}
	}
	if len(tok.Toks) == 0 {
		return nil
	}
	main := tok.Toks[0]
	if main == nil || len(main.Toks) < 2 {
		return nil
	}
	return Exec(main.Toks[1], ctx)
	*/
	return nil
}

func execFunction(tok *Token, ctx *ThreadContext) error {
	if len(tok.Toks) < 2 {
		return nil
	}
	return Exec(tok.Toks[1], ctx)
}

func execKeyword(tok *xxisToken.Token, ctx *ThreadContext) error {
   /*
	if len(tok.Buf) == 0 {
		return nil
	}

	switch tok.Buf[0] {
	case xxisToken.IMPORT:
		if len(tok.Toks) != 2 {
			return fmt.Errorf("invalid import token")
		}
		imported, err := ctx.Global.GetOrImport(tokenText(tok.Toks[0]))
		if err != nil {
			return err
		}
		return ctx.SetVar(xxisToken.VarScopeLocal, tokenText(tok.Toks[1]), ModuleValue{Module: imported})
	case xxisToken.SOURCE:
		if len(tok.Toks) != 1 {
			return fmt.Errorf("invalid source token")
		}
		return ctx.Global.Source(ctx.Module, tokenText(tok.Toks[0]))
	case xxisToken.IF:
		return nil
	case xxisToken.VAR:
		return execVar(tok, ctx)
	}
	*/
	return nil
}

func execVar(tok *xxisToken.Token, ctx *ThreadContext) error {
   /*
	if len(tok.Buf) < 4 || len(tok.Toks) == 0 {
		return fmt.Errorf("invalid var token")
	}
	name := tokenText(tok.Toks[0])
	value := ""
	if len(tok.Toks) > 1 && tok.Toks[1] != nil {
		value = tokenText(tok.Toks[1])
	}
	return ctx.ApplyVar(tok.Buf[1], name, tok.Buf[3], value)
	*/
	return nil
}
