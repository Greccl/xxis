package debug

import (
	"fmt"
	"os"
	"unicode/utf8"
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

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func DebugAst(ast *Token) {
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

	boxh := 18
	inode := 0
	node0 := 0
	node1 := boxh - 1

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
				s.PutStrStyled(0, i, node.prefix, defStyle)
				if in == inode {
					s.PutStrStyled(l, i, node.name, nodeStyle)
				} else {
					s.PutStrStyled(l, i, node.name, defStyle)
				}
				in++
			}
		}

		s.Sync()
	}

	drawAll()
	for {
		ev := <-s.EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
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