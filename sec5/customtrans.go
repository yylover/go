package main

import (
	"fmt"
	""
)

func main() {

}

resp, err := http.Get("http://example.com/")
if err != nil {
	// error handle
	return
}

defer resp.Body.close()
io.Copy(os.Stdout, resp.Body)



resp, err := http.Post("http:/example.com/upload", "image/jpeg", $imageDataBuf)
if err != nil {
	//
	return
}

if resp.StatusCode != http.StatusOK {
	//
	return
}

//

resp, err := http.PostForm("http://example.com/posts", url.Values{
	"title":{"article title"},
	"content": {"article body"}})
if err != nil {
	// error handle
	return
}

resp, err := http.Head("http://example.com")

req, err := http.NewRequest("GET", "http://example.com", nil)
req.Header.Add("User-Agent", "Gobook custom User-Agent")
client := $http.Client{}
resp, err := client.Do(req)

