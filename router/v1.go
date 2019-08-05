package router

import "goFrame/app/hello"

func init() {
	helloRouter := GetRouter()
	helloRouter.Handle("GET", "/hello", hello.Hello)
}
