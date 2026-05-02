package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"unicode/utf8"
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
	tok      *Token
	prefix   string
	name     string
	parent   int
	children []int
}

type SourceLine struct {
	text  string
	start int
	end   int
}

func countNodes(tok *Token) int {
	if tok == nil {
		return 0
	}
	n := len(tok.Toks)
	for _, t := range tok.Toks {
		n += countNodes(t)
	}
	return n
}

func New_AstView(ast *Token) []AstNode {
	nodes := make([]AstNode, 0, countNodes(ast)+1)
	parents := make([]int, 0)
	printer := func(t *Token, p, n string) {
		depth := utf8.RuneCountInString(p) / 4
		if depth < len(parents) {
			parents = parents[:depth]
		}

		parent := -1
		if depth > 0 && depth-1 < len(parents) {
			parent = parents[depth-1]
		}

		index := len(nodes)
		nodes = append(nodes, AstNode{
			tok:    t,
			prefix: p,
			name:   n,
			parent: parent,
		})
		if parent >= 0 {
			nodes[parent].children = append(nodes[parent].children, index)
		}

		if depth == len(parents) {
			parents = append(parents, index)
		} else {
			parents[depth] = index
		}
	}
	ast.BuildNodes(printer)
	return nodes
}

func moveToSibling(nodes []AstNode, inode, delta int) int {
	if inode < 0 || inode >= len(nodes) {
		return inode
	}

	parent := nodes[inode].parent
	if parent < 0 {
		return inode
	}

	siblings := nodes[parent].children
	for i, sibling := range siblings {
		if sibling != inode {
			continue
		}
		next := i + delta
		if next < 0 || next >= len(siblings) {
			return inode
		}
		return siblings[next]
	}
	return inode
}

func moveToParent(nodes []AstNode, inode int) int {
	if inode < 0 || inode >= len(nodes) || nodes[inode].parent < 0 {
		return inode
	}
	return nodes[inode].parent
}

func moveToFirstChild(nodes []AstNode, inode int) int {
	if inode < 0 || inode >= len(nodes) || len(nodes[inode].children) == 0 {
		return inode
	}
	return nodes[inode].children[0]
}

// func (self )

func buildSourceLines(lines []string) []SourceLine {
	out := make([]SourceLine, 0, len(lines))
	off := 0
	for _, line := range lines {
		size := len([]rune(line))
		out = append(out, SourceLine{
			text:  line,
			start: off,
			end:   off + size,
		})
		off += size + 1
	}
	return out
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func drawHighlightedLine(s tcell.Screen, x, y int, line SourceLine, selStart, selEnd int, defStyle, hiStyle tcell.Style) {
	if selEnd-selStart == 0 {
		s.PutStrStyled(x, y, line.text, defStyle)
		return
	}
	runes := []rune(line.text)
	a := clamp(selStart-line.start, 0, len(runes))
	b := clamp(selEnd-line.start, 0, len(runes))

	if a >= b {
		s.PutStrStyled(x, y, line.text, defStyle)
		return
	}

	s.PutStrStyled(x, y, string(runes[:a]), defStyle)
	s.PutStrStyled(x+a, y, string(runes[a:b]), hiStyle)
	s.PutStrStyled(x+b, y, string(runes[b:]), defStyle)
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

	sourceLines := buildSourceLines(lines)
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
	textStyle := tcell.StyleDefault.Foreground(color.Black).Background(color.NewRGBColor(50, 250, 0))

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
	base := boxh + 1

	drawAll := func() {
		s.SetStyle(defStyle)
		s.Clear()

		in := inode
		if in <= node0 {
			node0 = in - 1
			node1 = node0 + boxh - 1
		}
		if in >= node1 {
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
			}
		}

		tok := view[inode].tok
		if tok == nil {
			tok = &Token{}
		}
		selStart := tok.Start
		selEnd := tok.End
		if selEnd <= selStart {
			selEnd = selStart + 1
		}
		l1 := getLine(selStart)
		l2 := getLine(selEnd-1)
		r := l2 - l1 + 1
		l0 := 0
		if len(lines) <= boxh {
			l0 = 0
		} else if r >= boxh {
			l0 = l1
		} else {
			l0 = l1 - (boxh-r)/2
			max0 := len(lines) - boxh
			if max0 < 0 {
				max0 = 0
			}
			l0 = clamp(l0, 0, max0)
		}

		for i := 0; i < boxh; i++ {
			l := l0 + i
			if l < len(sourceLines) {
				s.PutStr(0, i, fmt.Sprintf("%d", l+1))
				s.PutStr(3, i, "| ")
				drawHighlightedLine(s, 5, i, sourceLines[l], selStart, selEnd, defStyle, textStyle)
			}
		}

		s.Sync()
	}

	drawAll()
	for {
		s.Show()

		ev := <-s.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			// scrw, scrh = w.Size()
			drawAll()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				quit()
			case tcell.KeyUp:
				next := moveToSibling(view, inode, -1)
				if next != inode {
					inode = next
					drawAll()
				}
			case tcell.KeyDown:
				next := moveToSibling(view, inode, 1)
				if next != inode {
					inode = next
					drawAll()
				}
			case tcell.KeyLeft:
				next := moveToParent(view, inode)
				if next != inode {
					inode = next
					drawAll()
				}
			case tcell.KeyRight:
				next := moveToFirstChild(view, inode)
				if next != inode {
					inode = next
					drawAll()
				}
			case tcell.KeyPgUp:
			   inode--
			   if inode < 0 {
			      inode = 0
			   }
			   drawAll()
			case tcell.KeyPgDn:
			   inode++
			   if inode >= len(view) {
			      inode = len(view) - 1
			   }
			   drawAll()
			}
		}
	}
}

/*
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
*/
