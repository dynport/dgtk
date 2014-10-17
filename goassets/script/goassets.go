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
	tpl := template.Must(template.New("assets").Parse(tpl))
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

const byteLength = 12

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
		if len(buffer) == byteLength {
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
	return make([]string, 0, byteLength)
}
const tpl = `package {{ .Pkg }}

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	builtAt        = time.Now()
	compiledAssets = assetIntFS{}
	assets         AssetFileSystem
)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)

type assetProxy struct {
	devPath string
}

func (a *assetProxy) Open(name string) (http.File, error) {
	return a.fileSystem().Open(name)
}

func (a *assetProxy) AssetNames() []string {
	return a.fileSystem().AssetNames()
}

func (a *assetProxy) fileSystem() AssetFileSystem {
	dbg.Printf("getting file system for %q", a.devPath)
	if a.devPath != "" {
		dbg.Printf("using dev path %s", a.devPath)
		stat, e := os.Stat(a.devPath)
		if e == nil && stat.IsDir() {
			assets = &assetOsFS{root: a.devPath}
			return assets
		} else {
			dbg.Printf("dev path %s does not exist", a.devPath)
		}
	} else {
		dbg.Printf("dev path seems to be empty")
	}
	return compiledAssets
}

func FileSystem(devPath string) AssetFileSystem {
	return &assetProxy{devPath: devPath}
}

type AssetFileSystem interface {
	Open(name string) (http.File, error)
	AssetNames() []string
}

func assetNames() (names []string) {
	return assets.AssetNames()
}

func readAsset(key string) ([]byte, error) {
	r, e := assets.Open(key)
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = r.Close()
	}()

	p, e := ioutil.ReadAll(r)
	if e != nil {
		return nil, e
	}
	return p, nil
}

func mustReadAsset(key string) []byte {
	p, e := readAsset(key)
	if e != nil {
		panic("could not read asset with key " + key + ": " + e.Error())
	}
	return p
}

type assetOsFS struct{ root string }

func (aFS assetOsFS) Open(name string) (http.File, error) {
	p := filepath.Join(aFS.root, name)
	dbg.Printf("opening local file %q", p)
	f, e := os.Open(p)
	if e != nil {
		dbg.Printf("ERROR reading local file: %q", name)
		return nil, e
	}
	return f, nil
}

func (aFS *assetOsFS) AssetNames() []string {
	names, e := filepath.Glob(aFS.root + "/*")
	if e != nil {
		log.Print(e)
	}
	return names
}

type assetIntFS map[string][]byte

type assetNode struct {
	name string
	data *bytes.Reader
	dir  bool

	children map[string]*assetNode
}

func addNode(root *assetNode, path string, content *bytes.Reader) error {
	node := root
	pathSegments := strings.Split(path, "/")
	if len(pathSegments) > 1 {
		for i := 0; i < len(pathSegments)-1; i++ {
			if val, ok := node.children[pathSegments[i]]; ok {
				node = val
			} else {
				newNode := &assetNode{name: pathSegments[i], dir: true, children: map[string]*assetNode{}}
				node.children[pathSegments[i]] = newNode
				node = newNode
			}
		}
	}
	filename := pathSegments[len(pathSegments)-1]
	if _, ok := node.children[filename]; ok {
		return fmt.Errorf("node %q already exists", filename)
	}
	node.children[filename] = &assetNode{name: filename, data: content}
	return nil
}

func (node *assetNode) Traverse(path []string) (*assetNode, error) {
	switch len(path) {
	case 0:
		return node, nil
	default:
		child, ok := node.children[path[0]]
		if !ok {
			return nil, os.ErrNotExist
		}
		return child.Traverse(path[1:])
	}
}

func (node *assetNode) Name() string {
	return node.name
}

func (node *assetNode) ModTime() time.Time {
	return builtAt
}

func (node *assetNode) Mode() os.FileMode {
	if node.dir {
		return 0755
	}
	return 0644
}

func (node *assetNode) Sys() interface{} {
	return nil
}

func (node *assetNode) Size() int64 {
	if node.dir {
		return 0
	}
	return int64(node.data.Len())
}

func (node *assetNode) IsDir() bool {
	return node.dir
}

func (node *assetNode) Readdir(count int) (stats []os.FileInfo, e error) {
	if !node.dir {
		return nil, nil
	}

	for _, child := range node.children {
		stat, e := child.Stat()
		if e != nil {
			return nil, e
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

func (node *assetNode) Stat() (os.FileInfo, error) {
	return node, nil
}

func (node *assetNode) Close() error {
	return nil
}

func (node *assetNode) Read(p []byte) (int, error) {
	return node.data.Read(p)
}

func (node *assetNode) Seek(offset int64, whence int) (int64, error) {
	if node.dir {
		return 0, nil
	}
	return node.data.Seek(offset, whence)
}

func (node *assetNode) Open(name string) (af http.File, e error) {
	dbg.Printf("opening tpl %s", name)
	if name == "." {
		return node, nil
	}
	name = strings.TrimPrefix(name, "/")
	nameSegments := strings.Split(name, "/")
	return node.Traverse(nameSegments)
}

func (afs assetIntFS) AssetNames() (names []string) {
	names = make([]string, 0, len(afs))
	for k, _ := range afs {
		names = append(names, k)
	}
	return names
}

func (afs assetIntFS) Open(name string) (af http.File, e error) {
	dbg.Printf("opening tpl %s", name)

	switch name {
	case "":
		name = "index.html"
	case "/":
		name = ""
	default:
		name = strings.TrimPrefix(name, "/")
	}

	// single asset referenced, load it directly
	if asset, found := afs[name]; found {
		reader, e := createReader(asset)
		af = &assetNode{data: reader, name: name}
		return af, e
	}

	// directory request?
	switch {
	case name == "":
		// ignore
	case !strings.HasSuffix(name, "/"):
		name += "/"
	}
	root := &assetNode{dir: true, name: ".", children: map[string]*assetNode{}}
	for k, v := range afs {
		if strings.HasPrefix(k, name) {
			reader, e := createReader(v)
			if e != nil {
				return nil, e
			}
			dbg.Printf("adding node %q", k)
			addNode(root, strings.TrimPrefix(k, name), reader)
		}
	}
	if len(root.children) > 0 {
		return root, nil
	}

	dbg.Printf("ERROR: index %s does not exist. known keys: %#v", name, afs.AssetNames())
	return nil, os.ErrNotExist
}

func createReader(data []byte) (*bytes.Reader, error) {
	decomp, e := gzip.NewReader(bytes.NewBuffer(data))
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = decomp.Close()
	}()
	b, e := ioutil.ReadAll(decomp)
	if e != nil {
		return nil, e
	}
	return bytes.NewReader(b), nil
}

func init() {
        {{ range .Assets }}compiledAssets["{{ .Key }}"] = []byte{
                {{ .Bytes }}
        }
        {{ end }}
}
`
