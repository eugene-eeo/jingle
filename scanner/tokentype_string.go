// Code generated by "stringer -type=TokenType"; DO NOT EDIT.

package scanner

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenEOF-0]
	_ = x[TokenLet-1]
	_ = x[TokenOr-2]
	_ = x[TokenAnd-3]
	_ = x[TokenFn-4]
	_ = x[TokenEnd-5]
	_ = x[TokenFor-6]
	_ = x[TokenWhile-7]
	_ = x[TokenNil-8]
	_ = x[TokenBoolean-9]
	_ = x[TokenString-10]
	_ = x[TokenNumber-11]
	_ = x[TokenIdent-12]
	_ = x[TokenComma-13]
	_ = x[TokenSemicolon-14]
	_ = x[TokenSeparator-15]
	_ = x[TokenLParen-16]
	_ = x[TokenRParen-17]
	_ = x[TokenLBrace-18]
	_ = x[TokenRBrace-19]
	_ = x[TokenLBracket-20]
	_ = x[TokenRBracket-21]
	_ = x[TokenBang-22]
	_ = x[TokenDot-23]
	_ = x[TokenPlus-24]
	_ = x[TokenMinus-25]
	_ = x[TokenMul-26]
	_ = x[TokenDiv-27]
	_ = x[TokenSet-28]
	_ = x[TokenEq-29]
	_ = x[TokenNeq-30]
	_ = x[TokenLt-31]
	_ = x[TokenGt-32]
	_ = x[TokenLeq-33]
	_ = x[TokenGeq-34]
}

const _TokenType_name = "TokenEOFTokenLetTokenOrTokenAndTokenFnTokenEndTokenForTokenWhileTokenNilTokenBooleanTokenStringTokenNumberTokenIdentTokenCommaTokenSemicolonTokenSeparatorTokenLParenTokenRParenTokenLBraceTokenRBraceTokenLBracketTokenRBracketTokenBangTokenDotTokenPlusTokenMinusTokenMulTokenDivTokenSetTokenEqTokenNeqTokenLtTokenGtTokenLeqTokenGeq"

var _TokenType_index = [...]uint16{0, 8, 16, 23, 31, 38, 46, 54, 64, 72, 84, 95, 106, 116, 126, 140, 154, 165, 176, 187, 198, 211, 224, 233, 241, 250, 260, 268, 276, 284, 291, 299, 306, 313, 321, 329}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}