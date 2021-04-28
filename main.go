package main

import (
	"bufio"
	"bytes"
	"html"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/color"
	"github.com/valyala/fasthttp"
)

type Song struct {
	Name   string
	Artist string
	Album  string
	Url    string
}

func main() {
	c := &fasthttp.Client{}
	var url string

	songs := make(map[string]Song)
	r := regexp.MustCompile(`"name": "(?P<name>.+)",.*\n.*"artist": "(?P<artist>.+)",.*\n.*"album": "(?P<album>.+)",.*\n.*"url": "(?P<url>.+)",.*`)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		url = scanner.Text()
		if !strings.HasPrefix(url, "http") {
			continue
		}

		code, body, err := c.Get(nil, url)
		if code != 200 && err != nil {
			log.Fatal("Error downloading: ", err)
		}

		//		/html/body/div[2]/h1/text()
		//		body > div:nth-child(4) > h1
		//		document.querySelector("body > div:nth-child(4) > h1")

		rdr := bytes.NewReader(body)
		gq, err := goquery.NewDocumentFromReader(rdr)
		if err != nil {
			log.Fatal("Error parsing html: ", err)
		}

		h1 := gq.Find("body > div:nth-child(4) > h1").Text()
		h1 = strings.TrimSpace(h1)
		st := strings.Split(h1, "\n")
		for i := range st {
			st[i] = strings.TrimSpace(st[i])
		}
		dn := strings.Join(st, " - ")

		res := r.FindAllSubmatch(body, -1)
		for i := range res {
			s := Song{
				Name:   html.UnescapeString(string(res[i][1])),
				Artist: html.UnescapeString(string(res[i][2])),
				Album:  html.UnescapeString(string(res[i][3])),
				Url:    string(res[i][4]),
			}
			songs[string(res[i][1])] = s
		}

		if err := os.Mkdir(dn, 0755); err != nil {
			color.Red.Println("Error create dir: ", err)
			continue
		}
		if err := os.Chdir(dn); err != nil {
			color.Red.Println("Error change dir: ", err)
			continue
		}
		for k, v := range songs {
			color.Printf("<yellow>Downloading</><white>: %s</>\n", k)
			if !strings.HasPrefix(v.Url, "http") {
				v.Url = "https://soundtracks.pro" + v.Url
			}
			//ariago.Aria(v.Url, "")
			Download(v.Url, c)
		}
		if err := os.Chdir(".."); err != nil {
			color.Red.Println("Error change dir to parrent: ", err)
		}
	}
}

func Download(u string, c *fasthttp.Client) {
	code, body, err := c.Get(nil, u)
	if code != 200 && err != nil {
		color.Red.Printf("Error downloading: %s\n", u)
		return
	}
	_, fn := path.Split(u)
	f, err := os.Create(fn)
	if err != nil {
		color.Red.Printf("Error writing file: %s\n", fn)
		return
	}
	defer f.Close()
	_, err = f.Write(body)
	if err != nil {
		color.Red.Printf("Error writing file: %s\n", fn)
	}
	color.Green.Println("Done.")
}
