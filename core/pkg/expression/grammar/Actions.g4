grammar Actions;

expression
    : e=expr EOF
    ;

expr
    : exprAccess
    | op=NOT expr
    | expr op=(LT|LTEQ|GTEQ|GT) expr
    | expr op=(EQUAL|NOTEQUAL) expr
    | expr op=AND expr
    | expr op=OR expr
    | literal
    ;

// exprAccess is splitted from expr because lhs of propertyAccess & indexAccess can't be a literal
// e.g 'foobar'['x'], true.really <= invalid
exprAccess
    : exprAccess (DOT props+=(IDENTIFIER | WILDCARD))+              # propertyAccess
    | exprAccess (LBRACK indexes+=expr RBRACK)+                     # indexAccess
    | identifier LPAREN (args+=expr (COMMA args+=expr)*)? RPAREN    # functionCall
    | LPAREN expr RPAREN                                            # wrap
    | identifier                                                    # variable
    ;

identifier
    : IDENTIFIER
    ;

literal
    : STRING
    | INTEGER
    | FLOAT
    | BOOLEAN
    | NULL
    ;


WS : [ \t\r\n]+ -> skip ; // skip whitespace

LPAREN   : '(' ;
RPAREN   : ')' ;
LBRACK   : '[' ;
RBRACK   : ']' ;
COMMA    : ',' ;
DOT      : '.' ;
EQUAL    : '==' ;
NOTEQUAL : '!=' ;
LT       : '<' ;
LTEQ     : '<=' ;
GT       : '>' ;
GTEQ     : '>=' ;
AND      : '&&' ;
OR       : '||' ;
NOT      : '!' ;
WILDCARD : '*' ;

fragment SIGN: [+-] ;
fragment DIGIT: [0-9] ;
fragment HEXDIGIT: [0-9a-fA-F] ;
fragment OCTDIGIT: [0-7] ;
fragment EXPONENT: [Ee] SIGN? DIGIT+ ;
fragment ALPHA: [a-zA-Z] ;
fragment ESC_SEQ: '\'\'' ;

NULL: 'null' ;

BOOLEAN
    : 'true'
    | 'false'
    ;

INTEGER
    : SIGN? DIGIT+
    | '0x' HEXDIGIT+  // Sign is not allowed in Hex number
    | '0o' OCTDIGIT+  // Sign is not allowed in Octal number
    ;

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/Sdk/ExpressionUtility.cs#L193
FLOAT
    : SIGN? DIGIT+ '.' DIGIT* EXPONENT? // fraction is optional, e.g "3.", "3.1e-1"
    | SIGN? DIGIT+ EXPONENT   // When fraction is missing, exponent is required to be a float number, e.g "2e3"
    | SIGN? '.' DIGIT+ EXPONENT?  // When decimal is missing, fraction is required, e.g ".5"
    | 'Infinity' | '-Infinity'
    | 'NaN'
    ;

// GHA string literal use single quote. Wrapping with double quotes (") will throw an error.
// Escape the literal single quote using an additional single quote ('').
// Other characters are raw, include \\, \r, \n \t,...
STRING: '\'' (ESC_SEQ | ~['])* '\'' ;

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/Sdk/ExpressionUtility.cs#L137
IDENTIFIER: (ALPHA | '_') (ALPHA | DIGIT | '_' | '-')* ;
