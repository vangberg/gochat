package main

import (
        "./gochat";
)

func main() {
        s := new(gochat.Server);
        s.StartServer();
}
