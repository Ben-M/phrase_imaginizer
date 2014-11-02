package main

import (
  // "encoding/json"
  "github.com/go-martini/martini"
  "net/http"
  "strings"
  "encoding/json"
  "image"
  "bytes"
  "image/png"
  "os"
  "fmt"
)

func main() {
  m := martini.Classic()

  // http://localhost:3000/hello/myname
  m.Get("/v1/imaginize/:phrase", func(params martini.Params, res http.ResponseWriter, req *http.Request) (int, image.Image) {
    words := strings.Split(params["phrase"], " ")

    var channels []chan image.Image

    for _, word := range words {
      channel := make(chan image.Image)
      channels = append(channels, channel)
      go getImage(word, channel)
    }

    // output := ""
    // for _, channel := range channels {
    //   image := <- channel

    //   #we are getting images
    // }

    image := <- channels[0]

    out, _ := os.Create("test.png")
    png.Encode(out, image)

    return 200, image 
  })
  
  m.Run()
}

func getImage(word string, c chan image.Image) {
  urls := getUrls(word)
  image := downloadImage(urls)
  c <- image
}

func getUrls(word string) []string {
  resp, _ := http.Get("http://localhost:3001/v1/imageUrls/" + word)
  
  decoder := json.NewDecoder(resp.Body)
  var urls []string
  decoder.Decode(&urls)
  return urls
}

type ImageRequest struct {
    Urls []string
    Size uint8
}

func downloadImage(urls []string) image.Image{  
  url := "http://localhost:3002/get_image.json"

  request := ImageRequest{
    Urls: urls,
    Size: 150,
  }

  jsonStr, _ := json.Marshal(request)
  fmt.Println(string(jsonStr))

  req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
  
  client := &http.Client{}
  resp, _ := client.Do(req)
  defer resp.Body.Close()

  img, _, _:= image.Decode(resp.Body)
  return img
}