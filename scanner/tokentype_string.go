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
	_ = x[TokenIn-8]
	_ = x[TokenDo-9]
	_ = x[TokenNil-10]
	_ = x[TokenBoolean-11]
	_ = x[TokenString-12]
	_ = x[TokenNumber-13]
	_ = x[TokenIdent-14]
	_ = x[TokenComma-15]
	_ = x[TokenSemicolon-16]
	_ = x[TokenSeparator-17]
	_ = x[TokenLParen-18]
	_ = x[TokenRParen-19]
	_ = x[TokenLBrace-20]
	_ = x[TokenRBrace-21]
	_ = x[TokenLBracket-22]
	_ = x[TokenRBracket-23]
	_ = x[TokenBang-24]
	_ = x[TokenDot-25]
	_ = x[TokenPlus-26]
	_ = x[TokenMinus-27]
	_ = x[TokenMul-28]
	_ = x[TokenDiv-29]
	_ = x[TokenSet-30]
	_ = x[TokenEq-31]
	_ = x[TokenNeq-32]
	_ = x[TokenLt-33]
	_ = x[TokenGt-34]
	_ = x[TokenLeq-35]
	_ = x[TokenGeq-36]
}

const _TokenType_name = "TokenEOFTokenLetTokenOrTokenAndTokenFnTokenEndTokenForTokenWhileTokenInTokenDoTokenNilTokenBooleanTokenStringTokenNumberTokenIdentTokenCommaTokenSemicolonTokenSeparatorTokenLParenTokenRParenTokenLBraceTokenRBraceTokenLBracketTokenRBracketTokenBangTokenDotTokenPlusTokenMinusTokenMulTokenDivTokenSetTokenEqTokenNeqTokenLtTokenGtTokenLeqTokenGeq"

var _TokenType_index = [...]uint16{0, 8, 16, 23, 31, 38, 46, 54, 64, 71, 78, 86, 98, 109, 120, 130, 140, 154, 168, 179, 190, 201, 212, 225, 238, 247, 255, 264, 274, 282, 290, 298, 305, 313, 320, 327, 335, 343}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
