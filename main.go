package main

import (
  "fmt"
  "os"
  "path"
  "io/ioutil"
  "strings"

  "github.com/Dr0m0/gogen/parser"
  "github.com/Dr0m0/gogen/sqt"
)

func main() {
  // Get genesis file information
  genesis, err := getGenesis()
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  file, err := ioutil.ReadFile("templates/handler.go")
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
  handler := string(file)

  // Copy files to build directory
  os.Mkdir("build", 0755)
  var h string
  for _, v := range genesis.Routes {
    dir, _ := path.Split(v.File)

    genPathDirs("build", dir)

    fromPath      := path.Join(genesis.Root, v.File)
    toPath        := path.Join("build", v.File)
    relativePath  := path.Dir(v.File)

    t, code, err := sqt.NewSqTemplate(fromPath, toPath, genesis.Root, relativePath)
    if err != nil {
      fmt.Fprintln(os.Stderr, err)
      os.Exit(1)
    }

    // Generate route handler
    h = strings.Replace(handler, "<name>", "Teste", 1)
    h = strings.Replace(h, "<path>", toPath, 1)
    structs := ""
    actions := ""
    for ka, a := range t.Actions {
      structs += fmt.Sprintf("type action%dData struct {\n", ka)
      actions += fmt.Sprintf("var d%d action%dData\n\na := &t.Actions[%d]\n\n",
                             ka, ka, ka)

      for kq, q := range a.Queries {
        structs += fmt.Sprintf("%s query%dResult struct {\n", q.Name, kq)
        actions +=
          fmt.Sprintf("if err := s.db.Raw(a.Queries[%d].Code).Scan(&d%d.q%d).Error; err != nil {\n",
                      kq, ka, kq)
        actions += fmt.Sprintf("s.error(w, err, http.StatusInternalServerError)\nreturn\n}\n")

        for _, c := range q.Columns {
          structs += fmt.Sprintf("%s interface{}\n", c)
        }
        structs += "}\n"
      }

      mark := fmt.Sprintf("@@%d@@", ka)
      i := strings.Index(code, mark)
      actions += fmt.Sprintf("io.WriteString(w, code[:%d])\n", i)
      actions += fmt.Sprintf("t.Generate.Execute(w, d%d)\n", ka)
      actions += fmt.Sprintf("io.WriteString(w, code[:%d])\n\n", i+len(mark))
      structs += "}\n"
    }
    h = strings.Replace(h, "<structs>", structs, 1)
    h = strings.Replace(h, "<actions>", actions, 1)
  }

  file, err = ioutil.ReadFile("templates/handlers.go")
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
  handler = string(file)
  h = strings.Replace(handler, "<handlers>", h, 1)
  err = ioutil.WriteFile("build/handlers.go", []byte(h), 0644)
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  // Generate server schemas
  _, _, err = parser.ToGolang(genesis.Schemas, nil)
  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  // Parse the template
  // for _, v := range genesis.Routes {
  //   data, _ := ioutil.ReadFile(path.Join(genesis.Root, v.File))

  //   cfg.RelativePath = path.Dir(v.File)
  //   template, err := sqt.Evaluate(string(data), cfg)
  //   if err != nil {
  //     fmt.Fprintln(os.Stderr, err)
  //     os.Exit(1)
  //   }

  //   fmt.Print(template.Html)
  // }
}
