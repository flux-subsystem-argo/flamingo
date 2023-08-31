package main

type Candidate struct {
	ArgoCD string `json:"argocd"`
	Fsa    string `json:"fsa"`
	Flux   string `json:"flux"`
}

type CandidateList struct {
	Candidates []Candidate `json:"candidates"`
}
