package cqut

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"encoding/json"
	"log"
	"sync"
)

//需要计算的学期
var terms = []string{"1", "2"}

//Cqut 格式化的教务系统数据获取器
type Cqut struct {
	//因为有csrf机制的存在，不能并发的访问教务系统的数据
	//添加一个互斥锁，防止并发时同时请求
	jwxt  *sync.Mutex
	query *cqutQuery
	//Info  *User
}

//NewCqut创建一个Cqut数据获取对象
func NewCqut(username, password string) *Cqut {
	return &Cqut{
		//Info:  NewUser(username, password),
		jwxt:  &sync.Mutex{},
		query: newCqutQuery(username, password),
	}
}

//Initialize 初始化请求,以用来登陆
func (c *Cqut) Initialize() error {
	return c.query.initialize()
}

//func (c *Cqut) GetGrades(force bool, pre bool, params ...string) []map[string]string {
//	grades := c.Info.GradesCount["grades"]
//	if grades == nil || force {
//		tgrades := c.getGrades(pre, params...)
//		c.Info.GradesCount["grades"] = tgrades
//		return tgrades
//	}
//	return grades.([]map[string]string)
//}
//
//func (c *Cqut) GetGradesPoints(force bool, pre bool, params ...string) map[string]interface{} {
//	gps := c.Info.GradesCount["gps"]
//	if gps == nil || force {
//		tgps := c.getGradesPoints(pre)
//		c.Info.GradesCount["gps"] = tgps
//		return tgps
//	}
//	return gps.(map[string]interface{})
//}

//返回筛选表格得到的map数据
func formatGradesTable(doc *goquery.Document) []map[string]string {
	if doc == nil {
		return nil
	}
	var attrs []string
	infos := make([]map[string]string, 0)
	//ps: 在#divNotPs下面还有一个表单,不要找tbody
	doc.Find("#divNotPs .datelist tr").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("datelisthead") {
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				attrs = append(attrs, DecodeGbk(s.Text()))
			})
		} else {
			info := make(map[string]string)
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				v := s.Text()
				if strings.TrimSpace(v) != "" {
					info[attrs[i]] = DecodeGbk(v)
				}
			})
			infos = append(infos, info)
		}
	})
	return infos
}

/*GetGrades 获取所有科目的成绩
pre 是否预处理
params[0]学年
params[1]学期
	`json格式`
	[
	{
        "学分": "1.0",
        "学年": "2015-2016",
        "学期": "1",
        "开课学院": "计算机科学与工程学院",
        "成绩": "优秀",
        "绩点": "   4.50",
        "课程代码": "00799",
        "课程名称": "程序设计基础课程设计【计算机科学与技术】",
        "课程性质": "集中实践教学环节",
        "辅修标记": "0"
    },....
	]
*/
func (c *Cqut) GetGrades(pre bool, params ...string) []map[string]string {
	c.jwxt.Lock()
	defer c.jwxt.Unlock()

	var doc *goquery.Document
	if pre {
		c.query.queryCountPre()
	}
	switch len(params) {
	case 2:
		doc, _ = c.query.queryCountNoPre(BtnXq, params[0], params[1]);
	case 1:
		doc, _ = c.query.queryCountNoPre(BtnXn, params[0])
	default:
		doc, _ = c.query.queryCountNoPre(BtnZcj)
	}

	return formatGradesTable(doc)
}

//GetGradesPoint 获取某一学年的绩点
//pre 是否进预处理
//params[0]学年
//params[1]学期
func (c *Cqut) GetGradesPoint(pre bool, params ...string) (map[string]string, error) {
	var doc *goquery.Document
	var err error
	if pre {
		c.query.queryCountPre()
	}
	switch len(params) {
	case 2:
		doc, err = c.query.queryCountNoPre(BtnCount, params[0], params[1]);
	case 1:
		doc, err = c.query.queryCountNoPre(BtnCount, params[0]);
	default:
		doc, err = c.query.queryCountNoPre(BtnCount)
	}
	if err != nil {
		return nil, err
	}
	gp := DecodeGbk(doc.Find("#pjxfjd").Text())
	gp = gp[strings.LastIndex(gp, "：")+len("："):]
	xfgpsum := DecodeGbk(doc.Find("#xfjdzh").Text())
	xfgpsum = xfgpsum[strings.LastIndex(xfgpsum, "：")+len("："):]
	return map[string]string{
		"gp":      gp,
		"xfgpsum": xfgpsum,
	}, nil
}

