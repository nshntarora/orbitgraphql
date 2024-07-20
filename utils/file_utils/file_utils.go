package file_utils

import (
	"fmt"
	"os"
)

type FileActor interface {
	Write(content string)
	Print()
	Open()
	Read() string
	Close()
	WriteJSON(content []interface{})
}

type File struct {
	ref *os.File
}

func NewFile(path string) *File {
	foc, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	return &File{ref: foc}
}

func (f *File) Write(content string) {
	f.ref.Write([]byte(content))
}

func (f *File) Read() string {
	output := make([]byte, 0)
	_, err := f.ref.Read(output)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (f *File) Print() {
	fmt.Println(f.Read())
}

func (f *File) Open() {
	if err := f.ref.Close(); err != nil {
		panic(err)
	}
}

func (f *File) Close() {
	if err := f.ref.Close(); err != nil {
		panic(err)
	}
}

func (f *File) WriteJSON(content []string) {
	f.Write("[")
	for idx, r := range content {
		// write a chunk
		if _, err := f.ref.Write([]byte(r)); err != nil {
			panic(err)
		}
		if idx != len(content)-1 {
			f.ref.Write([]byte(","))
		}
	}
	f.ref.Write([]byte("]"))
}
