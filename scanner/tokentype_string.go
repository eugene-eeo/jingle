// Code generated by "stringer -type=TokenType"; DO NOT EDIT.

package scanner

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenEOF-0]
	_ = x[TokenOr-1]
	_ = x[TokenAnd-2]
	_ = x[TokenFn-3]
	_ = x[TokenEnd-4]
	_ = x[TokenFor-5]
	_ = x[TokenWhile-6]
	_ = x[TokenIn-7]
	_ = x[TokenDo-8]
	_ = x[TokenNil-9]
	_ = x[TokenBoolean-10]
	_ = x[TokenString-11]
	_ = x[TokenNumber-12]
	_ = x[TokenIdent-13]
	_ = x[TokenComma-14]
	_ = x[TokenSemicolon-15]
	_ = x[TokenSeparator-16]
	_ = x[TokenLParen-17]
	_ = x[TokenRParen-18]
	_ = x[TokenLBrace-19]
	_ = x[TokenRBrace-20]
	_ = x[TokenLBracket-21]
	_ = x[TokenRBracket-22]
	_ = x[TokenBang-23]
	_ = x[TokenDot-24]
	_ = x[TokenPlus-25]
	_ = x[TokenMinus-26]
	_ = x[TokenMul-27]
	_ = x[TokenDiv-28]
	_ = x[TokenSet-29]
	_ = x[TokenEq-30]
	_ = x[TokenNeq-31]
	_ = x[TokenLt-32]
	_ = x[TokenGt-33]
	_ = x[TokenLeq-34]
	_ = x[TokenGeq-35]
}

const _TokenType_name = "TokenEOFTokenOrTokenAndTokenFnTokenEndTokenForTokenWhileTokenInTokenDoTokenNilTokenBooleanTokenStringTokenNumberTokenIdentTokenCommaTokenSemicolonTokenSeparatorTokenLParenTokenRParenTokenLBraceTokenRBraceTokenLBracketTokenRBracketTokenBangTokenDotTokenPlusTokenMinusTokenMulTokenDivTokenSetTokenEqTokenNeqTokenLtTokenGtTokenLeqTokenGeq"

var _TokenType_index = [...]uint16{0, 8, 15, 23, 30, 38, 46, 56, 63, 70, 78, 90, 101, 112, 122, 132, 146, 160, 171, 182, 193, 204, 217, 230, 239, 247, 256, 266, 274, 282, 290, 297, 305, 312, 319, 327, 335}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
