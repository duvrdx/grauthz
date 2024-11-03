package models

type AccessResult struct {
	Clause string `json:"clause"`
	Ok     bool   `json:"ok"`
}

type FilterResult struct {
	Identifiers []string `json:"identifiers"`
}

type PaginatedFilterResult struct {
	Total       int            `json:"total"`
	CurrentPage int            `json:"current_page"`
	Data        []FilterResult `json:"data"`
}
