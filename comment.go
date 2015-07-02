package main

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
)

// Comment contains the data for a comment
// the xml schema matches what pyblosxom comments use
// the json schema is for public serving and hides email address and ip address
// it does provide an md5 hash of email address, so gravatar can be used
type Comment struct {
	XMLName   xml.Name  `xml:"item"`
	Parent    string    `xml:"parent"`
	Ts        time.Time `xml:"w3cdate"`
	Seq       uint      `xml:"-"`
	Message   string    `xml:"description"`
	Ipaddress string    `xml:"ipaddress" json:"-"`
	Author    string    `xml:"author"`
	Email     string    `xml:"email" json:"-"`
	Hash      string    `xml:"-"`
	Link      string    `xml:"link"`
}

func (c Comment) String() string {
	return fmt.Sprintf("<Comment at %s by %s>", c.Ts, c.Author)

}

type ByTsAsc []Comment

func (a ByTsAsc) Len() int           { return len(a) }
func (a ByTsAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTsAsc) Less(i, j int) bool { return a[i].Ts.Before(a[j].Ts) }

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
	c.Seq = uint(seq)
	c.Hash = fmt.Sprintf("%x", md5.Sum([]byte(c.Email)))
	return &c, nil
}

func (c Comment) tmpName() string {
	return fmt.Sprintf("%s-%d.%d.tmp", c.Parent, c.Ts.Unix(), c.Seq)
}

func (c Comment) fileName() string {
	return fmt.Sprintf("%s-%d.%d.cmt", c.Parent, c.Ts.Unix(), c.Seq)
}

func (c Comment) Save(path string) error {
	data, err := xml.Marshal(c)
	if err != nil {
		fmt.Printf("comment %s can't be marshalled: %s\n", c, err)
		return err
	}
	if c.Seq == 0 {
		rand.Seed(time.Now().UnixNano())
		c.Seq = uint(rand.Uint32())
	}
	fullName := path + "/" + c.tmpName()
	f, err := os.Create(fullName)
	if err != nil {
		fmt.Printf("comment %s can't open file %s: %s\n", c, fullName, err)
		return err
	}
	defer f.Close()
	fmt.Println("saving", c)
	_, err = f.Write(data)
	if err != nil {
		fmt.Printf("comment %s can't be written: %s\n", c, err)
		return err
	}

	err = os.Rename(fullName, path+"/"+c.fileName())
	if err != nil {
		fmt.Printf("comment %s can't be renamed to final file: %s\n", c, err)
	}
	return err
}

func FindComments(path, slug string) ([]Comment, error) {
	re := regexp.MustCompile("^" + slug + "-([0-9]{10})\\.([0-9]+)\\.cmt$")
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
