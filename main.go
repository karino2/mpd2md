package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

/*
{
   "cell_type":"code","execution_count":1,"metadata":{"collapsed":false},
   "outputs":[
        {"data":
            {"image/png":"iVBORw0KGgoAAAANSUhEUgAAAfwAAAHgCAYAAABARt/zAAAAB
                        HNCSVQICAgIfAhkiAAAIABJREFU\neJzsnXd4FFX3x79b03svhFRICCVIV6TDG0QQKeJLlA4CFnoRRYUXVORK5CYII=\n",
             "text/plain":["<IPython.core.display.Image object>"]
            },
         "execution_count":1,"metadata":{},"output_type":"execute_result"
        }
    ],
    "source":["MeatPieImage()"]
    },

{"cell_type":"markdown","metadata":{"updated_at":1526482569783},
"source":["縦に話をする事で、それぞれの分野がどう違うのか、というのが、\nそんなに長い修行期間を経なくても"]}],
"metadata":{"kernelspec":{"display_name":"Python 2","language":"python","name":"python2"},
"lanbuage_info":{"codemirror_mode":{"name":"ipython","version":2},"file_extension":".py","mimetype":"text/x-python",
"name":"python","nbconvert_exporter":"python","pygments_lexer":"ipython2","version":"2.7.11"}},"nbformat":4,"nbformat_minor":0}


data class Cell(
        @SerializedName("cell_type")
        val _cellType: String,
        @SerializedName("source")
        val _source: JsonElement,
        val executionCount: Int? = null,
        val metadata: JsonElement? = null,
        // always 1 element.
        val outputs: List<Output>? = null) {
*/

/*
   data class Output(val name: String = "",
                     // (name, _text) or _data
                     val outputType: String? = null,
                     @SerializedName("text")
                     val _text: JsonElement? = null,
                     @SerializedName("data")
                     val _data: Map<String, JsonElement>? = null,

                     val executionCount: Int? = null
   ) {

*/

type data struct {
	ImagePng  string `json:"image/png"`
	ImageJpeg string `json:"image/jpeg"`
}

type output struct {
	Data data
}

type cell struct {
	CellType string   `json:"cell_type"`
	Source   []string `json:"source"`
	Outputs  []output
}

type note struct {
	//	Cells []map[string]interface{}
	Cells []cell
}

func parseTest() {
	// dat, _ := ioutil.ReadFile("p_space.ipynb")
	dat, _ := ioutil.ReadFile("intro.ipynb")
	// var m map[string]interface{}
	var m note
	json.Unmarshal(dat, &m)
	// fmt.Println(json.Unmarshal(dat, &m))
	// fmt.Printf("%d, %+v\n", len(m.Cells), m)
	for _, c := range m.Cells {
		fmt.Printf("%s\n", c.CellType)
		if c.CellType == "markdown" {
			fmt.Println(c.Source[0][0:20])
		} else {
			fmt.Println(c.Outputs[0].Data.ImagePng[0:10])
			// fmt.Printf("outs: %+v\n", c.Outputs)
		}

	}
}

func readAsNote(filename string) note {
	dat, _ := ioutil.ReadFile(filename)
	var n note
	json.Unmarshal(dat, &n)
	return n
}

func toImage(outfname string, base64str string) {
	writer, _ := os.Create(outfname)
	defer writer.Close()

	binary, _ := base64.StdEncoding.DecodeString(base64str)
	writer.Write(binary)

}

func writeHeader(file *os.File, cell string) {
	kvmap := map[string]string{
		"Title": "Default title",
	}
	for _, line := range strings.Split(cell, "\n") {
		kvarr := strings.SplitN(line, ":", 2)
		if len(kvarr) == 2 {
			kvmap[strings.Trim(kvarr[0], " ")] = strings.Trim(kvarr[1], " ")
		}
	}

	fmt.Fprint(file, `---
title: "`)
	fmt.Fprint(file, kvmap["Title"])
	fmt.Fprintln(file,
		`"
layout: page	
---
`)
}

func printMarkDown(file *os.File, contents string, ispandoc bool) {
	if ispandoc {
		patMathStates := regexp.MustCompile(`^\$\$[^\$]+\$\$$`)
		patMathInline := regexp.MustCompile(`\$\$[^\$]+\$\$`)
		patDolDol := regexp.MustCompile(`\$\$`)
		for _, line := range strings.Split(contents, "\n") {
			if patMathStates.MatchString(line) {
				// for only $$ $$ line, just as is for indenpent block.
				fmt.Fprintln(file, line)
			} else if patMathInline.MatchString(line) {
				// $$ $$ inside some sentence, replace $$ with $.
				fmt.Fprintln(file, patDolDol.ReplaceAllString(line, "$"))

			} else {
				fmt.Fprintln(file, line)
			}
		}
	} else {
		fmt.Fprint(file, contents)
	}
	fmt.Fprint(file, "\n\n")

}

func toMarkDown(filename string, ispandoc bool, ispost bool) {
	imgcount := 0
	dest := "work"

	basename := strings.TrimSuffix(filename, filepath.Ext(filename))
	// fmt.Println(basename)

	imgrel := fmt.Sprintf("images/%s", basename)
	_ = os.MkdirAll(fmt.Sprintf("%s/%s", dest, imgrel), 0777)

	n := readAsNote(filename)

	mdfname := fmt.Sprintf("%s/%s.md", dest, basename)
	mdf, _ := os.Create(mdfname)
	defer mdf.Close()

	head := n.Cells[0]
	writeHeader(mdf, head.Source[0])

	for _, c := range n.Cells[1:] {
		// fmt.Printf("%s\n", c.CellType)
		if c.CellType == "markdown" {
			printMarkDown(mdf, c.Source[0], ispandoc)
		} else {
			var imgname string
			imgdata := c.Outputs[0].Data
			if imgdata.ImagePng != "" {
				imgname = fmt.Sprintf("%s/%04d.png", imgrel, imgcount)
				imgpath := fmt.Sprintf("%s/%s", dest, imgname)
				toImage(imgpath, imgdata.ImagePng)
			} else {
				imgname = fmt.Sprintf("%s/%04d.jpg", imgrel, imgcount)
				imgpath := fmt.Sprintf("%s/%s", dest, imgname)
				toImage(imgpath, imgdata.ImageJpeg)
			}
			imgcount++

			if ispost {
				// ![imgs/2018-08-18-220928/0000.png]({{"/assets/images/2018-08-18-220928/0000.png" | absolute_url }})
				fmt.Fprintf(mdf, "![%s]({{\"/assets/%s\" | absolute_url}})\n\n", imgname, imgname)
			} else {
				fmt.Fprintf(mdf, "![%s](%s)\n\n", imgname, imgname)
			}
		}

	}

}

func main() {

	mdtype := flag.String("type", "jekyll", "pandoc, jekyll, post. For pandoc, math handling is a little different. For post, image link should be assets.")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Usage: mpd2md [-type=pandoc] target.ipynb")
		return
	}

	toMarkDown(flag.Args()[0], *mdtype == "pandoc", *mdtype == "post")
	// toMarkDown("intro.ipynb")
	// toMarkDown("jpgtest.ipynb")
	/*
		dest := "work"
		filename := "intro.ipynb"

		basename := strings.TrimSuffix(filename, filepath.Ext(filename))
		fmt.Println(basename)

		imgdest := fmt.Sprintf("%s/imgs/%s", dest, basename)
		_ = os.MkdirAll(imgdest, 0700)
	*/

	// fmt.Printf("%04d\n", 1)

	/*
		_ = os.Mkdir(dest, 0700)
	*/

}
