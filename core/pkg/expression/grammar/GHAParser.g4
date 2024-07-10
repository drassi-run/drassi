parser grammar GHAParser;

options {
    tokenVocab = GHALexer;
}

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
