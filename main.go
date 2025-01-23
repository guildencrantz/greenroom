package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"github.com/cmars/greenroom/pkg/mkdocs"
	"github.com/princjef/gomarkdoc"
	"github.com/princjef/gomarkdoc/lang"
	"github.com/princjef/gomarkdoc/logger"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v3"
)

var (
	outputDir = flag.String("output", ".", "directory where mkdocs documentation will be created")
	siteName  = flag.String("site-name", "", "override mkdocs site_name (default is module dirname)")
	addNav    = flag.Bool("add-nav", false, "add flattened nav links. If you don't specify any existing nav this will prevent accessing your other documents")
)

func main() {
	flag.Parse()

	docsRoot := filepath.Join(*outputDir, "docs")

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
	}, flag.Args()...)
	if err != nil {
		log.Fatalf("failed to load packages: %v", err)
	}

	mkDocs := mkdocs.NewMkDocs()
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			// skip main packages (binaries)
			continue
		}
		if strings.HasSuffix(pkg.Name, "_test") {
			// skip test packages
			continue
		}
		if pkg.Module == nil {
			log.Printf("warning: failed to resolve package %v", pkg)
			log.Printf("try `go get -u %v`?", pkg)
			continue
		}
		log.Printf("%s %#v", pkg.PkgPath, pkg.Module)
		err := func() error {
			doc := &PackageDoc{pkg}
			docFilePath := doc.DocBase(docsRoot) + ".md"
			docDir := filepath.Dir(docFilePath)
			err := os.MkdirAll(docDir, 0777)
			if err != nil {
				return err
			}
			docFile, err := os.Create(docFilePath)
			if err != nil {
				return err
			}
			defer docFile.Close()
			err = doc.Generate(docFile)
			if err != nil {
				return err
			}

			if mkDocs.SiteName == "" {
				mkDocs.SiteName = filepath.Base(pkg.Module.Dir)
			}
			docLinkPath := doc.DocBase(".")

			if *addNav {
				mkDocs.Nav = append(mkDocs.Nav, mkdocs.NavItem{
					Name: pkg.PkgPath,
					Path: docLinkPath + ".md",
				})
			}
			return nil
		}()
		if err != nil {
			log.Printf("warning: failed to generate docs for package %q: %v", pkg.PkgPath, err)
		}
	}

	mkDocsPath := filepath.Join(*outputDir, "mkdocs.yml")

	existing := mkdocs.MkDocs{}
	if existingContent, err := os.ReadFile(mkDocsPath); err == nil {
		err = yaml.Unmarshal(existingContent, &existing)
		if err != nil {
			log.Fatal(err)
		}
	} else if !os.IsNotExist(err) {
		log.Fatal(err)
	}

	err = mergo.Merge(&existing, mkDocs)
	if err != nil {
		log.Fatal(err)
	}

	mkDocsFile, err := os.Create(mkDocsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer mkDocsFile.Close()

	mkDocsContent, err := yaml.Marshal(&existing)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(mkDocsFile, bytes.NewBuffer(mkDocsContent))
	if err != nil {
		log.Fatal(err)
	}
}

type PackageDoc struct {
	*packages.Package
}

func (p *PackageDoc) DocBase(docsRoot string) string {
	return filepath.Join(docsRoot, p.Package.PkgPath)
}

func (p *PackageDoc) PackageDir() string {
	return strings.Replace(p.Package.PkgPath, p.Package.Module.Path, p.Package.Module.Dir, 1)
}

func (p *PackageDoc) Generate(w io.Writer) error {
	out, err := gomarkdoc.NewRenderer()
	if err != nil {
		return err
	}
	pkgDir := p.PackageDir()
	buildPkg, err := build.ImportDir(pkgDir, build.ImportComment)
	if err != nil {
		return fmt.Errorf("failed to import package from %s: %v", pkgDir, err)
	}

	log := logger.New(logger.DebugLevel)
	langPkg, err := lang.NewPackageFromBuild(log, buildPkg)
	if err != nil {
		return fmt.Errorf("failed to extract docs from package %s: %v", p.PkgPath, err)
	}

	doc, err := out.Package(langPkg)
	if err != nil {
		return fmt.Errorf("failed to render docs for package %s: %v", p.PkgPath, err)
	}

	_, err = w.Write([]byte(doc))
	return err
}
