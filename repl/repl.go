package repl

import (
	"bufio"
	"fmt"
	"gomadoufu/monkey-interpreter-go/lexer"
	"gomadoufu/monkey-interpreter-go/token"
	"io"
)

const PROMPT = ">> "

// NOTE: Rustでは:qでquitする機能つけたいね
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf("%s", PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
		}
	}
}
