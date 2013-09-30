package controllers

import (
  dcli "github.com/fsouza/go-dockerclient"
  "github.com/dotcloud/docker"
  "code.google.com/p/go.net/websocket"
  "github.com/robfig/revel"
  "github.com/sridatta/instaworkshop/modules/docker/app"
  "io"
  "io/ioutil"
  "archive/tar"
  "text/template"
  "bytes"
  "fmt"
  "strings"
)

const DockerFileTemplate = `
from ubuntu:12.04

{{ if .Uploads }}
add {{.Uploads}} /{{.Uploads}}
run mkdir -p /workshop
run tar xzf /{{.Uploads}} -C /workshop
{{ end }}

{{ if .Script }}
add {{.Script}} /{{.Script}}
run /{{.Script}}
{{ end }}
`

type NullWriter struct {

}

func (n NullWriter) Write(p []byte) (int, error) {
  fmt.Println(string(p[:len(p)]))
  return len(p), nil
}

type DockerFile struct {
  Uploads string
  Script string
}

type App struct {
	*revel.Controller
  instaworkshop.Docker
}

func (c App) Index() revel.Result {
  return c.Redirect("/app")
}

func (c App) Create() revel.Result {
  revel.ParseParams(c.Params, c.Request)

  // Create an wrapper tarfile
  dockerFile := DockerFile{}

  tarFile, err := ioutil.TempFile("/tmp", "workshops")
  if err != nil {
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }
  tarWriter := tar.NewWriter(tarFile)

  // Add uploaded tar to wrapper tar
  var uploadArchive io.Reader
  c.Params.Bind(&uploadArchive, "file")

  if uploadArchive != nil {
    dockerFile.Uploads = "uploads.tar.gz"
    writeReaderToTar("uploads.tar.gz", &uploadArchive, tarWriter)
  }

  var script string
  script = c.Params.Values.Get("script")
  script = strings.Replace(script, "\r", "\n", -1)
  if script != "" {
    dockerFile.Script = "script.sh"
    writeBytesToTar("script.sh", []byte(script), tarWriter)
  }

  // Create a Dockerfile from the script param
  dFileString, err := renderDockerfile(dockerFile)
  if err != nil {
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }
  writeBytesToTar("Dockerfile", []byte(dFileString), tarWriter)
  tarWriter.Close()

  //Send this shit to the docker api
  buildOpts := dcli.BuildImageOptions {
    "workshops:"+c.Params.Values.Get("name"),
  }

  var out NullWriter
  tarFile.Seek(0, 0)
  c.DockerClient.BuildImage(buildOpts, tarFile, out)
  tarFile.Close()

  if err != nil {
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }


  return c.RenderText("OKAY MAYBE")
}

func (c App) Attach(user string, ws *websocket.Conn) revel.Result {
  config := docker.Config{}
  config.Cmd= []string{"/bin/bash"}
  config.Image= "workshops:"+c.Params.Values.Get("image")
  config.Tty = true
  config.AttachStdin = true
  config.AttachStdout = true
  config.OpenStdin = true
  config.StdinOnce = true

  container, err := c.DockerClient.CreateContainer(&config)
  if err != nil {
    fmt.Println(err.Error())
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }

  id := container.ShortID()
  err = c.DockerClient.StartContainer(id)
  if err != nil {
    fmt.Println(err.Error())
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }

  endpoint := "ws://127.0.0.1:4243/v1.5/containers/"+id+"/attach/ws?logs=1&stderr=1&stdout=1&stream=1&stdin=1"
  containerWs, err := websocket.Dial(endpoint, "", "http://localhost")
  if err != nil {
    fmt.Println(err.Error())
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }

  finished := make(chan bool)

  go func() {
    defer containerWs.Close()
    buf := make([]byte, 40*1024)

    for {
      n, err := containerWs.Read(buf)
      if err != nil {
        break
      }

      if n != 0 {
        ws.Write(buf[0:n])
      }

    }
    finished <- true
  }()

  go func() {
    defer ws.Close()
    defer containerWs.Close()
    buf := make([]byte, 40*1024)

    for {
      n, err := ws.Read(buf)
      if err != nil {
        break
      }

      if n != 0 {
        containerWs.Write(buf[0:n])
      }
    }

    finished <- true
  }()

  <- finished
  <- finished

  return c.RenderText("OK")
}

func (c App) List() revel.Result {
  images, err := c.DockerClient.ListImages(true)
  if err != nil {
    c.Response.Status = 500
    return c.RenderText(err.Error())
  }

  workshopImages := []docker.APIImages{}
  for _, image:= range images {
    if image.Repository == "workshops" {
      workshopImages = append(workshopImages, image)
    }
  }
  return c.RenderJson(workshopImages)
}

func writeReaderToTar(fname string, r *io.Reader, tw *tar.Writer) error {
  b, err := ioutil.ReadAll(*r)
  if err != nil {
    return err
  }

  return writeBytesToTar(fname, b, tw)
}

func writeBytesToTar(fname string, b []byte, tw *tar.Writer) error {
  err := tw.WriteHeader(&tar.Header{
    Name: fname,
    Size: int64(len(b)),
    Mode: 0777,
  })

  _, err = tw.Write(b)
  if err != nil {
    return err
  }

  return nil
}

func renderDockerfile(dockerFile DockerFile) (string, error) {
  template, err := template.New("dockerfile").Parse(DockerFileTemplate)
  if err != nil {
    return "", err
  }

  doc := new(bytes.Buffer)
  template.Execute(doc, dockerFile)
  return doc.String(), nil
}
