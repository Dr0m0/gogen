package sqt

import (
  "fmt"
)

// lexemeType identifies the type of lexeme.
type lexemeType int
const (
  lexemeError lexemeType = iota // error ocurred;

  lexemeHTML                    // HTML code
  lexemeStartAction             // meta string '@@'
  lexemeEndAction               // meta string '@@'
  lexemePaste                   // paste keyword
  lexemePasteFile               // file path to be paste
  lexemeQuery                   // query start keyword
  lexemeSelect                  // select keyword
  lexemeFrom                    // from keyword
  lexemeSQL                     // sql code
  lexemeAgg                     // SQL aggregator
  lexemeColumn                  // query column name
  lexemeAlias                   // column alias
  lexemeStartNaming             // named keyword
  lexemeName                    // query name
  lexemeGenerating              // generate keyword
  lexemeTemplate                // golang template
  lexemeEOF
)
const (
  strMeta       = "@@"
  strPaste      = "PASTE"
  strQuery      = "QUERY"
  strNamed      = "NAMED"
  strGenerate   = "GENERATE"
  strSelect     = "SELECT"
  strAs         = "AS"
  strDistinct   = "DISTINCT"
  strFrom       = "FROM"
  strDigit      = "0123456789"
  // IMPORTANT: dont change the dot position from this string
  // some code depends on this.
  strValidName  = ".abcdefghijklmnopqrstuvwxyz_0123456789"
)
var (
  strAggs = []string{"COUNT","AVG","SUM","MIN","MAX"}
)

// lexeme represents a token returned from the scanner.
type lexeme struct {
  typ lexemeType  // Type, such as lexemePath
  val string      // Value, such as "/style.css"
}

func (l lexeme) String() string {
  switch l.typ {
  case lexemeEOF:
    return "EOF"
  case lexemeError:
    return l.val
  }

  if len(l.val) > 15 {
    return fmt.Sprintf("%.15q...", l.val)
  }
  return fmt.Sprintf("%q", l.val)
}
