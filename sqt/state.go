package sqt

import (
  "strings"
  "unicode"
  "html/template"

  "github.com/xwb1989/sqlparser"
)

// lexHTML search for a start meta string to begin parser.
func lexHTML(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], strMeta) {
      if l.pos > l.start {
        l.emit(lexemeHTML)
      }
      return lexStartAction
    }
    if l.next() == eof { break }
  }

  // Correctly reached EOF.
  if l.pos > l.start {
    l.emit(lexemeHTML)
  }
  l.emit(lexemeEOF) // Useful to make EOF a token.
  return nil        // Stop the run loop.
}

// lexStartAction starts an action.
func lexStartAction(l *lexer) stateFn {
  l.pos += len(strMeta)
  l.emit(lexemeStartAction)
  return lexInsideAction
}

// lexInsideAction computes the action inside the meta
// string.
func lexInsideAction(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], strMeta) {
      return l.errorf("Empty action")
    }
    if strings.HasPrefix(l.input[l.pos:], strQuery) {
      return lexQuery
    }
    if strings.HasPrefix(l.input[l.pos:], strPaste) {
      return lexPaste
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unclosed action")
    case unicode.IsSpace(r):
      l.ignore()
    }
  }
}

// lexEndAction finish an action
func lexEndAction(l *lexer) stateFn {
  l.pos += len(strMeta)
  l.emit(lexemeEndAction)
  return lexHTML
}

// lexQuery start an sql query.
func lexQuery(l *lexer) stateFn {
  l.pos += len(strQuery)
  l.emit(lexemeQuery)
  return lexStartSQL
}

// lexPaste defines the paste action.
func lexPaste(l *lexer) stateFn {
  l.pos += len(strPaste)
  l.emit(lexemePaste)
  return lexPasteFile
}

// lexPasteFile gets file to be paste.
func lexPasteFile(l *lexer) stateFn {
  // Ignore spaces
  for loop := true; loop; {
    switch r := l.next(); {
    case r == eof:
      return l.errorf("empty action")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      l.backup()
      loop = false
    }
  }

  // Read path, cannot have spaces
  for loop := true; loop; {
    if strings.HasPrefix(l.input[l.pos:], strMeta) {
      l.emit(lexemePasteFile)
      return lexEndAction
    }
    switch r := l.next(); {
    case r == eof:
      return l.errorf("unclosed action")
    case unicode.IsSpace(r):
      l.backup()
      l.emit(lexemePasteFile)
      loop = false
    }
  }

  // Look for ending the action
  for {
    if strings.HasPrefix(l.input[l.pos:], strMeta) {
      return lexEndAction
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unclosed action")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      return l.errorf("invalid character '%c', expecting '@@'", r)
    }
  }
}

// lexStartSQL start to read the SQL code.
func lexStartSQL(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], strSelect) {
      l.pos += len(strSelect)
      l.emit(lexemeSelect)
      return lexStartSelect
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      return l.errorf("invalid character, only accepts SELECT commands")
    }
  }
}

// lexStartSelect start a select statement.
func lexStartSelect(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], strDistinct) {
      return lexColumns
    }
    if i := findPrefix(l.input[l.pos:], strAggs); i != -1 {
      l.pos += len(strAggs[i])
      l.emit(lexemeAgg)
      return lexSelectAgg
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      l.backup()
      return lexColumns
    }
  }
}

// lexSelectAgg get the aggregator name.
func lexSelectAgg(l *lexer) stateFn {
  for {
    switch r := l.next(); {
    case r == eof:
      return l.errorf("")
    case unicode.IsSpace(r):
      l.ignore()
    case r == '(':
      return lexAggColumn
    default:
      return l.errorf("invalid character '%c', expecting '('.", r)
    }
  }
}

// lexAggColumn get the parameter of an aggregator.
func lexAggColumn(l *lexer) stateFn {
  // Cannot start with a digit
  if l.accept(strDigit) {
    return l.errorf("a column name cannot start with a digit.")
  }

  // Get name
  l.acceptRun(strValidName)
  if l.peek() != ')' {
    return l.errorf("invalid character '%c', expect ')'.", l.next())
  }
  if l.start == l.pos {
    return l.errorf("expecting a name.")
  }

  l.emit(lexemeColumn)
  l.ignore()  // ignore ')' character

  // Look for FROM keyword
  for {
    if strings.HasPrefix(l.input[l.pos:], strFrom) {
      return lexEndSelect
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code.")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      return l.errorf("invalid character '%c', expecting FROM.", r)
    }
  }
}

