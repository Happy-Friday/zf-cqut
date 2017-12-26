package cqut

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

//Cqut 格式化的教务系统数据获取器
type Cqut struct {
	query *cqutQuery
}

//NewCqut创建一个Cqut数据获取对象
func NewCqut(username, password string) *Cqut {
	return &Cqut{
		query: newCqutQuery(username, password),
	}
}

//Initialize 初始化请求,以用来登陆
func (c *Cqut) Initialize() {
	c.query.initialize()
}

//返回筛选表格得到的map数据
func tableOfGrades(doc *goquery.Document) []map[string]string {
	if doc == nil {
		return nil
	}

	var attrs []string
	infos := make([]map[string]string, 0)
	doc.Find("#divNotPs tbody tr").Each(func(i int, s *goquery.Selection) {
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
					infos = append(infos, info)
				}
			})
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
	var doc *goquery.Document
	if pre {
		c.query.queryCountPre()
	}
	switch len(params) {
	case 2:
		doc, _ = c.query.queryCountNoPre(BtnXq, params[0], params[1]);
	case 1:
		doc, _ = c.query.queryCountNoPre(BtnXn, params[0])
	case 0:
		doc, _ = c.query.queryCountNoPre(BtnZcj)
	}

	return tableOfGrades(doc)
}

//GetGradesPoint 获取某一学年的绩点
//pre 是否进预处理
//params[0]学年
//params[1]学期
func (c *Cqut) GetGradesPoint(pre bool, params ...string) (string, error) {
	var doc *goquery.Document
	var err error
	if pre {
		c.query.queryCountPre()
	}
	switch len(params) {
	case 1:
		doc, err = c.query.queryCountNoPre(BtnCount, params[0]);
	case 0:
		doc, err = c.query.queryCountNoPre(BtnCount)
	}
	if err != nil {
		return "", err
	}
	gp := DecodeGbk(doc.Find("#pjxfjd").Text())
	gp = gp[strings.LastIndex(gp, "：")+len("："):]
	return gp, nil
}

//GetGradesPoint 获取某一学生所有学年的绩点
//pre 是否进预处理
func (c *Cqut) GetGradesPoints(pre bool) map[string]interface{} {
	gps := make(map[string]interface{})
	if pre {
		c.query.queryCountPre()
	}
	for _, term := range c.query.terms {
		if gp, err := c.GetGradesPoint(false, term); err == nil {
			gps[term] = gp
		}
	}
	return gps
}

//分析学生的课表，生成对应的map
func tableOfCourses(doc *goquery.Document) map[string]interface{} {
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
	ct["lessons"] = lessons
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
	var doc *goquery.Document
	var err error
	if pre {
		doc, err = c.query.queryCoursesPre()
	}

	switch len(params) {
	case 2:
		doc, err = c.query.queryCoursesNoPre(params[0], params[1])
	case 1:
		fallthrough
	case 0:
		doc, err = c.query.queryCoursesPre()
	}

	if err != nil {
		return nil
	}

	return tableOfCourses(doc)
}
