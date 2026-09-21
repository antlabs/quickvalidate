package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antlabs/quickvalidate/pkg/generator"
	"github.com/antlabs/quickvalidate/pkg/parser"
)

func main() {
	var (
		inputPath  string
		outputPath string
		pkgName    string
	)
	flag.StringVar(&inputPath, "i", "", "Input file or directory path")
	flag.StringVar(&outputPath, "o", "", "Output file path")
	flag.StringVar(&pkgName, "pkg", "", "Package name (default: derived from input path)")
	flag.Parse()

	if inputPath == "" {
		fmt.Println("Error: Input path is required")
		flag.Usage()
		os.Exit(1)
	}
	if outputPath == "" {
		fmt.Println("Error: Output path is required")
		flag.Usage()
		os.Exit(1)
	}

	paths, err := getFilePaths(inputPath)
	if err != nil {
		fmt.Printf("Error getting file paths: %v\n", err)
		os.Exit(1)
	}

	pkg, err := parser.ParseFiles(paths, nil)
	if err != nil {
		fmt.Printf("Error parsing input: %v\n", err)
		os.Exit(1)
	}
	if pkgName != "" {
		pkg.Name = pkgName
	}
	if pkg.Name == "" {
		pkg.Name = derivePackageName(inputPath)
	}

	if len(pkg.Structs) == 0 {
		fmt.Println("No structs with validation tags found")
		os.Exit(1)
	}

	code, err := generator.NewGenerator(pkg).Generate()
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, code, 0o644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated validation code to %s\n", outputPath)
}

func derivePackageName(path string) string {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return filepath.Base(path)
	}
	return filepath.Base(filepath.Dir(path))
}

// getFilePaths returns the Go files to process, skipping generated files.
func getFilePaths(path string) ([]string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		if strings.HasSuffix(path, ".go") {
			return []string{path}, nil
		}
		return nil, fmt.Errorf("input file must be a Go file")
	}

	var paths []string
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if info.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		if strings.HasSuffix(name, "_gen.go") {
			return nil
		}
		paths = append(paths, p)
		return nil
	})
	return paths, err
}
