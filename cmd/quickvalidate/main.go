package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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

	// Get file paths to process
	filePaths, err := getFilePaths(inputPath)
	if err != nil {
		fmt.Printf("Error getting file paths: %v\n", err)
		os.Exit(1)
	}

	// Parse files
	var allStructs []parser.StructInfo
	for _, filePath := range filePaths {
		structs, err := parser.ParseFile(filePath)
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", filePath, err)
			continue
		}
		allStructs = append(allStructs, structs...)
	}

	if len(allStructs) == 0 {
		fmt.Println("No structs with validation tags found")
		os.Exit(1)
	}

	// Determine package name if not provided
	if pkgName == "" {
		if stat, err := os.Stat(inputPath); err == nil && stat.IsDir() {
			pkgName = filepath.Base(inputPath)
		} else {
			dir := filepath.Dir(inputPath)
			pkgName = filepath.Base(dir)
		}
	}

	// Generate validation code
	gen := generator.NewGenerator(pkgName, allStructs)
	code, err := gen.Generate()
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write to output file
	err = ioutil.WriteFile(outputPath, code, 0644)
	if err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated validation code to %s\n", outputPath)
}

// getFilePaths returns a list of Go files to process
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

	var filePaths []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	return filePaths, err
}
