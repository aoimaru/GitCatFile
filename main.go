package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"reflect"
	// "log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Object struct {
	Type string
	Size int
	Data []byte
}

type Blob struct {
	Size int
	Data []byte
}

type Column struct {
	Type string
	Name string
	Hash string
}

type Tree struct {
	Size    int
	Columns []Column
}

func (t *Tree) Format() {
	for _, v := range t.Columns {
		if v.Type == "tree" {
			fmt.Printf("040000 %s %s    %s\n", v.Type, v.Hash, v.Name)
		} else {
			fmt.Printf("100644 %s %s    %s\n", v.Type, v.Hash, v.Name)
		}
	}
}

type Parent struct {
	Hash string
}

type Sign struct {
	Name      string
	Email     string
	TimeStamp time.Time
}

type Commit struct {
	Tree      string
	Parents   []Parent
	Author    Sign
	Committer Sign
	Message   string
}

func (b *Blob) out_content() {
	fmt.Println(string(b.Data))
}

func (b *Blob) out_header() {
	fmt.Printf("type: %s size:%d \n", "blob", b.Size)
}

func extract(zr io.Reader) (io.Reader, error) {
	return zlib.NewReader(zr)
}

func FmtObject(S *[]byte) [][]byte {
	objs := make([][]byte, 0)
	obj := make([]byte, 0)
	for _, s := range *S {
		if s == 0 {
			if len(obj) <= 1 {
				continue
			}
			objs = append(objs, obj)
			obj = make([]byte, 0)
		}
		obj = append(obj, s)
	}
	return objs
}

func reverse(array *[]byte) []byte {
	for k, v := range *array {
		if v == 0 {
			continue
		}
		fmt.Println(k, v)
	}
	return *array
}

func HashSample() []byte {
	f, err := os.Open("/home/aoimaru/document/go_project/HashObject/sample")
	if err != nil {
		return []byte("hello")
	}
	defer f.Close()
	dc := make([]byte, 1024)
	(*f).Read(dc)
	buf := bytes.NewBuffer(dc)
	r, err := extract(buf)
	if err != nil {
		return []byte("hello")
	}
	c := make([]byte, 1024)
	r.Read(c)
	return c
}

func GetObject(hash string) []byte {
	t_rep := "/mnt/c/Users/81701/Desktop/AtCoder/.git/objects/"
	first_hash, second_hash := hash[:2], hash[2:]
	f, err := os.Open(t_rep + "/" + first_hash + "/" + second_hash)
	if err != nil {
		return []byte(hash)
	}
	defer f.Close()
	dc := make([]byte, 1024)
	(*f).Read(dc)
	buf := bytes.NewBuffer(dc)
	r, err := extract(buf)
	if err != nil {
		return []byte(hash)
	}
	c := make([]byte, 1024)
	r.Read(c)
	return c
}

var (
	emailRegexpString     = "([a-zA-Z0-9_.+-]+@([a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\\.)+[a-zA-Z]{2,})"
	timestampRegexpString = "([1-9][0-9]* \\+[0-9]{4})"
	sha1Regexp            = regexp.MustCompile("[0-9a-f]{20}")
	signRegexp            = regexp.MustCompile("^[^<]* <" + emailRegexpString + "> " + timestampRegexpString + "$")
)

