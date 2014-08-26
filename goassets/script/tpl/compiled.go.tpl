
func init() {
        {{ range .Assets }}compiledAssets["{{ .Key }}"] = []byte{
                {{ .Bytes }}
        }
        {{ end }}
}
