package main

import (
	"./gochat";
	"fmt";
)

func main() {
	fmt.Println(" * Starting server…");
	s := new(gochat.Server);
	s.StartServer();
}
