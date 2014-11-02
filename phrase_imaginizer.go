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
  "image/draw"
  "fmt"
)

func main() {
  m := martini.Classic()

  // http://localhost:3000/hello/myname
  m.Get("/v1/imaginize/:phrase", func(params martini.Params, res http.ResponseWriter, req *http.Request) (int, string) {
    words := strings.Split(params["phrase"], " ")

    var channels []chan image.Image

    for _, word := range words {
      channel := make(chan image.Image)
      channels = append(channels, channel)
      go getImage(word, channel)
    }

    var image_strip image.Image
    for _, channel := range channels {
      image := <- channel

      if image_strip == nil {
        image_strip = image
      } else {
        image_strip = combineImages(image_strip, image)
      }
    }

    out := new(bytes.Buffer)
    png.Encode(out, image_strip)
    return 200, string(out.Bytes())
  })
  
  m.Run()
}

func getImage(word string, c chan image.Image) {
  many_urls := getUrls(word)
  urls := many_urls[1:10]
  fmt.Println(urls) 
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
  Urls    []string    `json:"urls"`
  Height  uint        `json:"height_px"`
}

func downloadImage(urls []string) image.Image{  
  url := "http://localhost:3002/get_image.json"

  request := ImageRequest{
    Urls: urls,
    Height: 150,
  }

  jsonStr, _ := json.Marshal(request)

  req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
  
  client := &http.Client{}
  resp, _ := client.Do(req)
  defer resp.Body.Close()

  img, _, _:= image.Decode(resp.Body)
  return img
}

func combineImages(left_image image.Image, right_image image.Image) image.Image {
  
  separator_width := 10
  left_image_width := getWidth(left_image)
  right_image_width := getWidth(right_image)
  total_width := left_image_width + separator_width + right_image_width

  canvas := image.NewRGBA(image.Rect(0, 0, total_width, 150))

  drawImageAtPosition(canvas, left_image, image.Point{0,0})
  drawImageAtPosition(canvas, right_image, image.Point{left_image_width + separator_width,0})
  
  return canvas
}

func getWidth(image image.Image) int{
  return image.Bounds().Max.X
}

func drawImageAtPosition(canvas draw.Image, image_to_draw image.Image, position image.Point) {
  sr := image.Rect(0, 0, getWidth(image_to_draw), 150)
  r := image.Rectangle{position, position.Add(sr.Size())}
  
  draw.Draw(canvas, r, image_to_draw, sr.Min, draw.Src)
  return
}

