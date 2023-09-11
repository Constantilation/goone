package main

import (
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type DirNames struct {
	name  string
	index int
}

type DirParams struct {
	dir         os.DirEntry
	isLastDir   bool
	isLastFile  bool
	layerNumber int
}

type Node struct {
	data DirParams
	next *Node
}

type List struct {
	head *Node
}

func (l *List) add(value DirParams) {
	newNode := &Node{data: value}

	if l.head == nil {
		l.head = newNode
		return
	}

	curr := l.head
	for curr.next != nil {
		curr = curr.next
	}

	curr.next = newNode
}

func (l *List) remove(value DirParams) {
	if l.head == nil {
		return
	}

	if l.head.data == value {
		l.head = l.head.next
		return
	}

	curr := l.head
	for curr.next != nil && curr.next.data != value {
		curr = curr.next
	}

	if curr.next != nil {
		curr.next = curr.next.next
	}
}

func ReadDir(path string, printFiles bool) (dirEntry []os.DirEntry, err error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	dirs, err := file.ReadDir(-1)
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })

	if !printFiles {
		for index, dir := range dirs {
			if !dir.IsDir() {
				if index != len(dirs) {
					dirs = append(dirs[:index], dirs[index+1:]...)
				} else {
					dirs = dirs[:index]
				}
			}
		}
	}

	return dirs, err
}

func printList(l *List, out io.Writer) (err error) {
	tab := "\t"
	lastFile := "└───"
	file := "├───"
	upperSeparator := "│"

	curr := l.head

	var symbolsArray []string
	isLastDir := false
	curLayer := 0

	for curr != nil {
		currName := curr.data.dir.Name()

		if !curr.data.dir.IsDir() {
			currInfo, _ := curr.data.dir.Info()
			if currInfo.Size() == 0 {
				currName = currName + " (empty)"
			} else {
				currName = currName + " (" + strconv.FormatInt(currInfo.Size(), 10) + "b)"
			}
		}

		var fileSeparator string
		if curr.data.isLastDir || curr.data.isLastFile {
			fileSeparator = lastFile
		} else {
			fileSeparator = file
		}

		if curr.data.layerNumber > curLayer {
			if !isLastDir {
				symbolsArray = append(symbolsArray, upperSeparator+tab)
			} else {
				symbolsArray = append(symbolsArray, tab)
			}

			curLayer = curr.data.layerNumber
		} else if curr.data.layerNumber < curLayer {
			symbolsArray = symbolsArray[:curr.data.layerNumber]
			curLayer = curr.data.layerNumber
		}

		isLastDir = curr.data.isLastDir

		var resultSeparator string

		for _, symbol := range symbolsArray {
			resultSeparator += symbol
		}

		_, err := out.Write([]byte(resultSeparator + fileSeparator + currName + "\n"))
		if err != nil {
			return err
		}

		curr = curr.next
	}

	return err
}

func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	lastDirArray := make([]DirNames, 0)
	var dirList List

	dirs, err := ReadDir(path, printFiles)

	currentPath := path

	for i := 0; i <= len(dirs); i++ {
		if i < len(dirs) {
		AnotherDir:
			switch len(dirs) > 0 && dirs[i].IsDir() {
			case true:
				lastDirArray = append(lastDirArray, DirNames{name: currentPath, index: i})
				currentPath = currentPath + `/` + dirs[i].Name()

				dirsAmount := 0
				for _, dir := range dirs {
					if dir.IsDir() {
						dirsAmount++
					}
				}

				dirParams := DirParams{dir: dirs[i], isLastDir: i == len(dirs)-1, layerNumber: strings.Count(currentPath, "/") - 1}
				dirList.add(dirParams)
				dirs, err = ReadDir(currentPath, printFiles)
				i = 0
				goto AnotherDir
			default:
				if printFiles {
					dirParams := DirParams{dir: dirs[i], isLastFile: i == len(dirs)-1, layerNumber: strings.Count(currentPath, "/")}
					dirList.add(dirParams)
				}
			}
		}

		if i == len(dirs)-1 || i == len(dirs) {
			if len(lastDirArray) != 0 {
				dirs, err = ReadDir(lastDirArray[len(lastDirArray)-1].name, printFiles)
				i = lastDirArray[len(lastDirArray)-1].index
				currentPath = lastDirArray[len(lastDirArray)-1].name
				lastDirArray = lastDirArray[:len(lastDirArray)-1]
			}
		}
	}

	return printList(&dirList, out)
}

func main() {
	// Первый мр на Ревью. Описание лежит в 1/99_hw/tree/hw1.md
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
>>>>>>> 2d01912 (Первый мр)
