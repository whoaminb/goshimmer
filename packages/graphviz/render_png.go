package graphviz

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/emicklei/dot"
)

func RenderPNG(graph *dot.Graph, fileName string) (err error) {
	if err = ioutil.WriteFile("_tmp.dot", []byte(graph.String()), os.ModePerm); err != nil {
		return
	}

	if _, err = exec.Command("dot", "-Tpng", "_tmp.dot", "-o", fileName).Output(); err != nil {
		return
	}

	if err = os.Remove("_tmp.dot"); err != nil {
		return
	}

	return
}