// lexColumns get all columns from a SELECT statement.
func lexColumns(l *lexer) stateFn {
  naming := false
  // Process column name
  for loop := true; loop; {
    if l.accept(strValidName) {
      naming = true
      continue
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code")
    case unicode.IsSpace(r):
      if naming {
        l.backup()
        l.emit(lexemeColumn)
        l.next()
        loop = false
      } else {
        l.ignore()
      }
    case r == ',':
      l.backup()
      l.emit(lexemeColumn)
      l.next()
      return lexColumns
    default:
      return l.errorf("invalid character '%c'.", r)
    }
  }

  // See if will look for more names, or
  // if the name has an alias.
  for {
    if strings.HasPrefix(l.input[l.pos:], strAs) {
      return lexColumnAlias
    }
    if strings.HasPrefix(l.input[l.pos:], strFrom) {
      return lexEndSelect
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code.")
    case unicode.IsSpace(r):
      l.ignore()
    case r == ',':
      return lexColumns
    default:
      return l.errorf("invalid character '%c', expecting ',' or 'AS'.", r)
    }
  }
}

// lexColumnAlias defines the alias of a column.
func lexColumnAlias(l *lexer) stateFn {
  // ignore 'AS'
  l.pos += len(strAs)
  l.ignore()

  if !unicode.IsSpace(l.peek()) {
    return l.errorf("invalid chracter '%c', expecting space.", l.next())
  }

  // Look for a the start of a name
  for loop := true; loop; {
    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code.")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      loop = false
    }
  }

  // Get alias name
  l.acceptRun(strValidName[1:]) // accept name WITHOUT '.' char
  if !unicode.IsSpace(l.peek()) {
    return l.errorf("invalid character '%c', expecting space.", l.next())
  }
  l.emit(lexemeAlias)

  // Finish select state or continue to look for columns names
  for {
    if strings.HasPrefix(l.input[l.pos:], strFrom) {
      return lexEndSelect
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code.")
    case unicode.IsSpace(r):
      l.ignore()
    case r == ',':
      return lexColumns
    default:
      return l.errorf("invalid character '%c', expecting ',' or FROM.", r)
    }
  }
}

// lexEndSelect finish a SELECT command.
func lexEndSelect(l *lexer) stateFn {
  l.pos += len(strFrom)
  l.emit(lexemeFrom)
  return lexEndSQL
}

// lexEndSQL gets all SQL code after the first SELECT command.
func lexEndSQL(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], strNamed) {
      // Parse SQL query 
      stmt, err := sqlparser.Parse("SELECT * FROM " + l.input[l.start:l.pos-1])
      if err != nil {
        return l.errorf("SQL format: %v", err)
      }

      // Only allows SELECT type statements
      switch stmt.(type) {
      case *sqlparser.Select:
        l.emit(lexemeSQL)
        return lexStartNaming
      default:
        return l.errorf("Can only make SELECT statements")
      }
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("unfinish SQL code")
    }
  }
}

// lexStartNaming begin to name a query.
func lexStartNaming(l *lexer) stateFn {
  l.pos += len(strNamed)
  l.emit(lexemeStartNaming)

  // Demands a first space and then ignore the rest.
  r := l.next()
  if !unicode.IsSpace(r) {
    return l.errorf("invalid character '%c', expected a space.", r)
  }
  for loop := true; loop; {
    switch r = l.next(); {
    case r == eof:
      return l.errorf("unfinish naming")
    case unicode.IsSpace(r):
      l.ignore()
    default:
      l.backup()
      loop = false
    }
  }

  return lexName
}

// lexName gets a name
func lexName(l *lexer) stateFn {
  // Cannot start with a number.
  if l.accept(strDigit) {
    return l.errorf("a name cannot begin with a number.")
  }

  // Accepts snake case only
  l.acceptRun(strValidName)
  if !unicode.IsSpace(l.peek()) {
    return l.errorf("invalid character '%c', name needs to be in snake case.", l.next())
  }
  if l.start == l.pos {
    return l.errorf("needs a name")
  }
  l.emit(lexemeName)

  // After the name emit procced to go to a query or a generate action
  for {
    if strings.HasPrefix(l.input[l.pos:], strQuery) {
      return lexQuery
    }
    if strings.HasPrefix(l.input[l.pos:], strGenerate) {
      return lexStartGenerating
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("Unfinish naming")
    case unicode.IsSpace(r):
      l.ignore()
    }
  }
}

// lexStartGenerating start a GENERATE action.
func lexStartGenerating(l *lexer) stateFn {
  l.pos += len(strGenerate)
  l.emit(lexemeGenerating)
  return lexTemplate
}

// lexTemplate computes golang template.
func lexTemplate(l *lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], strMeta) {
      _, err := template.New("tmp").Parse(l.input[l.start:l.pos-1])
      if err != nil {
        return l.errorf("Invalid template: %v", err)
      }

      l.emit(lexemeTemplate)
      return lexEndAction
    }

    switch r := l.next(); {
    case r == eof:
      return l.errorf("Unclosed action")
    }
  }
}
