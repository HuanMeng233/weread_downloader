package decrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/yeka/zip"
)

func initHeader(req *http.Request, vid, skey string) *http.Request {
	req.Header["accessToken"] = []string{skey}
	req.Header["vid"] = []string{vid}
	req.Header["baseapi"] = []string{"34"}
	req.Header["appver"] = []string{"7.5.0.10162554"}
	req.Header["User-Agent"] = []string{"WeRead/7.5.0 WRBrand/xiaomi Dalvik/2.1.0 (Linux; U; Android 14; 2304FPN6DC Build/UKQ1.230804.001)"}
	req.Header["osver"] = []string{"14"}
	req.Header["channelId"] = []string{"12"}
	req.Header["basever"] = []string{"7.5.0.10162554"}
	req.Header["Connection"] = []string{"Keep-Alive"}
	return req
}
func getKeyAndIV(vid int) ([]byte, []byte) {

	remapArr := [10]byte{0x2d, 0x50, 0x56, 0xd7, 0x72, 0x53, 0xbf, 0x22, 0xfb, 0x20}
	vidLen := len(strconv.Itoa(vid))
	vidRemap := make([]byte, vidLen)

	for i := 0; i < vidLen; i++ {
		vidRemap[i] = remapArr[strconv.Itoa(vid)[i]-'0']
	}
	key := make([]byte, 36)
	for i := 0; i < 36; i++ {
		key[i] = vidRemap[i%vidLen]
	}
	iv := make([]byte, 16)
	for i := 0; i < 16; i++ {
		iv[i] = key[i+7]
	}
	key = key[0:16]
	return key, iv
}

func getPassword(vid string, encryptKey string) string {
	vidInt, _ := strconv.Atoi(vid)
	key, iv := getKeyAndIV(vidInt)
	encryptData, _ := base64.StdEncoding.DecodeString(encryptKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, 16)
	blockMode.CryptBlocks(decryptedData, encryptData)
	pwdStr := ""
	for i := 0; i < len(decryptedData); i++ {
		if decryptedData[i] < 32 || decryptedData[i] > 126 {
			continue
		}
		pwdStr += string(decryptedData[i])
	}
	fmt.Println("pwd", pwdStr)
	return pwdStr
}

