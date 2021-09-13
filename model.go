package main

import (
  "fmt"
  "strings"
)

type Genesis struct {
  Name    string                  `yaml:"name"`
  Routes  []Route                 `yaml:"routes"`
  Schemas map[string]interface{}  `yaml:"schemas"`

  Root string // Root is the directory of the genesis.yaml file
}

type Route struct {
  Web   string `yaml:"web"`
  File  string `yaml:"file"`
}

func (r *Route) check() error {
  if i := strings.Index(r.Web, "./"); i != -1 {
    return fmt.Errorf("Invalid character[%d] '.' in web route '%s'.", i, r.Web)
  } else if i := strings.Index(r.File, "./"); i != -1 {
    return fmt.Errorf("Invalid character[%d] '.' in file route '%s'.", i, r.File)
  } else if i := strings.Index(r.Web, "/."); i != -1 {
    return fmt.Errorf("Invalid character[%d] '.' in web route '%s'.", i, r.Web)
  } else if i := strings.Index(r.File, "/."); i != -1 {
    return fmt.Errorf("Invalid character[%d] '.' in file route '%s'.", i, r.File)
  } else if i := strings.Index(r.Web, "//"); i != -1 {
    return fmt.Errorf("Invalid characters[%d] '//' in web route '%s'.", i, r.Web)
  } else if i := strings.Index(r.File, "//"); i != -1 {
    return fmt.Errorf("Invalid characters[%d] '//' in web route '%s'.", i, r.File)
  }

  return nil
}

func (r *Route) normalize() {
  if r.File[0] == '/' {
    r.File = r.File[:1]
  }
}
