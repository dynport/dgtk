package main

type issuesBrowse struct {
}

func (r *issuesBrowse) Run() error {
	u, e := githubUrl()
	if e != nil {
		return e
	}
	u += "/issues"
	logger.Printf("opening %q", u)
	return openUrl(u)
}
