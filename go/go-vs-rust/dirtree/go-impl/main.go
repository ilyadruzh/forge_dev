package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func printDir(path string) {
	// Печатать только директории

	var buffer bytes.Buffer

	buffer.WriteString("  ")

	items, err := ioutil.ReadDir(path)
	if err == nil {
		for i := 0; i < len(items); i++ {
			if items[i].IsDir() {
				fmt.Printf("├───%s\n", items[i].Name())
				recursiveFunc(buffer, path, items[i].Name(), false)
			}
		}
	}
}

func printFile(path string) {
	// Печатать директории и файлы в них

	var buffer bytes.Buffer
	buffer.WriteString("  ")

	items, err := ioutil.ReadDir(path)

	if err == nil {
		for i := 0; i < len(items); i++ {
			if !items[i].IsDir() {
				if len(items)-1 == i {
					fmt.Printf("└───%s (%d)\n", items[i].Name(), items[i].Size())
				} else {
					fmt.Printf("├───%s (%d)\n", items[i].Name(), items[i].Size())
				}
			} else {
				fmt.Printf("├───%s\n", items[i].Name())
				recursiveFunc(buffer, path, items[i].Name(), true)
			}
		}
	}
}

func recursiveFunc(format bytes.Buffer, father string, son string, printFiles bool) {

	// fmt.Printf("father = %s,\n son = %s\n ", father, son)

	// графическое дерево нарисовать
	// добавить аккaмулятор
	buffer := format

	fullpath := father + string(os.PathSeparator) + son
	items, err := ioutil.ReadDir(fullpath)

	if err == nil && printFiles == true {
		for i := 0; i < len(items); i++ {
			if items[i].IsDir() {
				if len(items)-1 != i {
					fmt.Printf("|%s├───%s\n", buffer.String(), items[i].Name())
					buffer.WriteString("  ")
					recursiveFunc(buffer, fullpath, items[i].Name(), true)
				} else {
					fmt.Printf("|%s└───%s \n", buffer.String(), items[i].Name())
					buffer.WriteString("  ")
					recursiveFunc(buffer, fullpath, items[i].Name(), true)
				}
			} else {
				if len(items)-1 == i {
					fmt.Printf("|%s└───%s (%d)\n", buffer.String(), items[i].Name(), items[i].Size())
				} else {
					fmt.Printf("|%s├───%s (%d)\n", buffer.String(), items[i].Name(), items[i].Size())
				}
			}
		}
	}
	if err == nil && printFiles == false {
		for i := 0; i < len(items); i++ {
			// если последняя папка в папке
			if items[i].IsDir() && len(items)-1 == i {
				fmt.Printf("|%s└───%s \n", buffer.String(), items[i].Name())
				buffer.WriteString("  ")
				recursiveFunc(buffer, fullpath, items[i].Name(), false)
			} else {
				// если не последняя папка
				if items[i].IsDir() && len(items)-1 != i {
					fmt.Printf("|%s├───%s\n", buffer.String(), items[i].Name())
					// buffer.WriteString("  ")
					recursiveFunc(buffer, fullpath, items[i].Name(), false)
				}
			}

		}
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	switch printFiles {
	case true:
		printFile(path)
	case false:

		items, err := ioutil.ReadDir(path)
		if err == nil {
			for i := 0; i < len(items); i++ {
				if items[i].IsDir() {
					fmt.Println(items[i].Name())
					printDir(items[i].Name())
				}
			}
		}

	default:
		fmt.Println("Something wrong")
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
