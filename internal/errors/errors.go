package errors

import (
   "fmt"
   "os"
   "bufio"
)

type ErrorMessage struct {
   Msg string
}

type ErrorWithRange struct {
   Msg string
   Path string
   Start int
   End int
}

func HandlePanicForPath(path string) {
   if err := recover(); err != nil {
      switch x := err.(type) {
      case ErrorWithRange:
         x.Path = path
         Print(x)
         os.Exit(1)
      }
      panic(err)
   }
}

type SourceLine struct {
	Text  string
	Start int
	End   int
}

func getLine(lines []SourceLine, offset int) int {
   for i, l := range lines {
      if offset <= l.End {
         return i
      }
   }
   return 0
}

func buildSourceLines(lines []string) []SourceLine {
	out := make([]SourceLine, 0, len(lines))
	off := 0
	for _, line := range lines {
		size := len([]rune(line))
		out = append(out, SourceLine{
			Text:  line,
			Start: off,
			End:   off + size,
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

func drawSourceLine(line SourceLine, index, selStart, selEnd int) {
   if selEnd == -1 {
      fmt.Printf("\033[38;2;235m   |\033[0m\n")
      return
   }

   l := selEnd - selStart

   if l == 0 {
      fmt.Print("\033[38;2;235m")
   }

   fmt.Printf("%3d| ", index+1)
	if l == 0 {
		fmt.Printf("%s\033[0m\n", line.Text)
		return
	}

	runes := []rune(line.Text)
	a := clamp(selStart-line.Start, 0, len(runes))
	b := clamp(selEnd-line.Start, 0, len(runes))

	if a >= b {
		fmt.Printf("%s\n", line.Text)
		return
	}

   fmt.Print("\033[38;2;255m")
	fmt.Printf("%s", string(runes[:a]))
	fmt.Printf("\033[31m%s\033[0m", string(runes[a:b]))
	fmt.Printf("%s\n", string(runes[b:]))
}

func Print(e ErrorWithRange) {
	file, err := os.Open(e.Path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close() // Ensure the file is closed

	scanner := bufio.NewScanner(file)

	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text()) // .Text() returns the line as a string
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sourceLines := buildSourceLines(lines)

   num := getLine(sourceLines, e.Start)
   fmt.Printf("%s, line %d: %s\n", e.Path, num+1, e.Msg)
   n0 := num - 2
   if n0 < 0 { n0 = 0 }
   for i:=0; i<5; i++ {
      n := n0 + i
      if n == num {
         drawSourceLine(sourceLines[n], n, e.Start, e.End)
      } else if n < len(sourceLines) {
         drawSourceLine(sourceLines[n], n, 0, 0)
      } else {
         drawSourceLine(sourceLines[0], n, 0, -1)
      }
   }
}