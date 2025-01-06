package app

const (
	TOKEN_NAME string = "TOKEN_NAME"
	TOKEN_OPAREN string = "TOKEN_OPAREN"
	TOKEN_CPAREN string = "TOKEN_CPAREN"
	TOKEN_OCURLY string = "TOKEN_OCURLY"
	TOKEN_CCURLY string = "TOKEN_CCURLY"
	TOKEN_COMMA string = "TOKEN_COMMA"
	TOKEN_SEMICOLON string = "TOKEN_SEMICOLON"
	TOKEN_NUMBER string = "TOKEN_NUMBER"
	TOKEN_STRING string = "TOKEN_STRING"
	TOKEN_RETURN string = "TOKEN_RETURN"
)

type Token struct {
	Type string
	Value any
	Loc Loc
}