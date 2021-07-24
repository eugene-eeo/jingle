package lexer

import "jingle/token"

func (l *Lexer) emitRuneToken(
	t token.TokenType,
	ch rune,
) token.Token {
	return token.Token{
		Type:    t,
		Literal: string(ch),
		LineNo:  l.lineNo,
		Column:  l.column,
	}
}

func (l *Lexer) emitStringToken(
	t token.TokenType,
	lit string,
) token.Token {
	return token.Token{
		Type:    t,
		Literal: lit,
		LineNo:  l.lineNo,
		Column:  l.column,
	}
}

func isPunctuation(ch rune) bool {
	return false ||
		ch == '=' || ch == '+' || ch == '-' || ch == '!' || ch == '*' || ch == '/' || ch == ':' ||
		ch == '|' || ch == '&' ||
		ch == '<' || ch == '>' ||
		ch == ',' || ch == ';' ||
		ch == '(' || ch == '{' || ch == '[' ||
		ch == ')' || ch == '}' || ch == ']'
}

func isWhiteSpace(ch rune) bool {
	return ch == '\t' || ch == '\n' || ch == ' ' || ch == '\r'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
