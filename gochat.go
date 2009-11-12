package main

import (
        "./server";
)

func main() {
        s := new(server.Server);
        s.Listen();
}
