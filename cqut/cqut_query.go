package cqut

import (
	"net/http"
	"net/http/cookiejar"
	"io"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strings"
	"errors"
	"log"
	"time"
	"fmt"
)

/**
Todo :
1. Get response of courses
2. Get response of all grades
3. Get response of plans
 */
const (
	GET       = "GET"
	POST      = "POST"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.40"
)

const (
	LoginPortal1     = "http://i.cqut.edu.cn/"
	LoginPortal2     = "http://i.cqut.edu.cn/zfca/login?service=http%3A%2F%2Fi.cqut.edu.cn%2Fportal.do"
	Login            = "http://i.cqut.edu.cn/zfca/login?service=http%3A%2F%2Fi.cqut.edu.cn%2Fportal.do"
	JwxtLink         = "http://i.cqut.edu.cn/zfca?yhlx=student&login=0122579031373493728&url=xs_main.aspx"
	XgxtLink         = "http://i.cqut.edu.cn/zfca?yhlx=student&login=122579031373493679&url=stuPage.jsp"
	XgxtInfoNoSqid   = "http://xgxt.i.cqut.edu.cn/xgxt/xsxx_xsxxxg.do?method=xgsq&type=query&timestamp="
	XgxtInfoWithSqid = "http://xgxt.i.cqut.edu.cn/xgxt/xsxx_xsxxxg.do?method=getXgzdList&timestamp="
)

const (
	BtnZcj   = "btn_zcj" //list all lesson's grades include grade point
	BtnXq    = "btn_xq"  //list all lesson's grades by term
	BtnXn    = "btn_xn"  //list all lesson's grades by year
	BtnCount = "Button1" //count grades by year and term
)

type cqutQuery struct {
	username string
	password string
	terms []string
	*cqut
}

func newCqutQuery(username, password string) *cqutQuery {
	return &cqutQuery{
		username: username,
		password: password,
		cqut:     newCqut(),
	}
}

func (c *cqutQuery) initialize() {
	log.Println("start to login system.....")
	c.login(c.username, c.password)
	log.Println("start to load Jwxt cookie.....")
	c.loginJwxt()
	log.Println("start to load Xgxt cookie.....")
	c.loginXgxt()
}

//模拟MVC将csrf标签保存在map里面
func (c *cqutQuery) updateJwxtTokens(rep *http.Response) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromResponse(rep);
	if err != nil {
		return nil, err
	}
	if v, ok := doc.Find("input[name=__EVENTTARGET]").Attr("value"); ok {
		c.jwxtTokens.Set("__EVENTTARGET", v)

	}
	if v, ok := doc.Find("input[name=__EVENTARGUMENT]").Attr("value"); ok {
		c.jwxtTokens.Set("__EVENTARGUMENT", v)

	}
	if v, ok := doc.Find("input[name=__VIEWSTATE]").Attr("value"); ok {
		c.jwxtTokens.Set("__VIEWSTATE", v)
	}
	return doc, nil
}
//生成一个含有上一次请求的csrf数据的表单
func (c *cqutQuery) formValues() url.Values {
	return url.Values{
		"__EVENTTARGET":   {c.jwxtTokens.Get("__EVENTTARGET")},
		"__EVENTARGUMENT": {c.jwxtTokens.Get("__EVENTARGUMENT")},
		"__VIEWSTATE":     {c.jwxtTokens.Get("__VIEWSTATE")},
	}
}

//用于获取查询当前学期的课程，并且初始化__VIEWSTATE
//这个方法必须在queryCoursesNoPre之前调用
func (c *cqutQuery) queryCoursesPre() (*goquery.Document, error) {
	req := commonRequest(GET, c.jwxtLinks["courses"], nil)
	rep, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return c.updateJwxtTokens(rep)
}
//查询指定学年和学期的课表, 这之前必须得调用queryCoursesPre初始化__VIEWSTATE
func (c *cqutQuery) queryCoursesNoPre(year, term string) (*goquery.Document, error) {
	v := c.formValues()
	//设置学年
	v.Set("xnd", year)
	//设置学期
	v.Set("xqd", term)
	req := commonRequest(POST, c.jwxtLinks["courses"], strings.NewReader(v.Encode()))
	rep, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return c.updateJwxtTokens(rep)
}
//查询指定学年和学期的课表，并且自动进行预处理
func (c *cqutQuery) queryCoursesWithPre(year, term string)(*goquery.Document, error) {
	c.queryCoursesPre()
	return c.queryCoursesNoPre(year, term)
}

