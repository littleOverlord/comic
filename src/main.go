package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"ni/logger"

	"github.com/PuerkitoBio/goquery"
)

var waitCh = make(chan int, 20)

type imgContrast [][2]string

func main() {
	defer func() {
		if p := recover(); p != nil {
			fmt.Println(p)
			logger.Error(p.(string))
		}
	}()
	// itemInfo("http://www.huhudm.com/huhu31897.html", "Dr.STONE")
	// singlePage("http://www.huhudm.com/hu353720/1.html?s=6", 1, "./res/Dr.STONE")
	// load volume
	// findPage("http://www.huhudm.com/comic/")
	// load image
	onlyLoadImg()
	select {}
}

func findPage(_url string) {
	res, err := http.Get(_url)
	defer func() {
		if err != nil {
			logger.Error("func findPage \n" + _url + "\n" + err.Error())
		}
	}()
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logger.Error(fmt.Sprintf("findPage status code error: %d %s \n url: %s", res.StatusCode, res.Status, _url))
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}

	// Find the review items
	b := doc.Find(".cComicPageChange b").Last()
	// Each(func(i int, s *goquery.Selection) {
	// 	fmt.Println(s.Text())
	// })
	pageAll, err := strconv.Atoi(b.Text())
	if err != nil {
		return
	}
	// fmt.Println(pageAll)
	findItem(pageAll)
}

func findItem(all int) {
	for i := 1; i <= all; i++ {
		waitCh <- 1
		go func(p int) {
			find(p)
		}(i)
	}
}

func dealItem(s *goquery.Selection) {
	a := s.Find("a").First()
	url, ok := a.Attr("href")
	if !ok {
		fmt.Println("don't find href")
	}
	img, ok := s.Find("img").Attr("src")
	if !ok {
		fmt.Println("don't find src")
	}
	title := a.Text()
	if checkFileIsExist("./res/"+title+"/info.json") != 0 {
		<-waitCh
		return
	}
	err := os.MkdirAll("./res/"+title, os.ModePerm)
	if err != nil {
		logger.Error(err.Error())
	}
	ext := path.Ext(img)

	writeImg(img, "./res/"+title+"/title"+ext)
	waitCh <- 1
	go itemInfo("http://www.huhudm.com"+url, title)
}

func itemInfo(url string, name string) {
	res, err := http.Get(url)
	defer func() {
		if err != nil {
			logger.Error("func itemInfo \n" + url + "\n" + err.Error())
		}
		<-waitCh
	}()
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logger.Error(fmt.Sprintf("itemInfo status code error: %d %s \n url: %s", res.StatusCode, res.Status, url))
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	// fmt.Println(name)
	var (
		author string
		status string
		volume string
		des    string
	)
	// Find the review items
	doc.Find("#about_kit ul li").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if i == 1 {
			author = s.Text()
			author = strings.ReplaceAll(author, "作者:", "")
			// fmt.Println(author)
		}
		if i == 2 {
			status = s.Text()
			status = strings.ReplaceAll(status, "状态:", "")
			// fmt.Println(status)
		}
		if i == 3 {
			volume = s.Text()
			volume = strings.ReplaceAll(volume, "集数:", "")
			// fmt.Println(volume)
		}
		if i == 7 {
			des = s.Text()
			des = strings.ReplaceAll(des, "简介:", "")
			des = strings.ReplaceAll(des, "\n", "")
			// fmt.Println(des)
		}
	})
	f, err := os.Create("./res/" + name + "/info.json") //创建文件
	if err != nil {
		return
	}

	var list = "{"
	lis := doc.Find(".cVolUl li")
	fmt.Println(lis.Length())
	lis.Each(func(i int, s *goquery.Selection) {
		nd := s.Find("a")
		_url, ok := nd.Attr("href")
		if !ok {
			return
		}
		title, ok := nd.Attr("title")
		if !ok {
			return
		}
		reg := regexp.MustCompile(`.+ `)
		title = reg.ReplaceAllString(title, "")
		// title = strings.ReplaceAll(title, name+" ", "")
		err := os.MkdirAll("./res/"+name+"/"+title, os.ModePerm)
		if err != nil {
			logger.Error(err.Error())
		}
		if i == 0 {
			list = list + (`"./res/` + name + "/" + title + `":"` + _url + `"`)
		} else {
			list = list + (`,"./res/` + name + "/" + title + `":"` + _url + `"`)
		}
		// go findVolume("http://www.huhudm.com"+_url, "./res/"+name+"/"+title)
	})
	list += "}"
	_, err = io.WriteString(f, fmt.Sprintf(`{
		"title": "%s",
		"author": "%s",
		"status": "%s",
		"volume": "%s",
		"des":"%s",
		"list":%s
	}`, name, author, status, volume, des, list)) //写入文件(字符串)
	fmt.Println(name + " ok!")
	if err != nil {
		return
	}
}

