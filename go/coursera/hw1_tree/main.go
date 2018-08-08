package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func recursiveFunc(father string, son string) {
	fullpath := father + string(os.PathSeparator) + son
	items, err := ioutil.ReadDir(fullpath)

	if err == nil {
		for i := 0; i < len(items); i++ {
			if items[i].IsDir() {
				fmt.Println("DIR = ", items[i].Name())
				recursiveFunc(fullpath, items[i].Name())
			} else {
				fmt.Println("FILE = ", items[i].Name())
			}
		}
	} else {
		fmt.Println("\t ERROR = ", err)
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	if printFiles {

		items, err := ioutil.ReadDir(path)

		if err == nil {
			for i := 0; i < len(items); i++ {
				if !items[i].IsDir() {
					fmt.Println(items[i].Name())
				} else {
					fmt.Println(path + string(os.PathSeparator) + items[i].Name())
					recursiveFunc(path, items[i].Name())
				}
			}
		} else {
			fmt.Println("Error 1 = ", err)
		}
	}
	return nil
}

func main() {

	out := os.Stdout

	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	err := dirTree(out, path, printFiles)

	if err != nil {
		panic(err.Error())
	}
}