//查询成绩, 并且初始化__VIEWSTATE
//你必须在queryGradesDetail之前调用
//return
func (c *cqutQuery) queryGrades() (*goquery.Document, error) {
	req := commonRequest(GET, c.jwxtLinks["grades"], nil)
	rep, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return c.updateJwxtTokens(rep)
}

//查询指定学期的成绩
func (c *cqutQuery) queryGradesDetail(year, term string) (*goquery.Document, error) {
	v := c.formValues()
	//设置学年
	v.Set("ddlxn", year)
	//设置学期
	v.Set("ddlxq", term)
	//设置按钮，虽然没实际意义，但是还是要带上才能成功
	v.Set("btnCx", "must need")
	req := commonRequest(POST, c.jwxtLinks["grades"], strings.NewReader(v.Encode()))
	rep, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return c.updateJwxtTokens(rep)
}
//在预加载成绩统计的时候顺便获取学生的学期
func (c *cqutQuery) setTerms(doc *goquery.Document) {
	if c.terms == nil {
		c.terms = make([]string, 0)
		doc.Find("#ddlXN option").Each(func(i int, s *goquery.Selection) {
			if attr, ok := s.Attr("value"); ok && !isEmpty(attr){
				c.terms = append(c.terms, attr)
			}
		})
	}
}

//QueryCountPre is way to  get the token called __VIEWSTATE
//You must invoke it before invoking QueryCount first
func (c *cqutQuery) queryCountPre() (*goquery.Document, error) {
	req := commonRequest(GET, c.jwxtLinks["count"], nil)

	rep, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := c.updateJwxtTokens(rep)
	if err != nil {
		return nil, err
	}
	//截取学生的所有学期
	c.setTerms(doc)

	return doc, err
}

//QueryCount will query grades of term、year、ever, and count of grades
//params
//	params[0] type of query, details are in (BtnZcj, BtnXq, BtnXn, BtnCount)
//  params[1] year of query
// 	params[2] term of query
//  params[3] type of lession's property
func (c *cqutQuery) queryCountNoPre(params ...string) (*goquery.Document, error) {
	v := c.formValues()
	v.Set("ddl_kcxz", "");
	v.Set("ddlXQ", "");
	v.Set("ddlXN", "");
	v.Set("hidLanguage", "")
	if (len(params) == 0) {
		v.Set(BtnZcj, "")
	} else {
		switch len(params) {
		case 4:
			v.Set("ddl_kcxz", params[3]); fallthrough
		case 3:
			v.Set("ddlXQ", params[2]); fallthrough
		case 2:
			v.Set("ddlXN", params[1]); fallthrough
		case 1:
			v.Set(params[0], "");
		}
	}
	req := commonRequest(POST, c.jwxtLinks["count"], strings.NewReader(v.Encode()))
	rep, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return c.updateJwxtTokens(rep)
}

func (c *cqutQuery) queryCountWithPre(params ...string) (*goquery.Document, error) {
	c.queryCountPre()
	return c.queryCountNoPre(params...)
}

func (c *cqutQuery) querySqid() (string, bool) {
	req := commonRequest(POST, "http://xgxt.i.cqut.edu.cn/xgxt/xsxx_xsxxxg_xgsq.do", nil)
	rep, err := c.client.Do(req)
	if err != nil {
		return "", false
	}
	doc, err := goquery.NewDocumentFromResponse(rep)
	if err != nil {
		return "", false
	}
	return doc.Find("input[name=shzSqid]").Attr("value")
}

func (c *cqutQuery) queryUserInfo() (*goquery.Document, *goquery.Document, error) {
	sqid, exist := c.querySqid()
	if !exist {
		return nil, nil, errors.New("Not found Sqid")
	}
	var (
		doc1 *goquery.Document
		doc2 *goquery.Document
	)

	u := fmt.Sprintf("%s%d", XgxtInfoNoSqid, time.Now().Nanosecond())
	req := commonRequest(POST, u, strings.NewReader(url.Values{"xh": {c.username}}.Encode()))
	rep, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	doc1, _ = goquery.NewDocumentFromResponse(rep)

	if strings.TrimSpace(sqid) != "" {
		u = fmt.Sprintf("%s%d", XgxtInfoWithSqid, time.Now().Nanosecond())
		req = commonRequest(POST, u, strings.NewReader(url.Values{"sqid": {sqid}}.Encode()))
		rep, err = c.client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		doc2, _ = goquery.NewDocumentFromResponse(rep)
	}

	return doc1, doc2, err
}

