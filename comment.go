package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Comment struct {
	Parent    string    `xml:"parent"`
	Ts        time.Time `xml:"pubDate"`
	Seq       uint8     `xml:"-"`
	Message   string    `xml:"description"`
	Ipaddress string    `xml:"ipaddress"`
	Author    string    `xml:"author"`
	Email     string    `xml:"email" json:"-"`
	Link      string    `xml:"link"`
}

func NewComment(re *regexp.Regexp, path, fileName string) (*Comment, error) {
	res := re.FindStringSubmatch(fileName)
	ts, _ := strconv.Atoi(res[1])
	seq, _ := strconv.Atoi(res[2])
	content, err := ioutil.ReadFile(path + "/" + fileName)
	if err != nil {
		return nil, err
	}
	c := Comment{}
	err = xml.Unmarshal([]byte(content), &c)
	if err != nil {
		return nil, err
	}
	c.Ts = time.Unix(int64(ts), 0)
	c.Seq = uint8(seq)
	return &c, nil
}

func (c Comment) tmpName() string {
	return fmt.Sprintf("%s-%d.%d.tmp", c.Parent, c.Ts.Unix(), rand.Int())
}

func (c Comment) fileName() string {
	return fmt.Sprintf("%s-%d.%d.cmt", c.Parent, c.Ts.Unix(), c.Seq)
}

func (c Comment) Save(path string) error {
	data, err := xml.Marshal(c)
	if err != nil {
		return err
	}
	name := c.tmpName()
	f, err := os.Create(path + "/" + name)
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	err = fmt.Errorf("initial-value")
	for err != nil {
		err = os.Rename(path+"/"+name, path+"/"+c.fileName())
		c.Seq += 1
	}
	return nil
}

func FindComments(path, slug string) ([]Comment, error) {
	re := regexp.MustCompile("^" + slug + "-([0-9]{10})\\.([0-9]{2})\\.cmt$")
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	fi, err := dir.Stat()
	if err != nil {
		return nil, err
	}
	comments := make([]Comment, 0)
	if fi.IsDir() {
		fis, err := dir.Readdir(-1) // -1 means return all the FileInfos
		if err != nil {
			return nil, err
		}
		for _, fileinfo := range fis {
			if !fileinfo.IsDir() && re.MatchString(fileinfo.Name()) {
				fmt.Println(fileinfo.Name())
				c, err := NewComment(re, path, fileinfo.Name())
				if err != nil {
					return nil, err
				}
				comments = append(comments, *c)
			}
		}
	}
	return comments, nil
}
