package instaworkshop

import (
  dcli "github.com/fsouza/go-dockerclient"
  "github.com/robfig/revel"
)

var DockerClient *dcli.Client

func Init() {
  docker, err := dcli.NewClient("http://127.0.0.1:4243")
  DockerClient = docker
  if err != nil {
    revel.ERROR.Fatal(err)
  }
}

type Docker struct {
  *revel.Controller
  DockerClient *dcli.Client
}

// Begin a transaction
func (c *Docker) Begin() revel.Result {
  c.DockerClient = DockerClient
  return nil
}

func init() {
  revel.InterceptMethod((*Docker).Begin, revel.BEFORE)
}
