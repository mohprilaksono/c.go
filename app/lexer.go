package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	FilePath string
	Source string
	Cur int
	Bol int
	Row int
}

func NewLexer(filePath string, source string) *Lexer {
	return &Lexer{
		FilePath: filePath,
		Source: source,
	}
}

func (lexer *Lexer) IsNotEmpty() bool {
	return lexer.Cur < len(lexer.Source)
}

func (lexer *Lexer) IsEmpty() bool {
	return !lexer.IsNotEmpty()
}

func (lexer *Lexer) ChopChar() {
	if lexer.IsNotEmpty() {
		x := lexer.Source[lexer.Cur]
		lexer.Cur = lexer.Cur + 1
		if x == '\n' {
			lexer.Bol = lexer.Cur
			lexer.Row += 1
		}
	}
}

func (lexer *Lexer) Loc() *Loc {
	return &Loc{lexer.FilePath, lexer.Row, lexer.Cur - lexer.Bol}
}

func (lexer *Lexer) TrimLeft() {
	for lexer.IsNotEmpty() && unicode.IsSpace(rune(lexer.Source[lexer.Cur])) {
		lexer.ChopChar()
	}
}

func (lexer *Lexer) DropLine() {
	for lexer.IsNotEmpty() && lexer.Source[lexer.Cur] != '\n' {
		lexer.ChopChar()
	}

	if lexer.IsNotEmpty() {
		lexer.ChopChar()
	}
}

func (lexer *Lexer) NextToken() (Token, bool) {
	lexer.TrimLeft()
	for lexer.IsNotEmpty() {
		s := lexer.Source[lexer.Cur:]
		if !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, "//") {
			break
		}

		lexer.DropLine()
		lexer.TrimLeft()
	}

	if lexer.IsEmpty() {
		return Token{}, false
	}

	loc := lexer.Loc()
	first := lexer.Source[lexer.Cur]

	if unicode.IsLetter(rune(first)) {
		index := lexer.Cur
		for lexer.IsNotEmpty() && unicode.IsLetter(rune(lexer.Source[lexer.Cur])) || unicode.IsNumber(rune(lexer.Source[lexer.Cur])) {
			lexer.ChopChar()
		}

		value := lexer.Source[index:lexer.Cur - index]
		return Token{
			Loc: *loc,
			Type: TOKEN_NAME,
			Value: value,
		}, true
	}

	literalTokens := make(map[byte]string)
	literalTokens['('] = TOKEN_OPAREN
	literalTokens[')'] = TOKEN_CPAREN
	literalTokens['{'] = TOKEN_OCURLY
	literalTokens['}'] = TOKEN_CCURLY
	literalTokens[','] = TOKEN_COMMA
	literalTokens[';'] = TOKEN_SEMICOLON

	val, ok := literalTokens[first]
	if ok {
		lexer.ChopChar()
		return Token{
			Loc: *loc,
			Type: val,
			Value: string(first),
		}, true
	}

	if first == '"' {
		lexer.ChopChar()
		var literal string
		for lexer.IsNotEmpty() {
			ch := lexer.Source[lexer.Cur]
			switch ch {
			case '"':
			case '\\':
				lexer.ChopChar()
				if lexer.IsEmpty() {
					fmt.Printf("%v: ERROR: unfinished escape sequence\n", lexer.Loc())
					os.Exit(1)
				}

				escape := lexer.Source[lexer.Cur]
				if escape == 'n' {
					literal = literal + "\n"
					lexer.ChopChar()
				} else if escape == '"' {
					literal = literal + "\""
					lexer.ChopChar()
				} else {
					fmt.Printf("%v: ERROR: unknown escape sequence starts with %d\n", lexer.Loc(), escape)
				}
			default:
				literal = literal + string(ch)
				lexer.ChopChar()
			}
		}

		if lexer.IsNotEmpty() {
			lexer.ChopChar()
			return Token{
				Loc: *loc,
				Type: TOKEN_STRING,
				Value: literal,
			}, true
		}

		fmt.Printf("%s: ERROR: unclosed string literal\n", loc.Display())
		os.Exit(1)
	}

	if unicode.IsDigit(rune(first)) {
		start := lexer.Cur
		for lexer.IsNotEmpty() && unicode.IsDigit(rune(lexer.Source[lexer.Cur])) {
			lexer.ChopChar()	
		}

		value, err := strconv.Atoi(lexer.Source[start:lexer.Cur - start]) 
		if err != nil {
			panic(err)
		}

		return Token{
			Loc: *loc,
			Type: TOKEN_NUMBER,
			Value: value,
		}, true
	}

	fmt.Printf("%s: ERROR: unknown token starts with %d\n", loc.Display(), first)
	os.Exit(1)

	return Token{}, false
}
