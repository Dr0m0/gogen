package parser

import (
  "fmt"
  "strings"
  "regexp"

  "github.com/jinzhu/inflection"
)

func init() {
  inflection.AddSingular("(cave)s", "${1}")
}

// GolangFormatter is used to define the formatting convention
type GolangFormatter interface {
  // toGolangName define the name convention
  toGolangName(value string) string

  // toGolangType translate .yaml file type to golang type
  toGolangType(value string) string
}

type golangConvention struct {
  namePattern *regexp.Regexp
  typesMap    map[string]string
  pkPattern   *regexp.Regexp
  fkPattern   *regexp.Regexp
}

func newGolangConvention() *golangConvention {
  return &golangConvention{
    namePattern:  regexp.MustCompile(`^\w|_\w`),
    typesMap:     map[string]string{
                    "uuid": "uuid.Uuid",
                    "timestamp": "time.Time",
                  },
    pkPattern:    regexp.MustCompile(`^pk|(,)pk,|pk$`),
    fkPattern:    regexp.MustCompile(`fk_\w+`),
  }
}

func (f *golangConvention) toGolangName(value string) string {
  r := f.namePattern.ReplaceAllStringFunc(value, strings.ToUpper)

  return inflection.Singular(strings.ReplaceAll(r, "_", ""))
}

func (f *golangConvention) toGolangType(value string) string {
  var name string
  var constrains string

  if i := strings.Index(value, " "); i != -1 {
    name = value[:i]
    constrains = value[i+1:]
  } else {
    name = value
  }

  if newName, ok := f.typesMap[name]; ok {
    name = newName
  }

  if constrains != "" {
    constrains = f.pkPattern.ReplaceAllString(constrains, "${1}primaryKey$1")
    constrains = f.fkPattern.ReplaceAllStringFunc(constrains, f.replaceFk)
    return fmt.Sprintf("%s `gorm:\"%s\"`", name, constrains)
  }

  return name
}

func (f *golangConvention) replaceFk(old string) string {
  return fmt.Sprintf("foreignKey:%s", f.toGolangName(old[3:]))
}
