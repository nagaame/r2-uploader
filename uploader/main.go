package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	RootUrl        = ""
	uploaderSecret = ""
)

func main() {
	// change this to your own secret
	RootUrl = "https://xxxxxxxxxxxxxx.workers.dev/"
	uploaderSecret = "xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	var urls []string
	var files []string
	args := os.Args[1:]
	for _, arg := range args {
		if _, err := os.Stat(arg); err != nil {
			fmt.Println("File not found:", arg)
			return
		}
		files = append(files, arg)
	}

	if len(files) != 0 {
		current := -1
		for _, file := range files {
			current += 1
			key := genKey(file, current)
			uploaded := uploader(key, file)
			urls = append(urls, uploaded)
		}
		for _, url := range urls {
			fmt.Print(url)
		}
	}

}

func uploader(key string, file string) string {
	url := RootUrl + key
	reqTime := time.Now().Format(http.TimeFormat)
	sha := sha256.New()
	_, err := sha.Write([]byte(reqTime + key + uploaderSecret))
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	auth := hex.EncodeToString(sha.Sum(nil))
	data, err := os.OpenFile(file, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("file error:", err)
		return ""
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(data)
	if err != nil {
		fmt.Println("body error:", err)
		return ""
	}
	headers := http.Header{}
	headers.Set("Date", reqTime)
	headers.Set("Authorization", "Bearer "+auth)
	headers.Set("Content-Type", "application/octet-stream")
	headers.Set("Content-Length", strconv.Itoa(len(file)))
	request, err := http.NewRequest("PUT", url, bytes.NewReader(body.Bytes()))
	request.Header = headers
	rsp, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	if rsp.StatusCode != 200 {
		fmt.Println("Error:", rsp.StatusCode)
		return ""
	}
	var responseUrl map[string]string
	err = json.NewDecoder(rsp.Body).Decode(&responseUrl)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	err = rsp.Body.Close()
	if err != nil {
		return ""
	}
	return responseUrl["url"]
}

func genKey(file string, index int) string {
	sh := sha1.New()
	f, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	fb := bytes.Buffer{}
	_, err = fb.ReadFrom(f)
	if err != nil {
		return ""
	}
	_, err = sh.Write(fb.Bytes())
	if err != nil {
		return ""
	}
	hash := hex.EncodeToString(sh.Sum(nil))
	t := time.Now().Format("2006/01/02")
	extension := filepath.Ext(file)
	filePath := fmt.Sprintf("images/%s/%s/%s%s", t, strconv.Itoa(index), hash, extension)
	return filePath
}
