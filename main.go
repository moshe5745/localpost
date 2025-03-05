package main

import (
    "flag"
    "io/ioutil"
    "strings"

    "github.com/posener/complete"
)

var hello = flag.String("hello", "", "Say hello")

func main() {
    flag.Parse()
    cmd := complete.Command{
        Flags: complete.Flags{
            "-hello": complete.PredictFunc(func(args complete.Args) []string {
                content, _ := ioutil.ReadFile("completions.txt")
                return strings.Split(string(content), "\n")
            }),
        },
    }
    complete.Complete("mycli", cmd)
    if *hello != "" {
        println("Hello,", *hello)
    }
}