package models

type Relation struct {
	Name        string `json:"name"`
	SubjectType string `json:"type"`
}

type Action struct {
	Name string `json:"name"`
	Rule string `json:"rule"`
}

type Type struct {
	Name      string     `json:"name"`
	Relations []Relation `json:"relations"`
	Actions   []Action   `json:"actions"`
}

type Policy struct {
	Namespace string `json:"namespace"`
	Types     []Type `json:"types"`
}

func (p *Policy) GetRelationsByType(typeName string) []Relation {
	for _, t := range p.Types {
		if t.Name == typeName {
			return t.Relations
		}
	}
	return nil
}

func (p *Policy) GetActionsByType(typeName string) []Action {
	for _, t := range p.Types {
		if t.Name == typeName {
			return t.Actions
		}
	}
	return nil
}

func (p *Policy) GetAction(typeName, actionName string) *Action {
	actions := p.GetActionsByType(typeName)
	for _, a := range actions {
		if a.Name == actionName {
			return &a
		}
	}
	return nil
}

func (p *Policy) GetRelation(typeName, relationName string) *Relation {
	relations := p.GetRelationsByType(typeName)
	for _, r := range relations {
		if r.Name == relationName {
			return &r
		}
	}
	return nil
}
