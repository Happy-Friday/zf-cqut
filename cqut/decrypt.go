package cqut

import "github.com/axgle/mahonia"

func DecodeGbk(s string) string{
	dec := mahonia.NewDecoder("gbk")
	return dec.ConvertString(s)
}