func MergeTxtBook(bookName, bookPath string) {
	type BookInfo struct {
		BookId            string `json:"bookId"`
		Synckey           int    `json:"synckey"`
		ChapterUpdateTime int    `json:"chapterUpdateTime"`
		Chapters          []struct {
			ChapterUid  int     `json:"chapterUid"`
			ChapterIdx  int     `json:"chapterIdx"`
			UpdateTime  int     `json:"updateTime"`
			Title       string  `json:"title"`
			WordCount   int     `json:"wordCount"`
			Price       float64 `json:"price"`
			IsMPChapter int     `json:"isMPChapter"`
			Paid        int     `json:"paid"`
		} `json:"chapters"`
	}
	f, err := os.Open(bookPath + "info.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var bookInfo BookInfo
	err = json.NewDecoder(f).Decode(&bookInfo)
	if err != nil {
		panic(err)
	}
	//创建目录
	os.Mkdir(bookPath+"看这里", os.ModePerm)
	//创建txt文件
	bookFile, _ := os.Create(bookPath + "看这里/" + bookName + ".txt")
	defer bookFile.Close()
	//读取章节信息
	for _, chapter := range bookInfo.Chapters {
		oldName := fmt.Sprintf("%s_%d_o", bookInfo.BookId, chapter.ChapterUid)
		newName := fmt.Sprintf("第%d章 %s", chapter.ChapterIdx, chapter.Title)
		//读取章节内容
		chapterFile, err := os.Open(bookPath + oldName)

		if err != nil {
			panic(err)
		}

		//写入章节内容
		bookFile.WriteString(newName + "\n\n\n")
		buf := make([]byte, 1024)
		for {
			n, err := chapterFile.Read(buf)
			if err != nil {
				break
			}
			bookFile.Write(buf[:n])
		}
		chapterFile.Close()
		bookFile.WriteString("\n\n\n")
	}

}

func MergePdfBook(bookName, bookPath string) {
	type BookInfo struct {
		BookId            string `json:"bookId"`
		Synckey           int    `json:"synckey"`
		ChapterUpdateTime int    `json:"chapterUpdateTime"`
		Chapters          []struct {
			ChapterUid  int      `json:"chapterUid"`
			ChapterIdx  int      `json:"chapterIdx"`
			UpdateTime  int      `json:"updateTime"`
			Title       string   `json:"title"`
			WordCount   int      `json:"wordCount"`
			Price       int      `json:"price"`
			IsMPChapter int      `json:"isMPChapter"`
			Paid        int      `json:"paid"`
			Level       int      `json:"level"`
			Files       []string `json:"files"`
			Anchors     []struct {
				Title  string `json:"title"`
				Anchor string `json:"anchor"`
				Level  int    `json:"level"`
			} `json:"anchors,omitempty"`
		} `json:"chapters"`
	}
	page := `<div style="page-break-after: always;"></div>`
	htmlBody := `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">

<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title></title>
    <link href="../Styles/stylesheets.css" rel="stylesheet" type="text/css" />
	<link href="../Styles/stylesheet.css" rel="stylesheet" type="text/css" />

  </head>
	<body>
	  </body>
</html>
	`

	f, err := os.Open(bookPath + "info.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var bookInfo BookInfo
	err = json.NewDecoder(f).Decode(&bookInfo)

	os.Mkdir(bookPath+"看这里", os.ModePerm)
	coverFile, _ := os.Open(bookPath + bookInfo.Chapters[0].Files[0])
	defer coverFile.Close()

	htmlDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))

	for _, chapter := range bookInfo.Chapters {
		for _, file := range chapter.Files {
			f, _ := os.Open(bookPath + file)
			doc, _ := goquery.NewDocumentFromReader(f)
			docHtml, _ := doc.Html()
			docHtml = html.UnescapeString(docHtml)
			defer f.Close()

			ddd, _ := goquery.NewDocumentFromReader(strings.NewReader(docHtml))
			htmlData, _ := ddd.Html()
			htmlData = html.UnescapeString(htmlData)
			bodyData := strings.Split(htmlData, "</head>")[1]
			bodyData = strings.Split(bodyData, "</html>")[0]
			bodyData = strings.ReplaceAll(bodyData, "body", "div")

			htmlDoc.Find("body").AppendHtml(bodyData)
			htmlDoc.Find("body").AppendHtml(page)

		}
	}
	//创建bookFile
	bookFile, _ := os.Create(bookPath + "看这里/" + bookName + ".html")
	defer bookFile.Close()
	//写入bookFile
	h, _ := htmlDoc.Html()

	bookFile.WriteString(html.UnescapeString(h))

}

