package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const TPL = ""

var logger = log.New(os.Stderr, "", 0)

func main() {
	var targetFile = flag.String("file", "assets.go", "path of asset file")
	var pkgName = flag.String("pkg", "", "Package name (default: name of directory)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s <AssetPaths>...\n", "goassets.go")
		flag.PrintDefaults()
	}
	flag.Parse()
	a := &action{targetFile: *targetFile, assetPaths: flag.Args(), pkgName: *pkgName}
	if len(a.assetPaths) < 1 {
		flag.Usage()
		return
		//logger.Fatal("USAGE go run goassets.go <AssetPaths>...")
	}
	if e := a.run(); e != nil {
		logger.Fatal(e)
	}
}

type action struct {
	targetFile string
	pkgName    string
	assetPaths []string
}

func (a *action) run() error {
	var e error

	packageName := a.pkgName

	if packageName == "" {
		packageName, e = determinePackageByPath(a.targetFile)
		if e != nil {
			return e
		}
	}

	assets := &assets{
		Pkg:               packageName,
		customPackagePath: a.targetFile,
		paths:             a.assetPaths,
	}

	if e := assets.build(); e != nil {
		return e
	}

	return nil
}

func determinePackageByPath(targetFile string) (string, error) {
	wd, e := os.Getwd()
	if e != nil {
		return "", e
	}
	defer os.Chdir(wd)
	e = os.Chdir(path.Dir(targetFile))
	if e != nil {
		return "", e
	}
	result, e := exec.Command("go", "list", "-f", "{{ .Name }}").CombinedOutput()
	if e != nil {
		wd, e2 := os.Getwd()
		if e2 != nil {
			return "", e2
		}
		return path.Base(wd), nil
	}
	return strings.TrimSpace(string(result)), nil
}

type assets struct {
	Pkg               string
	customPackagePath string
	Assets            []*asset
	paths             []string
	builtAt           string
}

func (assets *assets) Bytes() (b []byte, e error) {
	tpl := template.Must(template.New("assets").Parse(TPL))
	buf := &bytes.Buffer{}
	assets.builtAt = time.Now().UTC().Format(time.RFC3339Nano)
	e = tpl.Execute(buf, assets)
	if e != nil {
		return b, e
	}
	return buf.Bytes(), nil
}

func (assets *assets) assetPaths() (out []*asset, e error) {
	out = []*asset{}
	packagePath, e := assets.packagePath()
	if e != nil {
		return out, e
	}
	for _, path := range assets.paths {
		tmp, e := assetsInPath(path, packagePath)
		if e != nil {
			return out, e
		}
		for _, asset := range tmp {
			asset.Key, e = removePrefix(asset.Path, path)
			if e != nil {
				return out, e
			}
			out = append(out, asset)
		}
	}
	return out, nil
}

func removePrefix(path, prefix string) (suffix string, e error) {
	absPath, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	absPrefix, e := filepath.Abs(prefix)
	if e != nil {
		return "", e
	}
	if strings.HasPrefix(absPath, absPrefix) {
		return strings.TrimPrefix(strings.TrimPrefix(absPath, absPrefix), "/"), nil
	}
	return "", fmt.Errorf("%s has no prefix %s", absPath, absPrefix)
}

func assetsInPath(path string, packagePath string) (assets []*asset, e error) {
	e = filepath.Walk(path, func(p string, stat os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if stat.IsDir() {
			return nil
		}
		abs, e := filepath.Abs(p)
		if e != nil {
			return e
		}
		if abs != packagePath {
			assets = append(assets, &asset{Path: p})
		}
		return nil
	})
	return assets, e
}

func (assets *assets) packagePath() (path string, e error) {
	path = assets.customPackagePath
	if path == "" {
		path = "./assets.go"
	}
	return filepath.Abs(path)
}

const BYTE_LENGTH = 12

type asset struct {
	Path  string
	Key   string
	Name  string
	Bytes string
}

func (asset *asset) Load() error {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	f, e := os.Open(asset.Path)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = io.Copy(gz, f)
	gz.Flush()
	gz.Close()
	if e != nil {
		return e
	}
	list := make([]string, 0, len(buf.Bytes()))
	for _, b := range buf.Bytes() {
		list = append(list, fmt.Sprintf("0x%x", b))
	}
	buffer := makeLineBuffer()
	asset.Name = path.Base(asset.Path)
	for _, b := range list {
		buffer = append(buffer, b)
		if len(buffer) == BYTE_LENGTH {
			asset.Bytes += strings.Join(buffer, ",") + ",\n"
			buffer = makeLineBuffer()
		}
	}
	if len(buffer) > 0 {
		asset.Bytes += strings.Join(buffer, ",") + ",\n"
	}
	return nil
}

var debugger = log.New(debugStream(), "", 0)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

func (assets *assets) doBuild() ([]byte, error) {
	if assets.Pkg == "" {
		assets.Pkg = "main"
	}
	debugger.Print("loading assets paths")
	paths, e := assets.assetPaths()
	debugger.Printf("got %d assets", len(paths))
	if e != nil {
		return nil, e
	}
	for _, asset := range paths {
		debugger.Printf("loading assets %q", asset.Key)
		e := asset.Load()
		if e != nil {
			return nil, e
		}
		assets.Assets = append(assets.Assets, asset)
	}
	return assets.Bytes()
}

func (assets *assets) build() error {
	b, e := assets.doBuild()
	if e != nil {
		return e
	}
	path, e := assets.packagePath()
	if e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		if os.IsExist(e) {
			return fmt.Errorf("File %q already exists (deleted it first?!?)", path)
		}
		return e
	}
	defer f.Close()
	_, e = f.Write(b)
	return e
}

func makeLineBuffer() []string {
	return make([]string, 0, BYTE_LENGTH)
}
