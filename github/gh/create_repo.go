package main

type createRepo struct {
	Orga string `cli:"opt --orga"`
	Name string `cli:"opt --name"`
}

func (r *createRepo) Run() error {
	cl, err := client()
	if err != nil {
		return err
	}
	_ = cl
	return nil
}

type createRepoInput struct {
	Org     string `json:"-"`
	Repo    string `json:"-"`
	Name    string `json:"name,omitempty"`
	Private bool   `json:"private,omitempty"`
	TeamID  int    `json:"team_id,omitempty"`
}
