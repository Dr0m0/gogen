package main

import (
  "os"
  "fmt"
  "path"
  "io/ioutil"

  "gopkg.in/yaml.v2"
)

func getGenesis() (*Genesis, error) {
  if len(os.Args) <= 1 {
    return nil, fmt.Errorf("Please insert the genesis file path")
  }

  // Get genesis file
  data, err := ioutil.ReadFile(os.Args[1])
  if err != nil {
    return nil, err
  }

  var genesis Genesis
  if err := yaml.Unmarshal(data, &genesis); err != nil {
    return nil, err
  }

  // Check and normalize genesis data
  for _, k := range genesis.Routes {
    if err := k.check(); err != nil {
      return nil, err
    }
    k.normalize()
  }

  genesis.Root = path.Dir(os.Args[1])

  return &genesis, nil
}

// genPathDirs make all directories of value inside root 
func genPathDirs(root string, value string) {
  if value[len(value)-1] == '/' {
    value = value[:len(value)-1]
  }
  value, dir := path.Split(value)

  if value != "." && value != "" {
    genPathDirs(root, value)
    os.Mkdir(fmt.Sprintf("%s/%s/%s", root, value, dir), 0755)
  } else {
    os.Mkdir(fmt.Sprintf("%s/%s", root, dir), 0755)
  }
}
