package main

type Candidate struct {
	Flamingo string `json:"flamingo"`
	ArgoCD   string `json:"argocd"`
	Image    string `json:"image"`
	Flux     string `json:"flux"`
}

type CandidateList struct {
	Candidates []Candidate `json:"candidates"`
}
