# zf-cqut
提供理工大学教务系统的部分api

##  功能

目前已有的功能

* 获取学生个人新信息
* 获取学生课表
* 获取学生绩点和绩点成绩总和
* 获取学生照片
* 获取学生课程和成绩

## 介绍

``Cqut``格式化查询结构, 负责查询数据, 并且返回一个的map结构

``Cqut``对象的查询中有一些包含预处理选项，预处理是在访问特定数据前初始化请求参数, 可以利用预处理选项降低请求服务器的次数. 比如你第一次操作访问了学生2016-2017的的成绩，下一次你想访问2017-2018的成绩，这时候就可以设置预处理为false, 不进行请求参数初始化, 因为你之前已经执行了一次访问成绩的操作了。所有含有预处理选项都可以，通过这样要减少请求次数.

``Cqut``对象是不支持并发操作的，所以内置了一个互斥锁，防止并发访问丢失数据.

## 使用

1. 基本使用方法

```go
cq := cqut.NewCqut("学号", "密码")
//请求登陆
//err主要分类两类: WrongPwdOrName(账号或密码错误), 连接超时
if err := cq.Initialize(); err != nil {
  	return 
}

```

2. 获取学生个人信息

```go
//[]map[string]string
infos := cq.GetUserInfo()
```

3. 获取学生课表

```go
/*
	[0] pre是否预处理
	[1] params参数 (选填: 默认查询当前学期)
		params[0] 学年
		params[1] 学期
	成功返回 map[string]inteface
	失败返回 nil
	func (c *Cqut) GetCoursesTable(pre bool, params ...string) map[string]interface{}
*/
coursesTable := cq.GetCoursesTable(true, "2016-2017", "1")
```

4. 获取学生绩点和绩点学分总和

```go
/*
	[0] pre是否预处理
	[1] params参数 (选填: 默认查询当前学期)
		params[0] 学年
		params[1] 学期
	成功返回 map[string]string
	失败返回 nil
func (c *Cqut) GetGradesPoint(pre bool, params ...string) (map[string]string, error)
*/
gps := cq.GetGradesPoint(true, "2016-2017", "1")
//获取所有学年和学期的绩点和绩点学分总和
gpsAll := cq.GetGradesPoints()
```

5. 查看学生照片

```go
/*
	[0] params参数 (选填: 默认查询自己照片)
		params[0] 查询学号
	成功返回 []byte照片数据
	失败返回 nil
	func (c *Cqut) GetPhoto(params ...string) []byte 
*/
photo := cq.GetPhoto()
```

6. 查看学生课程及课程成绩

```go
/*
	[0] pre是否预处理
	[1] params参数 (选填: 默认查询历年所有成绩)
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
	失败返回
		nil
	func (c *Cqut) GetGrades(pre bool, params ...string) []map[string]string
*/
grades := cq.GetGrades(true, "2017-2018", "1")
```

