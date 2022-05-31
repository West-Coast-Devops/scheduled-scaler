package main

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()

	executable, err := os.Executable()
	if err != nil {
		glog.Fatalf("can't get the executable: %v", err)
	}
	glog.V(1).Info("Running ", executable)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "tools/tools.go", nil, parser.ImportsOnly)
	if err != nil {
		glog.Fatal(err)
	}
	for _, imp := range f.Imports {
		var importPath string
		if err = json.Unmarshal([]byte(imp.Path.Value), &importPath); err != nil {
			glog.Fatalf("error unmarshalling %s: %v", imp.Path.Value, err)
		}
		glog.V(2).Info("import ", importPath)
		cmd := exec.Command("go", "install", importPath)
		glog.Info(cmd.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			glog.Fatal(err)
		}
	}
}