func main() {
	// c := GetObject("01fa50e0a408cb77e94ffdc161643c1ac65794bb")
	// c := GetObject("0821bf3154d58047ae43053e6660a7906cfa0855")
	// c := GetObject("b5b7fe8f6a4bbf175e2a32b8624fd234aeb02a69")
	c := GetObject("57a8901f011f9c65f1a33bd6a55990acc42935c6")
	// c := GetObject("98cdbc1ec60aa9b0f1142e0daab34ffc297955bd")
	// c := GetObject("b8bd4a446eceb5655176cf3c4168513bbd77fc46") <-Gitのガベージコレクションで消されてる 多分
	// c := GetObject("5519da6cca07470631a9e5dc9286fba3fbffb7d8")
	// c := GetObject("c0c67ed0a4de0b63eaedb344c5faa42b962c6667") <-バグる
	// c := HashSample()


	// fmt.Println(string(c))
	// fmt.Println(" ")
	// fmt.Println(" ")
	// fmt.Println(" ")
	// fmt.Println(" ")
	// fmt.Println(" ")

	objs := FmtObject(&c)
	if len(objs) < 2 {
		return
	}

	if strings.HasPrefix(string(c), "blob ") {
		Header := string(objs[0])
		Contents := objs[1:]

		// fmt.Println(Header, objs[0][5:], string(objs[0][5:]))
		fmt.Println(Header)
		for _, Content := range Contents {
			fmt.Println(string(Content))
		}
	} else if strings.HasPrefix(string(c), "tree ") {
		Header := string(objs[0])
		Contents := objs[1:]

		fmt.Println(Header)
		nContents := make([][]byte, 0)
		for _, Content := range Contents {
			if len(Content) >= 20 {
				hash := hex.EncodeToString(Content[1:21])
				meta := Content[21:]
				nContents = append(nContents, []byte(hash))
				nContents = append(nContents, meta)
			} else {
				meta := Content[1:]
				nContents = append(nContents, meta)
			}
		}

		
		columns := make([]Column, 0)
		for n, nContent := range nContents {
			if len(string(nContent)) <= 0 {
				continue
			}
			if n%2 == 0 {
				name := strings.Replace(string(nContent), "40000 ", "", -1)
				if strings.HasPrefix(string(nContent), "40000") {
					column := Column{
						Type: "tree",
						Name: name,
						Hash: string(nContents[n+1]),
					}
					columns = append(columns, column)
				} else {
					column := Column{
						Type: "blob",
						Name: name,
						Hash: string(nContents[n+1]),
					}
					columns = append(columns, column)
				}
			} else {
				continue
			}
		}
		tree_size := strings.Replace(Header, "tree ", "", -1)
		size, err := strconv.Atoi(tree_size)
		if err != nil {
			return
		}
		tree := Tree{
			Size:    size,
			Columns: columns,
		}
		tree.Format()

	} else if strings.HasPrefix(string(c), "commit ") {
		Header := string(objs[0])
		Contents := objs[1:]
		fmt.Println(reflect.TypeOf(Contents))
		fmt.Println(Header)

		bContents := make([]byte, 0)
		for _, Content := range Contents {
			bContents = append(bContents, Content...)
		}

		cReader := strings.NewReader(string(bContents[1:]))
		scanner := bufio.NewScanner(cReader)
		// fmt.Println(scanner)

		var commit Commit
		var parents []Parent

		for scanner.Scan() {
			text := scanner.Text()
			cols := strings.SplitN(text, " ", 2)
			if len(cols) != 2 {
				break
			}
			lineType := cols[0]
			lineMeta := cols[1]
			fmt.Println(" ")
			fmt.Println(cols)

			switch lineType {
			case "tree":
				Hash := strings.Replace(lineMeta, "tree ", "", -1)
				Hash = strings.ReplaceAll(Hash, " ", "")
				fmt.Println(Hash)
				commit.Tree = Hash
			case "parent":
				var parent Parent
				Hash := strings.Replace(lineMeta, "parent ", "", -1)
				Hash = strings.ReplaceAll(Hash, " ", "")
				parent.Hash = Hash
				parents = append(parents, parent)
				commit.Parents = parents
			case "author":
				if ok := signRegexp.MatchString(lineMeta); !ok {
					continue
				}
				sign1 := strings.SplitN(lineMeta, " <", 2)
				name := sign1[0]
				sign2 := strings.SplitN(sign1[1], "> ", 2)
				email := sign2[0]
				sign3 := strings.SplitN(sign2[1], " ", 2)
				unixTime, err := strconv.ParseInt(sign3[0], 10, 64)
				fmt.Println(unixTime, reflect.TypeOf(unixTime))
				fmt.Println("type:sign3[1]", reflect.TypeOf(sign3[1]), sign3[1])
				if err != nil {
					continue
				}
				var offsetHour, offsetMinute int
				if _, err := fmt.Sscanf(sign3[1], "+%02d%02d", &offsetHour, &offsetMinute); err != nil {
					continue
				}
				location := time.FixedZone(" ", 3600*offsetHour+60*offsetMinute)
				timestamp := time.Unix(unixTime, 0).In(location)
				time.Now().String()
				fmt.Println("timeStamp:", timestamp, reflect.TypeOf(timestamp))
				fmt.Println("name:", name)
				fmt.Println("email:", email)

				sign := Sign{
					Name:      name,
					Email:     email,
					TimeStamp: timestamp,
				}
				commit.Author = sign

			case "committer":
				if ok := signRegexp.MatchString(lineMeta); !ok {
					continue
				}
				sign1 := strings.SplitN(lineMeta, " <", 2)
				name := sign1[0]
				sign2 := strings.SplitN(sign1[1], "> ", 2)
				email := sign2[0]
				sign3 := strings.SplitN(sign2[1], " ", 2)
				unixTime, err := strconv.ParseInt(sign3[0], 10, 64)
				if err != nil {
					continue
				}
				var offsetHour, offsetMinute int
				if _, err := fmt.Sscanf(sign3[1], "+%02d%02d", &offsetHour, &offsetMinute); err != nil {
					continue
				}
				location := time.FixedZone(" ", 3600*offsetHour+60*offsetMinute)
				timestamp := time.Unix(unixTime, 0).In(location)
				time.Now().String()
				fmt.Println("timeStamp:", timestamp)
				fmt.Println("name:", name)
				fmt.Println("email:", email)
				sign := Sign{
					Name:      name,
					Email:     email,
					TimeStamp: timestamp,
				}
				commit.Committer = sign
			}

		}
		messages := make([]string, 0)
		for scanner.Scan() {
			messages = append(messages, scanner.Text())
		}
		message := strings.Join(messages, "\n")
		fmt.Println("message: ", message)
		commit.Message = message

		fmt.Printf("%+v\n", commit)

	}

}