func GetBookInfo(bookId, skey, vid string) (int64, int64, string, string) {

	url := "https://i.weread.qq.com/book/info?bookId=" + bookId + "&myzy=1&source=reading&teenmode=0"
	client := http.Client{}

	req, _ := http.NewRequest("GET", url, nil)
	req = initHeader(req, vid, skey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.StatusCode)
	defer resp.Body.Close()

	type Res struct {
		BookId         string  `json:"bookId"`
		Title          string  `json:"title"`
		Author         string  `json:"author"`
		Translator     string  `json:"translator"`
		Cover          string  `json:"cover"`
		Version        int64   `json:"version"`
		Format         string  `json:"format"`
		Type           int     `json:"type"`
		Price          float64 `json:"price"`
		OriginalPrice  int     `json:"originalPrice"`
		Soldout        int     `json:"soldout"`
		BookStatus     int     `json:"bookStatus"`
		PayType        int     `json:"payType"`
		Intro          string  `json:"intro"`
		CentPrice      int     `json:"centPrice"`
		Finished       int     `json:"finished"`
		MaxFreeChapter int     `json:"maxFreeChapter"`
		Free           int     `json:"free"`
		McardDiscount  int     `json:"mcardDiscount"`
		Ispub          int     `json:"ispub"`
		ExtraType      int     `json:"extra_type"`
		Cpid           int     `json:"cpid"`
		PublishTime    string  `json:"publishTime"`
		Category       string  `json:"category"`
		Categories     []struct {
			CategoryId    int    `json:"categoryId"`
			SubCategoryId int    `json:"subCategoryId"`
			CategoryType  int    `json:"categoryType"`
			Title         string `json:"title"`
		} `json:"categories"`
		HasLecture     int    `json:"hasLecture"`
		LPushName      string `json:"lPushName"`
		ShouldHideTTS  int    `json:"shouldHideTTS"`
		LastChapterIdx int64  `json:"lastChapterIdx"`
		PaperBook      struct {
			SkuId string `json:"skuId"`
		} `json:"paperBook"`
		BlockSaveImg          int     `json:"blockSaveImg"`
		Language              string  `json:"language"`
		HideUpdateTime        bool    `json:"hideUpdateTime"`
		IsEPUBComics          int     `json:"isEPUBComics"`
		PayingStatus          int     `json:"payingStatus"`
		ChapterSize           int64   `json:"chapterSize"`
		UpdateTime            int     `json:"updateTime"`
		OnTime                int     `json:"onTime"`
		LastChapterCreateTime int     `json:"lastChapterCreateTime"`
		UnitPrice             float64 `json:"unitPrice"`
		MarketType            int     `json:"marketType"`
		Isbn                  string  `json:"isbn"`
		Publisher             string  `json:"publisher"`
		TotalWords            int     `json:"totalWords"`
		PublishPrice          float64 `json:"publishPrice"`
		BookSize              int     `json:"bookSize"`
		Recommended           int     `json:"recommended"`
		LectureRecommended    int     `json:"lectureRecommended"`
		Follow                int     `json:"follow"`
		Secret                int     `json:"secret"`
		Offline               int     `json:"offline"`
		LectureOffline        int     `json:"lectureOffline"`
		FinishReading         int     `json:"finishReading"`
		HideReview            int     `json:"hideReview"`
		HideFriendMark        int     `json:"hideFriendMark"`
		Blacked               int     `json:"blacked"`
		IsAutoPay             int     `json:"isAutoPay"`
		Availables            int     `json:"availables"`
		Paid                  int     `json:"paid"`
		IsChapterPaid         int     `json:"isChapterPaid"`
		ShowLectureButton     int     `json:"showLectureButton"`
		Wxtts                 int     `json:"wxtts"`
		Myzy                  int     `json:"myzy"`
		MyzyPay               int     `json:"myzy_pay"`
		HasAuthorReview       int     `json:"hasAuthorReview"`
		Star                  int     `json:"star"`
		RatingCount           int     `json:"ratingCount"`
		RatingDetail          struct {
			One    int `json:"one"`
			Two    int `json:"two"`
			Three  int `json:"three"`
			Four   int `json:"four"`
			Five   int `json:"five"`
			Recent int `json:"recent"`
		} `json:"ratingDetail"`
		NewRating       int `json:"newRating"`
		NewRatingCount  int `json:"newRatingCount"`
		NewRatingDetail struct {
			Good     int    `json:"good"`
			Fair     int    `json:"fair"`
			Poor     int    `json:"poor"`
			Recent   int    `json:"recent"`
			MyRating string `json:"myRating"`
			Title    string `json:"title"`
		} `json:"newRatingDetail"`
		Ranklist struct {
			Seq           int    `json:"seq"`
			CategoryId    string `json:"categoryId"`
			CategoryName  string `json:"categoryName"`
			StoreSubType  int    `json:"storeSubType"`
			Scheme        string `json:"scheme"`
			RanklistCover struct {
				Tinycode                 string `json:"tinycode"`
				ChartTitle               string `json:"chart_title"`
				ChartDetailTitle         string `json:"chart_detail_title"`
				ChartDetailTitleDark     string `json:"chart_detail_title_dark"`
				ChartShareTitle          string `json:"chart_share_title"`
				ChartShareLogo           string `json:"chart_share_logo"`
				ChartBookDetialIcon      string `json:"chart_book_detial_icon"`
				ChartTag                 string `json:"chart_tag"`
				EinkChartTitle           string `json:"eink_chart_title"`
				ChartTitleMain           string `json:"chart_title_main"`
				ChartDetailTitleMain     string `json:"chart_detail_title_main"`
				ChartDetailTitleDarkMain string `json:"chart_detail_title_dark_main"`
				ChartBackgroundColor1    string `json:"chart_background_color_1"`
				ChartBackgroundColor2    string `json:"chart_background_color_2"`
				ChartTitleHeight         int    `json:"chart_title_height"`
				ChartTitleWidth          int    `json:"chart_title_width"`
				Desc                     string `json:"desc"`
			} `json:"ranklistCover"`
		} `json:"ranklist"`
		CopyrightInfo struct {
			Id      int    `json:"id"`
			Name    string `json:"name"`
			UserVid int    `json:"userVid"`
			Role    int    `json:"role"`
			Avatar  string `json:"avatar"`
		} `json:"copyrightInfo"`
		AuthorSeg []struct {
			Words     string `json:"words"`
			Highlight int    `json:"highlight"`
			AuthorId  string `json:"authorId"`
		} `json:"authorSeg"`
		TranslatorSeg []struct {
			Words     string `json:"words"`
			Highlight int    `json:"highlight"`
			AuthorId  string `json:"authorId"`
		} `json:"translatorSeg"`
		CoverBoxInfo struct {
			Blurhash string `json:"blurhash"`
			Colors   []struct {
				Key string `json:"key"`
				Hex string `json:"hex"`
			} `json:"colors"`
			DominateColor struct {
				Hex string    `json:"hex"`
				Hsv []float64 `json:"hsv"`
			} `json:"dominate_color"`
			CustomCover    string `json:"custom_cover"`
			CustomRecCover string `json:"custom_rec_cover"`
		} `json:"coverBoxInfo"`
	}

	var res Res
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		fmt.Println(err, "json")
	}
	bookChapterSize := res.ChapterSize + res.LastChapterIdx - 1
	bookVersion := res.Version
	bookFormat := res.Format
	fmt.Println(bookChapterSize, bookVersion, bookFormat)
	return bookChapterSize, bookVersion, bookFormat, res.Title
}
func DownloadBook(bookId, skey, vid string) string {
	fmt.Println("开始下载", bookId)
	bookChapterSize, bookVersion, bookFormat, bookName := GetBookInfo(bookId, skey, vid)
	url := fmt.Sprintf("https://i.weread.qq.com/book/chapterdownload?bookId=%s&chapters=0-%d&pf=wechat_wx-2001-android-100-weread&pfkey=pfKey&zoneId=1&bookVersion=%d&bookType=%s&quote=&release=1&stopAutoPayWhenBNE=1&preload=2&preview=0&offline=0&giftPayingCard=0&enVersion=7.5.0&modernVersion=7.5.0&teenmode=0", bookId, bookChapterSize, bookVersion, bookFormat)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req = initHeader(req, vid, skey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err, "do request")
		return "请求失败"
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return "登录超时,请清除登录数据，重新登录"
	}
	if resp.StatusCode == 402 {
		return "没有下载整本书的权限，请检查是否在微信读书中购买了整本书"
	}
	encryptKey := resp.Header.Get("encryptKey")
	pwdStr := getPassword(vid, encryptKey)
	os.MkdirAll("./book/"+vid+"/", os.ModePerm)
	f, err := os.Create("./book/" + vid + "/" + bookName + ".zip")
	if err != nil {
		fmt.Println(err, "create f")
		return "创建文件失败"
	}
	defer f.Close()
	//写出文件
	bookData, err := io.ReadAll(resp.Body)
	_, err = f.Write(bookData)
	if err != nil {
		fmt.Println(err, "write file")
		return "写出文件失败"
	}
	//解压文件
	zipReader, err := zip.NewReader(bytes.NewReader(bookData), int64(len(bookData)))
	if err != nil {
		fmt.Println(err, "new zip reader")
		return "解压文件失败"
	}
	for _, f := range zipReader.File {
		if f.IsEncrypted() {
			f.SetPassword(pwdStr)
		}
		r, err := f.Open()
		if err != nil {
			fmt.Println(err, "open file")
			return "打开文件失败"
		}
		fileName := "./book/" + vid + "/" + bookName + "/" + f.Name
		_, err = os.Stat(fileName)
		if err == nil {
			continue
		}
		dir := path.Dir(fileName)
		_, err = os.Stat(dir)
		if err != nil {
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				fmt.Println(err)
				return "创建文件夹失败"

			}
		}

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println(err)
			return "创建文件失败"
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Println(err)
			return "读取文件失败"
		}
		file.Write(b)
		file.Close()

	}
	//导出书籍
	if bookFormat == "epub" {
		MergePdfBook(bookName, "./book/"+vid+"/"+bookName+"/")
	} else {
		MergeTxtBook(bookName, "./book/"+vid+"/"+bookName+"/")
	}
	return "下载完成"
}
