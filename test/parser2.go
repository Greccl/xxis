package main

import (
	"fmt"
	"os"
	"log"
   "bufio"
	// "path/filepath"
	// xxisCompiler "github.com/Greccl/xxis/internal/compiler"
	xxisParser "github.com/Greccl/xxis/internal/parser"
	// xxisVm "github.com/Greccl/xxis/internal/vm"
	xxisToken "github.com/Greccl/xxis/internal/token"
   "github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type Token = xxisToken.Token

type AstNode struct {
   tok *Token
   prefix string
   name string
}

func countNodes(tok *Token) int {
   if tok == nil { return 0 }
   n := len(tok.Toks)
   for _, t := range tok.Toks {
      n += countNodes(t)
   }
   return n
}

func New_AstView(ast *Token) []AstNode {
   nodes := make([]AstNode, 0, countNodes(ast))
   printer := func(t *Token, p, n string) {
      nodes = append(nodes, AstNode{tok:t, prefix:p, name:n})
   }
   ast.BuildNodes(printer)
   return nodes
}


func main() {

	// home, _ := os.UserHomeDir()
	// path := filepath.Join(home, "dev", "xxis", "test1.txt")
	read, getLine := xxisParser.Enumerate_file("test/src0.txt")

   file, err := os.Open("test/src0.txt")
   if err != nil {
      log.Fatal(err)
   }
   defer file.Close() // Ensure the file is closed

   scanner := bufio.NewScanner(file)

   lines := make([]string, 0)
   for scanner.Scan() {
      lines = append(lines, scanner.Text()) // .Text() returns the line as a string
   }

   if err := scanner.Err(); err != nil {
      log.Fatal(err)
   }

   file.Close()

   /*
   defer func() {
      if err := recover(); err != nil {
         switch x := err.(type) {
         case xxisParser.ParseError:
            fmt.Println("error detected")
            fmt.Println(x)
         }
      }
   }()
   */

	// src0 := "cmd 1\nif cmd 2\n   cmd '3' $(hola)!(mundo)\n   if cmd 4   :   exit\n   last in block\nelse\n   it works\nend\ncmd 5 && cmd '6 $var6' || cmd 7"
	// read := xxisParser.Enumerate_string(src0)

	next := xxisParser.Enumerate_tokens(read)
	ast := xxisParser.Build_ast_from_tokens(next)

	// ast.Dump()
   view := New_AstView(ast)


   
   s, err := tcell.NewScreen()
   if err != nil {
      fmt.Printf("%+v", err)
      os.Exit(1)
   }
   if err := s.Init(); err != nil {
   	fmt.Printf("%+v", err)
   	os.Exit(1)
   }

   // Set default text style
   defStyle := tcell.StyleDefault.Background(color.Reset).Foreground(color.Reset)
   s.SetStyle(defStyle)

   nodeStyle := tcell.StyleDefault.Background(color.Reset).Foreground(color.NewRGBColor(50, 250, 0))

   // Clear screen
   s.Clear()

   quit := func() {
      s.Fini()
      os.Exit(0)
   }

   // scrw, scrh := s.Size()

boxh := 12
inode := 0
node0 := 0
node1 := boxh - 1
base := boxh

drawAll := func() {
   s.SetStyle(defStyle)
   s.Clear()

   in := inode
   if node0 >= in {
      node0 = in - 1
   }
   if node1 <= in {
      node1 = in + 1
      node0 = node1 - boxh + 1
   }
   if node0 < 0 {
      node0 = 0
      node1 = node0 + boxh - 1
   }

   in = node0
   for i := 0; i < boxh; i++ {
      if in < len(view) {
         node := &view[in]
         // s.SetStyle(defStyle)
         l := 0
         for range node.prefix {
            l++
         }
         s.PutStrStyled(0, base+i, node.prefix, defStyle)
         if in == inode {
            s.PutStrStyled(l, base+i, node.name, nodeStyle)
         } else {
            s.PutStrStyled(l, base+i, node.name, defStyle)
         }
         in++
      } else {
         
      }
   }

   tok := view[inode].tok
   l1 := getLine(tok.Start) - 1
   l2 := getLine(tok.End) - 1
   r := l2 - l1 + 1
   l0 := 0
   if len(lines) <= boxh {
      l0 = 0
   } else if r >= boxh {
      l0 = l1
   } else {
      l0 = (boxh - r) / 2
   }

   for i:=0; i<boxh; i++ {
      l := l0+i
      if l < len(lines) {
         s.PutStr(0, i, fmt.Sprintf("%d", l))
         s.PutStr(2, i, "| ")
         if l >= l1 && l <= l2 {
            s.PutStrStyled(4, i, lines[l], nodeStyle)
         } else {
            s.PutStr(4, i, lines[l])
         }
      }
   }

   s.Sync()
}

   for {
      // Update screen
      s.Show()

      // Poll event (can be used in select statement as well)
      ev := <-s.EventQ()

      // Process event
      switch ev := ev.(type) {
      case *tcell.EventResize:
         // scrw, scrh = w.Size()
         drawAll()
      case *tcell.EventKey:
         switch ev.Key() {
         case tcell.KeyEscape, tcell.KeyCtrlC:
            quit()
         case tcell.KeyDown:
            if inode < len(view) - 1 {
               inode++
               drawAll()
            }
         case tcell.KeyUp:
            if inode > 0 {
               inode--
               drawAll()
            }
         }
          // ev.Str()
         // switch s {
         // case ""
         // }
      }
   }
	
}



func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	var width int
	for text != "" {
		text, width = s.Put(col, row, text, style)
		col += width
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
		if width == 0 {
			// incomplete grapheme at end of string
			break
		}
	}
}
