package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
	"github.com/awgh/gobfuscate"
)

func main() {
	parser := argparse.NewParser("gobfuscate", "Obfuscates and builds a Go package.")
	encKey := parser.String("e", "enckey", &argparse.Options{Required: false, Help: "rename encryption key"})
	outputGopath := parser.Flag("a", "outdir", &argparse.Options{Required: false, Help: "output a full GOPATH"})
	keepTests := parser.Flag("k", "keeptests", &argparse.Options{Required: false, Help: "keep _test.go files"})
	winHide := parser.Flag("w", "winhide", &argparse.Options{Required: false, Help: "hide windows GUI"})
	noStaticLink := parser.Flag("n", "nostatic", &argparse.Options{Required: false, Help: "do not statically link"})
	preservePackageName := parser.Flag("p", "noencrypt", &argparse.Options{Required: false,
		Help: "no encrypted package name for go build command (works when main package has CGO code)"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "verbose mode"})

	if err := parser.Parse(os.Args); err != nil {
		log.Println(parser.Usage(err))
		return
	}

	if len(flag.Args()) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: gobfuscate [flags] pkg_name out_path")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pkgName := flag.Args()[0]
	outPath := flag.Args()[1]

	if *encKey == "" {
		buf := make([]byte, 32)
		rand.Read(buf)
		*encKey = string(buf)
	}

	if !gobfuscate.Obfuscate(gobfuscate.Config{PkgName: pkgName, OutPath: outPath,
		EncKey: *encKey, OutputGopath: *outputGopath, KeepTests: *keepTests, WinHide: *winHide,
		NoStaticLink: *noStaticLink, PreservePackageName: *preservePackageName, Verbose: *verbose}) {
		os.Exit(1)
	}
}
