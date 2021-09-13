package parser

import (
  "fmt"
  "strings"
  "io/ioutil"
)

type Table struct {
  Name    string
  Columns []Column
}

type Column struct {
  Name string
  Type string
}

// ToGolang compile the schemas to a golang using GORM.
func ToGolang(schemas map[string]interface{}, formatter GolangFormatter) ([]Table, string, error) {
  if formatter == nil {
    formatter = newGolangConvention()
  }

  var containsTimestamp bool
  var containsUuid bool
  var content string
  var tables []Table

  template, _ := ioutil.ReadFile("templates/model.txt")
  file := string(template)
  for ki, vi := range schemas {
    table := Table{Name: formatter.toGolangName(ki)}

    content += fmt.Sprintf("type %s struct {\n", table.Name)
    for kj, vj := range vi.(map[interface{}]interface{}) {
      if _, ok := vj.(string); !ok {
        return nil, "", fmt.Errorf("Invalid schema object on %v type:%v\n", kj, vj)
      }

      if !containsTimestamp && vj.(string) == "timestamp" {
        containsTimestamp = true
      }
      if !containsUuid && strings.HasPrefix(vj.(string), "uuid") {
        containsUuid = true
      }

      column := Column{
        Name: formatter.toGolangName(kj.(string)),
        Type: formatter.toGolangType(vj.(string)),
      }
      table.Columns = append(table.Columns, column)

      content += fmt.Sprintf("\t%s %s\n", column.Name, column.Type)
    }

    tables = append(tables, table)
    content += "}\n"
    content += fmt.Sprintf("func(%s)TableName()string{return \"%s\"}\n", table.Name, ki)
  }

  file = strings.Replace(file, "<tables>", content, 1)

  if containsTimestamp || containsUuid {
    var libs string
    if containsTimestamp {
      libs += "\t\"time\"\n"
    }
    if containsUuid {
      libs += "\t\"github.com/google/uuid\"\n"
    }
    file = strings.Replace(file, "<libs>", libs, 1)
  }

  var migrate string
  for _, v := range tables {
    migrate += fmt.Sprintf("\tdb.Migrator().AutoMigrate(&%s{})\n", v.Name)
  }
  file = strings.Replace(file, "<migrate>", migrate, 1)

  return tables, file, nil
}