//GetGradesPoint 获取某一学生所有学年的绩点
//pre 是否进预处理
func (c *Cqut) GetGradesPoints(pre bool) map[string]interface{} {
	c.jwxt.Lock()
	defer c.jwxt.Unlock()

	gps := make(map[string]interface{})
	if pre {
		c.query.queryCountPre()
	}
	for _, year := range c.query.years {
		tgp := make([]map[string]string, len(terms))
		for i, term := range terms {
			if gp, err := c.GetGradesPoint(false, year, term); err == nil {
				tgp[i] = gp
			}
		}
		log.Println(tgp)
		gps[year] = tgp
	}
	return gps
}

//分析学生的课表，生成对应的map
func formatCoursesTable(doc *goquery.Document) map[string]interface{} {
	ct := make(map[string]interface{})
	lessons := make([][]string, 7)
	doc.Find("#Table1 tr").Each(func(i int, s *goquery.Selection) {
		if i >= 2 && i%2 == 0 {
			index := 0
			s.Find("td").Each(func(j int, s *goquery.Selection) {
				ok := false
				//过滤掉2, 6, 10 是因为有额外的标签(早上，下午，晚上)
				if i == 2 || i == 6 || i == 10 {
					ok = j >= 2
				} else {
					//每一行的第一个都是写的第N节课
					ok = j >= 1
				}
				if ok {
					if v, _ := s.Html(); !isEmpty(v) {
						v = DecodeGbk(v)
						for _, ls := range strings.Split(v, "<br/><br/>") {
							lessons[index] =
								append(lessons[index], strings.Replace(ls, "<br/>", "", -1))
						}
					}
					index++
				}

			})
		}
	})
	ct["coursesTable"] = lessons
	return ct
}

/*GetCoursesTable 获取学生的课表
/pre 是否预处理
/params[0] 学年
/params[1]	学期
/	json格式
	{
    "lessons": [
        ["星期1的课程列表"],
        ["星期2的课程列表"],
        ["星期3的课程列表"],
        ["星期4的课程列表"],
        ["星期5的课程列表"],
        ["星期6的课程列表"],
        ["星期日的课程列表"],
    ]
	Or
	{
	"lesson":[
		nil,nil,nil,nil,nil,nil,nil
	]
	}
}
*/
func (c *Cqut) GetCoursesTable(pre bool, params ...string) map[string]interface{} {
	c.jwxt.Lock()
	defer c.jwxt.Unlock()
	var doc *goquery.Document
	var err error
	if pre {
		doc, err = c.query.queryCoursesPre()
	}

	switch len(params) {
	case 2:
		doc, err = c.query.queryCoursesNoPre(params[0], params[1])
	default:
		doc, err = c.query.queryCoursesPre()
	}

	if err != nil {
		return nil
	}

	return formatCoursesTable(doc)
}

//格式化不使用sqid的查询结果
func formatUserInfo1(doc *goquery.Document) map[string]interface{} {
	text := DecodeGbk(doc.Text())
	var tInfos map[string]interface{}
	json.Unmarshal([]byte(text), &tInfos)
	if tInfos == nil {
		return nil
	}
	infos := make(map[string]interface{})
	for k, v := range tInfos {
		if v != nil {
			infos[k] = v
		}
	}
	return infos
}

//格式化使用sqid的查询结果，并和之前的结果相结合
func formatUserInfo2(doc *goquery.Document, infos map[string]interface{}) map[string]interface{} {
	text := DecodeGbk(doc.Text())
	var tInfos []map[string]interface{}
	json.Unmarshal([]byte(text), &tInfos)
	if tInfos == nil {
		return nil
	}
	for _, tInfo := range tInfos {
		infos[tInfo["zd"].(string)] = tInfo["zdz"]
	}
	return infos
}

//GetUserInfo 获取用户的个人详细信息
func (c *Cqut) GetUserInfo() map[string]interface{} {
	doc1, doc2, err := c.query.queryUserInfo()
	if err != nil {
		return nil
	}

	infos := formatUserInfo1(doc1)
	if infos == nil {
		return nil
	}

	if doc2 != nil {
		formatUserInfo2(doc2, infos)
	}

	return infos
}
