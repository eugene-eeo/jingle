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
	_ = x[TokenIf-9]
	_ = x[TokenThen-10]
	_ = x[TokenElse-11]
	_ = x[TokenLet-12]
	_ = x[TokenNil-13]
	_ = x[TokenBoolean-14]
	_ = x[TokenString-15]
	_ = x[TokenNumber-16]
	_ = x[TokenIdent-17]
	_ = x[TokenComma-18]
	_ = x[TokenSeparator-19]
	_ = x[TokenLParen-20]
	_ = x[TokenRParen-21]
	_ = x[TokenLBrace-22]
	_ = x[TokenRBrace-23]
	_ = x[TokenLBracket-24]
	_ = x[TokenRBracket-25]
	_ = x[TokenBang-26]
	_ = x[TokenDot-27]
	_ = x[TokenPlus-28]
	_ = x[TokenMinus-29]
	_ = x[TokenMul-30]
	_ = x[TokenDiv-31]
	_ = x[TokenSet-32]
	_ = x[TokenEq-33]
	_ = x[TokenNeq-34]
	_ = x[TokenLt-35]
	_ = x[TokenGt-36]
	_ = x[TokenLeq-37]
	_ = x[TokenGeq-38]
}

const _TokenType_name = "TokenEOFTokenOrTokenAndTokenFnTokenEndTokenForTokenWhileTokenInTokenDoTokenIfTokenThenTokenElseTokenLetTokenNilTokenBooleanTokenStringTokenNumberTokenIdentTokenCommaTokenSeparatorTokenLParenTokenRParenTokenLBraceTokenRBraceTokenLBracketTokenRBracketTokenBangTokenDotTokenPlusTokenMinusTokenMulTokenDivTokenSetTokenEqTokenNeqTokenLtTokenGtTokenLeqTokenGeq"

var _TokenType_index = [...]uint16{0, 8, 15, 23, 30, 38, 46, 56, 63, 70, 77, 86, 95, 103, 111, 123, 134, 145, 155, 165, 179, 190, 201, 212, 223, 236, 249, 258, 266, 275, 285, 293, 301, 309, 316, 324, 331, 338, 346, 354}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
