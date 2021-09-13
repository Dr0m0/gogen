package sqt

// lexer holds the state of the scanner.
type lexer struct {
  name    string        // used only for error reports.
  input   string        // the string being scanned.
  start   int           // start position of this lexeme.
  pos     int           // current position in the input.
  width   int           // width of last rune read.
  state   stateFn       // current lexer state.
  lexemes chan lexeme   // channel of scanned lexemes.
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

// newLexer creates a new scanner for the input string.
func newLexer(name, input string) *lexer {
  l := &lexer{
    name:     name,
    input:    input,
    state:    lexHTML,
    lexemes:  make(chan lexeme, 2),
  }
  return l
}

// nextLexeme returns the next lexeme from the input.
func (l *lexer) nextLexeme() lexeme {
  for {
    select {
    case lexeme := <-l.lexemes:
      return lexeme
    default:
      l.state = l.state(l)
    }
  }
  panic("not reached")
}