func findVolume(url string, _path string) {
	res, err := http.Get(url)
	defer func() {
		if err != nil {
			logger.Error("func findVolume \n" + url + "\n" + err.Error())
		}
	}()
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logger.Error(fmt.Sprintf("findVolume status code error: %d %s \n url: %s", res.StatusCode, res.Status, url))
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	count, _ := doc.Find("#hdPageCount").Attr("value")

	c, err := strconv.Atoi(count)
	fmt.Println(c)
	if err != nil {
		return
	}
	for i := 1; i <= c; i++ {
		// waitCh <- 1
		go singlePage(url, i, _path)
	}
}

func singlePage(url string, page int, _path string) {
	reg := regexp.MustCompile(`/\d+\.html`)
	_url := reg.ReplaceAllString(url, fmt.Sprintf("/%d.html", page))
	fmt.Println(_url)
	res, err := http.Get(_url)
	defer func() {
		if err != nil {
			logger.Error("func itemInfo \n" + url + "\n" + err.Error())
		}
	}()
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logger.Error(fmt.Sprintf("singlePage status code error: %d %s \n url: %s", res.StatusCode, res.Status, url))
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}

	// Find the review items
	src, _ := doc.Find("#iBody img").Last().Attr("src")
	ext := path.Ext(src)
	go writeImg(src, fmt.Sprintf("%s/%d%s", _path, page, ext))
}

func find(page int) {
	// Request the HTML page.
	_url := fmt.Sprintf("http://www.huhudm.com/comic/%d.html", page)
	fmt.Println(_url)
	res, err := http.Get(_url)
	defer func() {
		if err != nil {
			logger.Error("func find \n" + _url + "\n" + err.Error())
		}

	}()
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logger.Error(fmt.Sprintf("find status code error: %d %s \n url: %s", res.StatusCode, res.Status, _url))
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	<-waitCh
	// Find the review items
	doc.Find(".cComicList li").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		waitCh <- 1
		go dealItem(s)
	})
}

func writeImg(url string, path string) {
	// fmt.Println(url + " :: " + path)
	var err error
	defer func() {
		if err != nil {
			logger.Error("func writeImg \n" + url + "\n" + err.Error())
		}
		<-waitCh
	}()
	cf := checkFileIsExist(path)
	if cf == 1 {
		return
		// file, err = os.OpenFile(path, os.O_APPEND, 0666) //打开文件
	}
	// if url != "" {
	// 	return
	// }
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	// 获得get请求响应的reader对象
	reader := bufio.NewReaderSize(res.Body, 32*1024)
	var file *os.File
	if cf == -1 {
		file, err = os.OpenFile(path, os.O_APPEND, 0666) //打开文件
	} else {
		file, err = os.Create(path)
	}
	if err != nil {
		return
	}
	// 获得文件的writer对象
	writer := bufio.NewWriter(file)

	written, _ := io.Copy(writer, reader)

	fmt.Printf("Total length: %d; path: %s \n", written, path)
}

func onlyLoadImg() {
	currentPath, err := os.Getwd()
	fmt.Println("currentPath :: " + currentPath)
	readNeedLoadImg(path.Join(currentPath, "loaded"))
	if err != nil {
		fmt.Println(err)
	}
}
func readNeedLoadImg(absDir string) {
	fmt.Println("absDir :: " + absDir)
	// table := make(map[string]imgContrast)
	dir, err := os.Open(absDir)
	if err != nil {
		panic("readNeedLoadImg os.Open error : " + err.Error())
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		panic("readNeedLoadImg Readdir error : " + err.Error())
	}
	// fmt.Println(files[0].Name())
	for _, file := range files {
		var imgc imgContrast
		err = parseJSON(path.Join(absDir, file.Name()), &imgc)
		if err != nil {
			panic("readNeedLoadImg parseJSON error : " + err.Error())
		}
		for _, img := range imgc {
			waitCh <- 1
			go writeImg(img[0], img[1])
		}
	}

	// table[files[0].Name()] = imgc
	// fmt.Println(table)
	// fmt.Println(len(files))
	// for _, file := range files {
	// 	// fmt.Println(file.Name())
	// 	var conf interface{}
	// 	err = parseJSON(path.Join(absDir, file.Name()), &conf)
	// 	if err != nil {
	// 		panic("loadConfig parseJSON error : " + err.Error())
	// 	}
	// 	fmt.Println(path.Join(rd, file.Name()), Table[path.Join(rd, file.Name())])
	// 	Table[path.Join(rd, file.Name())] = conf
	// }
}

// 读取文件并解析
func parseJSON(p string, c interface{}) error {
	file, err := os.Open(p)
	if err != nil {
		return err
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, c)
	if err != nil {
		return err
	}
	return nil
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) int {
	var exist = 1
	if f, err := os.Stat(filename); os.IsNotExist(err) {
		exist = 0
	} else if f.Size() < 50*1024 && f.Size() != 36974 {
		exist = -1
	}
	return exist
}
