package cqut

type User struct {
	Username     string
	Password     string
	Infos        map[string]interface{}
	CoursesTable map[string]interface{}
	GradesCount  map[string]interface{}
}

func NewUser(username, password string) *User {
	return &User{
		Username:     username,
		Password:     password,
		Infos:        make(map[string]interface{}),
		CoursesTable: make(map[string]interface{}),
		GradesCount:  make(map[string]interface{}),
	}
}
