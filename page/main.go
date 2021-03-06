/*
  Copyright (c) 2012-2013 José Carlos Nieto, http://xiam.menteslibres.org/

  Permission is hereby granted, free of charge, to any person obtaining
  a copy of this software and associated documentation files (the
  "Software"), to deal in the Software without restriction, including
  without limitation the rights to use, copy, modify, merge, publish,
  distribute, sublicense, and/or sell copies of the Software, and to
  permit persons to whom the Software is furnished to do so, subject to
  the following conditions:

  The above copyright notice and this permission notice shall be
  included in all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
  MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
  LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
  OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
  WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package page

import (
	"html/template"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"fmt"
)

// This structure holds information on the current document served by Luminos.
type Page struct {

	// Page title, guessed from the current document. (Looks for the first H1, H2, ..., H6 tag)
	Title string

	// The HTML of the current document.
	Content template.HTML

	// The HTML of the _header.md or _header.html file on the current document's directory.
	ContentHeader template.HTML

	// The HTML of the _footer.md or _footer.html file on the current document's directory.
	ContentFooter template.HTML

	// An array of maps that contains names and links of all the items on the document root.
	// Names begginning with "." or "_" are ignored in this list.
	Menu []map[string]interface{}

	// An array of maps that contains names and links of all the items on the current document's directory.
	// Names begginning with "." or "_" are ignored in this list.
	SideMenu []map[string]interface{}

	// An array of maps that contains names and links of the current document's path.
	BreadCrumb []map[string]interface{}

	// A map that contains name and link of the current page.
	CurrentPage map[string]interface{}

	// Absolute path of the current document.
	FilePath string

	// Absolute parent directory of the current document.
	FileDir string

	// Relative path of the current document.
	BasePath string

	// Relative parent directory of the current document.
	BaseDir string

	// True if the current document is / (home).
	IsHome bool
}

var extensions = []string{".html", ".md", ""}

// Just a list of files that can be sorted.
type fileList []os.FileInfo

func (f fileList) Len() int {
	return len(f)
}

func (f fileList) Less(i, j int) bool {
	return f[i].Name() < f[j].Name()
}

func (f fileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type byName struct{ fileList }

const (
	PS = string(os.PathSeparator)
)

// Strips out known extensions for a given file name.
func removeKnownExtension(s string) string {
	fileExt := path.Ext(s)

	for _, ext := range extensions {
		if ext != "" {
			if fileExt == ext {
				return s[:len(s)-len(ext)]
			}
		}
	}

	return s
}

// Returns files in a directory passed through a filter.
func filterList(directory string, filter func(os.FileInfo) bool) fileList {
	var list fileList

	fp, err := os.Open(directory)
	defer fp.Close()

	if err != nil {
		panic(err)
	}

	ls, err := fp.Readdir(-1)

	if err != nil {
		panic(err)
	}

	for _, file := range ls {
		fmt.Printf("Considering >>[%s]\n", file.Name())

		if filter(file) == true {
			list = append(list, file)
		}
	}

	sort.Sort(byName{list})

	return list
}

// A filter for filterList. Returns all directories except those that begin with "." or "_".
func directoryFilter(f os.FileInfo) bool {
	if strings.HasPrefix(f.Name(), ".") == false && strings.HasPrefix(f.Name(), "_") == false {
		return f.IsDir()
	}
	return false
}

// A filter for filterList. Returns all files except for those that 
// begin with "." or "_", or end with "~" (applies to directory names, too,
// unlike the original luminos)

func mdFilter(f os.FileInfo) bool {
	n := f.Name()
	if strings.HasPrefix(n, ".") {
		return false
	}
	if strings.HasPrefix(n, "_") {
		return false
	}
	if !strings.HasSuffix(n, ".md") {
		return false
	}
	return true
}

// Returns a stylized human title, given a file name.
func createTitle(s string) string {
	s = removeKnownExtension(s)

	re, _ := regexp.Compile("[-_]")
	s = re.ReplaceAllString(s, " ")

	return strings.Title(s[:1]) + s[1:]
}

// Returns a link.
func (p *Page) CreateLink(file os.FileInfo, prefix string) map[string]interface{} {
	item := map[string]interface{}{}

	if file.IsDir() == true {
		item["link"] = prefix + file.Name() + "/"
	} else {
		item["link"] = prefix + removeKnownExtension(file.Name())
	}

	item["text"] = createTitle(file.Name())

	return item
}

func (p *Page) CreateMenu() {
	var item map[string]interface{}
	p.Menu = []map[string]interface{}{}

	fmt.Printf("Creating menu...\n")
	files := filterList(p.FileDir, directoryFilter)
	fmt.Printf("done building files (%d entries)\n", len(files))

	for _, file := range files {
		item = p.CreateLink(file, p.BasePath)
		fmt.Printf("Considering [%s]\n", p.FileDir+PS+file.Name())
		children := filterList(p.FileDir+PS+file.Name(), 
			directoryFilter)
		fmt.Printf("   found %d children\n", len(children))
		if len(children) > 0 {
			item["children"] = []map[string]interface{}{}
			for _, child := range children {
				fmt.Printf("   matched [%s]\n", child)
				childItem := p.CreateLink(child, p.BasePath+file.Name()+"/")
				item["children"] = append(item["children"].([]map[string]interface{}), childItem)
			}
		}
		p.Menu = append(p.Menu, item)
	}
}

// Populates Page.BreadCrumb with links.
func (p *Page) CreateBreadCrumb() {

	p.BreadCrumb = []map[string]interface{}{
		map[string]interface{}{
			"link": "/",
			"text": "Home",
		},
	}

	chunks := strings.Split(strings.Trim(p.BasePath, "/"), "/")

	prefix := ""

	for _, chunk := range chunks {
		if chunk != "" {
			item := map[string]interface{}{}
			item["link"] = prefix + "/" + chunk + "/"
			item["text"] = createTitle(chunk)
			prefix = prefix + PS + chunk
			p.BreadCrumb = append(p.BreadCrumb, item)
			p.CurrentPage = item
		}
	}

}

// Populates Page.SideMenu with files on the current document's directory.
func (p *Page) CreateSideMenu() {
	var item map[string]interface{}
	p.SideMenu = []map[string]interface{}{}

	fmt.Printf("Creating side menu\n");
	files := filterList(p.FileDir, mdFilter)
	fmt.Printf("   done with %d entries\n", len(files));

	for _, file := range files {
		item = p.CreateLink(file, p.BasePath)
		if strings.ToLower(item["text"].(string)) != "index" {
			p.SideMenu = append(p.SideMenu, item)
		}
	}
}