//cqut is a embedded struct to load import cookies at initialization time
//import cookie include cookie of login、jwxt、xgxt
type cqut struct {
	client     *http.Client
	ticket     string
	jwxtURL    string
	jwxtTokens url.Values
	jwxtLinks  map[string]string
}

func newCqut() *cqut {
	jar, _ := cookiejar.New(nil)
	return &cqut{
		client: &http.Client{
			Jar: jar,
		},
		jwxtTokens: url.Values{},
	}
}

//Generate a common request
func commonRequest(method string, url string, reader io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, reader);
	req.Header.Add("User-Agent", UserAgent);
	req.Header.Add("Connection", "keep-alive");
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req;
}

//Set whether request follows location
//When b is true, It will be
func (c *cqut) setCheckDirect(b bool) {
	c.client.CheckRedirect = func(req *http.Request, vias []*http.Request) error {
		if b {
			return nil
		}
		return http.ErrUseLastResponse
	}
}

//Load the first cookie before login
func (c *cqut) loadLoginCookie1() {
	req := commonRequest(GET, LoginPortal1, nil)
	c.client.Do(req);
}

//Load the second cookie and get the _csrf(lt) before login
func (c *cqut) loadLoginCookie2() (string, bool) {
	req := commonRequest(GET, LoginPortal2, nil)
	rep, err := c.client.Do(req)
	if err != nil {
		return "reponse failed", false
	}

	doc, err := goquery.NewDocumentFromResponse(rep)
	if err != nil {
		return "generate goquery document failed", false
	}

	return doc.Find("input[name=lt]").Attr("value")
}

//Login the server to get the import cookie
//Must set Reference, or you cannot get right result
func (c *cqut) login(username, password string) (*http.Response, error) {
	log.Println("load cookie1")
	c.loadLoginCookie1()
	log.Println("load cookie2")
	lt, ok := c.loadLoginCookie2()
	if !ok {
		return nil, errors.New(lt)
	}
	v := url.Values{
		"lt":              {lt},
		"ip":              {""},
		"username":        {username},
		"password":        {password},
		"_eventId":        {"submit"},
		"useValidateCode": {"0"},
		"isremenberme":    {"0"},
		"losetime":        {"30"},
	}
	req := commonRequest(POST, Login, strings.NewReader(v.Encode()))
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//Referer: Must be with it
	req.Header.Set("Referer", "http://i.cqut.edu.cn/zfca/login")
	req.Header.Set("Origin", "http://i.cqut.edu.cn")
	req.Header.Set("Host", "i.cqut.edu.cn")
	log.Println("start to login..")
	rep, err := c.client.Do(req);
	if err != nil {
		return nil, err
	}
	c.ticket = rep.Request.URL.Query().Get("ticket")
	return rep, nil
}

//登陆教务系统，获取cookies，并且截取重要的链接
func (c *cqut) loginJwxt() error {
	req := commonRequest(GET, JwxtLink, nil)
	rep, err := c.client.Do(req)
	if err != nil {
		return err
	}
	url := rep.Request.URL.String()
	c.jwxtURL = url[:strings.LastIndex(url, "/")]
	c.jwxtLinks = analyseJwxtList(c.jwxtURL, rep)
	return nil
}

//分析正方的Jwxt的菜单列表，并且截取一部分链接
//比如成绩统计,培养计划，课表
func analyseJwxtList(base string, rep *http.Response) map[string]string {
	lists := make(map[string]string)
	doc, _ := goquery.NewDocumentFromResponse(rep)
	doc.Find(".nav .top").Each(func(i int, selection *goquery.Selection) {
		topLink := DecodeGbk(selection.Find(".top_link").Text())
		if strings.TrimSpace(topLink) == "信息查询" {
			selection.Find(".sub li").Each(func(i int, selection *goquery.Selection) {
				if a, ok := selection.Find("a").Attr("href"); ok {
					a = base + "/" + DecodeGbk(a)
					switch i {
					case 0:
						lists["courses"] = a;
					case 1:
						lists["count"] = a;
					case 3:
						lists["plans"] = a;
					case 4:
						lists["grades"] = a;
					}
				}
			})
		}
	})
	return lists
}

//登入学工系统，获取cookie
func (c *cqut) loginXgxt() error {
	req := commonRequest(GET, XgxtLink, nil)
	_, err := c.client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
