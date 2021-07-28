package lexer

import "jingle/token"

func (l *Lexer) emitRuneToken(
	t token.TokenType,
	ch rune,
) token.Token {
	return token.Token{
		Type:    t,
		Literal: string(ch),
		LineNo:  l.input.lineNo,
		Column:  l.input.column,
	}
}

func (l *Lexer) emitStringToken(
	t token.TokenType,
	lit string,
) token.Token {
	return token.Token{
		Type:    t,
		Literal: lit,
		LineNo:  l.input.lineNo,
		Column:  l.input.column,
	}
}

func isPunctuation(ch rune) bool {
	return false || // aesthetic
		ch == '=' || ch == '+' || ch == '-' || ch == '!' || ch == '*' || ch == '/' || ch == ':' || ch == '.' ||
		ch == '|' || ch == '&' ||
		ch == '<' || ch == '>' ||
		ch == ',' || ch == ';' ||
		ch == '(' || ch == '{' || ch == '[' ||
		ch == ')' || ch == '}' || ch == ']'
}

func isSeparator(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

func isWhiteSpace(ch rune) bool {
	return ch == '\t' || ch == ' '
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
