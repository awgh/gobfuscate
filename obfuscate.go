package gobfuscate

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Config - Command line arguments.for gobfuscate
type Config struct {
	PkgName             string
	OutPath             string
	EncKey              string
	OutputGopath        bool
	KeepTests           bool
	WinHide             bool
	NoStaticLink        bool
	PreservePackageName bool
	Verbose             bool
}

// Obfuscate - main entry point
func Obfuscate(c Config) bool {
	var newGopath string
	if c.OutputGopath {
		newGopath = c.OutPath
		if err := os.Mkdir(newGopath, 0755); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create destination:", err)
			return false
		}
	} else {
		var err error
		newGopath, err = ioutil.TempDir("", "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create temp dir:", err)
			return false
		}
		defer os.RemoveAll(newGopath)
	}

	log.Println("Copying GOPATH...")

	if err := CopyGopath(c.PkgName, newGopath, c.KeepTests); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to copy into a new GOPATH:", err)
		return false
	}

	enc := &Encrypter{Key: c.EncKey}
	log.Println("Obfuscating package names...")
	if err := ObfuscatePackageNames(newGopath, enc); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate package names:", err)
		return false
	}
	log.Println("Obfuscating strings...")
	if err := ObfuscateStrings(newGopath); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate strings:", err)
		return false
	}
	log.Println("Obfuscating symbols...")
	if err := ObfuscateSymbols(newGopath, enc); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to obfuscate symbols:", err)
		return false
	}

	if c.OutputGopath {
		return true
	}

	ctx := build.Default

	newPkg := c.PkgName
	if !c.PreservePackageName {
		newPkg = encryptComponents(c.PkgName, enc)
	}

	ldflags := `-ldflags=-s -w`
	if c.WinHide {
		ldflags += " -H=windowsgui"
	}
	if !c.NoStaticLink {
		ldflags += ` -extldflags "-static"`
	}

	goCache := newGopath + "/cache"
	os.Mkdir(goCache, 0755)

	arguments := []string{"build", ldflags, "-o", c.OutPath, newPkg}
	environment := []string{
		"GOROOT=" + ctx.GOROOT,
		"GOARCH=" + ctx.GOARCH,
		"GOOS=" + ctx.GOOS,
		"GOPATH=" + newGopath,
		"PATH=" + os.Getenv("PATH"),
		"GOCACHE=" + goCache,
	}

	cmd := exec.Command("go", arguments...)
	cmd.Env = environment
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if c.Verbose {
		fmt.Println()
		fmt.Println("[Verbose] Temporary path:", newGopath)
		fmt.Println("[Verbose] Go build command: go", strings.Join(arguments, " "))
		fmt.Println("[Verbose] Environment variables:")
		for _, envLine := range environment {
			fmt.Println(envLine)
		}
		fmt.Println()
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to compile:", err)
		return false
	}

	return true
}

func encryptComponents(pkgName string, enc *Encrypter) string {
	comps := strings.Split(pkgName, "/")
	for i, comp := range comps {
		comps[i] = enc.Encrypt(comp)
	}
	return strings.Join(comps, "/")
}
