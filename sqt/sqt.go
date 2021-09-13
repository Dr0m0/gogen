package sqt

import (
  "fmt"
  "strings"
  "regexp"
  "strconv"
  "io/ioutil"
  "path"
  "html/template"
)

// SqTemplate defines the sqt result
type SqTemplate struct {
  Actions []Action
}

type Action struct {
  Queries   []Query
  Generate  *template.Template
  Index     int
}

type Query struct {
  Name    string
  Columns []string
  Code    string
}

func (q *Query) String() string {
  return fmt.Sprintf("Name:\t\t%s\n\t%v\nCode:\n%s",
                      q.Name, q.Columns, q.Code)
}

// TODO: check for infinity recursive loop in PASTE commands
// WriteSqTemplate
func NewSqTemplate(fromPath, toPath, rootPath, relativePath string) (*SqTemplate, string, error) {
  data, err := ioutil.ReadFile(fromPath)
  if err != nil {
    return nil, "", err
  }

  var paste     string
  var newAction bool

  t := &SqTemplate{
    Actions:  []Action{},
  }
  result  := string(data)
  lexer   := newLexer("", result)
  offset  := 0
  start   := 0
  index   := 0

  for lexeme := lexer.nextLexeme();
      lexeme.typ != lexemeEOF;
      lexeme = lexer.nextLexeme() {

    switch lexeme.typ {
    case lexemeStartAction:
      start = lexer.pos
      newAction = true
    case lexemeEndAction:
      if paste != "" {
        result = fmt.Sprintf("%s%s%s",
                             result[:start-offset-2],
                             paste,
                             result[lexer.pos-offset:])
        offset += lexer.pos-start-len(paste)+2
        paste = ""
      } else {
        result = fmt.Sprintf("%s@@%d@@%s",
                             result[:start-offset-2],
                             index-1,
                             result[lexer.pos-offset:])
        offset += lexer.pos-4-start+len(strconv.Itoa(index))
      }
    case lexemePasteFile:
      var p string

      if lexeme.val[0] == '/' {
        p = path.Join(rootPath, lexeme.val[1:])
      } else {
        p = path.Join(rootPath, path.Join(relativePath, lexeme.val))
      }

      f, err := ioutil.ReadFile(p)
      if err != nil {
        return nil, "", err
      }

      paste = string(f)
    case lexemeSelect:
      if newAction {
        t.Actions = append(t.Actions, Action{})
        newAction = false
        index += 1
      }
      action := &t.Actions[len(t.Actions)-1]
      query := Query{}
      query.Columns = []string{}
      query.Code = "SELECT"
      action.Queries = append(action.Queries, query)
    case lexemeColumn:
      action := &t.Actions[len(t.Actions)-1]
      query := &action.Queries[len(action.Queries)-1]
      if len(query.Columns) > 0 {
        query.Code = fmt.Sprintf("%s, %s", query.Code, lexeme.val)
      } else {
        query.Code = fmt.Sprintf("%s %s", query.Code, lexeme.val)
      }
      query.Columns = append(query.Columns, toName(lexeme.val))
    case lexemeAlias:
      action := &t.Actions[len(t.Actions)-1]
      query := &action.Queries[len(action.Queries)-1]
      query.Code = fmt.Sprintf("%s AS %s", query.Code, lexeme.val)
      query.Columns[len(query.Columns)-1] = toName(lexeme.val)
    case lexemeFrom:
      action := &t.Actions[len(t.Actions)-1]
      action.Queries[len(action.Queries)-1].Code += " FROM "
    case lexemeSQL:
      action := &t.Actions[len(t.Actions)-1]
      action.Queries[len(action.Queries)-1].Code += lexeme.val
    case lexemeName:
      action := &t.Actions[len(t.Actions)-1]
      query := &action.Queries[len(action.Queries)-1]
      query.Name = toName(lexeme.val)
    case lexemeError:
      return nil, "", fmt.Errorf(lexeme.val)
    }
  }

  err = ioutil.WriteFile(toPath, []byte(result), 0644)
  if err != nil {
    return nil, "", err
  }

  return t, result, nil
}

// LoadSqTemplate
func LoadSqTemplate(code string) (*SqTemplate, error) {
  lexer := newLexer("", code)
  t := &SqTemplate{
    Actions:  []Action{},
  }
  offset := 0
  start := 0

  for lexeme := lexer.nextLexeme();
      lexeme.typ != lexemeEOF;
      lexeme = lexer.nextLexeme() {

    switch lexeme.typ {
    case lexemeStartAction:
      start = lexer.pos
      t.Actions = append(t.Actions, Action{})
    case lexemeEndAction:
      i := len(t.Actions)-1
      offset += lexer.pos-4-start+len(strconv.Itoa(i))
    case lexemeSelect:
      action := &t.Actions[len(t.Actions)-1]
      query := Query{}
      query.Columns = []string{}
      query.Code = "SELECT"
      action.Queries = append(action.Queries, query)
    case lexemeColumn:
      action := &t.Actions[len(t.Actions)-1]
      query := &action.Queries[len(action.Queries)-1]
      if len(query.Columns) > 0 {
        query.Code = fmt.Sprintf("%s, %s", query.Code, lexeme.val)
      } else {
        query.Code = fmt.Sprintf("%s %s", query.Code, lexeme.val)
      }
      query.Columns = append(query.Columns, toName(lexeme.val))
    case lexemeAlias:
      action := &t.Actions[len(t.Actions)-1]
      query := &action.Queries[len(action.Queries)-1]
      query.Code = fmt.Sprintf("%s AS %s", query.Code, lexeme.val)
      query.Columns[len(query.Columns)-1] = toName(lexeme.val)
    case lexemeFrom:
      action := &t.Actions[len(t.Actions)-1]
      action.Queries[len(action.Queries)-1].Code += " FROM "
    case lexemeSQL:
      action := &t.Actions[len(t.Actions)-1]
      action.Queries[len(action.Queries)-1].Code += lexeme.val
    case lexemeName:
      action := &t.Actions[len(t.Actions)-1]
      query := &action.Queries[len(action.Queries)-1]
      query.Name = toName(lexeme.val)
    case lexemeError:
      return nil, fmt.Errorf(lexeme.val)
    }
  }

  return t, nil
}

var (
  namePattern = regexp.MustCompile(`^\w|_\w`)
)

func toName(column string) string {
  if i := strings.Index(column, "."); i != -1 {
    return toName(column[i+1:])
  }
  s := namePattern.ReplaceAllStringFunc(column, strings.ToUpper)
  return strings.ReplaceAll(s, "_", "")
}
