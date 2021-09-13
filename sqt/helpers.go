package sqt

import (
  "fmt"
  "unicode/utf8"
  "strings"
)

const (
  eof = -1
)

// next returns the next rune in the input.
func (l *lexer) next() (lrune rune) {
  if l.pos >= len(l.input) {
    l.width = 0
    return eof
  }
  lrune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
  l.pos += l.width
  return lrune
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
  l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
  l.pos -= l.width
}

// peek returns but does not consume the next rune in the
// input.
func (l *lexer) peek() rune {
  lrune := l.next()
  l.backup()
  return lrune
}

// accept consumes the next rune.
// If it's from the valid set.
func (l *lexer) accept(valid string) bool {
  if strings.IndexRune(valid, l.next()) >= 0 {
    return true
  }
  l.backup()
  return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
  for strings.IndexRune(valid, l.next()) >= 0 {
  }
  l.backup()
}

// errorf returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
  l.lexemes <- lexeme{
    lexemeError,
    fmt.Sprintf(format, args...),
  }
  return nil
}

// emit passes an item back to the client.
func (l *lexer) emit(t lexemeType) {
  l.lexemes <- lexeme{t, l.input[l.start:l.pos]}
  l.start = l.pos
}

// findPrefix returns the prefix index if found it, -1 otherwise.
func findPrefix(s string, prefixes []string) int {
  for k,v := range prefixes {
    if strings.HasPrefix(s, v) {
      return k
    }
  }
  return -1
}
