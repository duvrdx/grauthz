package services

import (
	"fmt"
	"strings"

	"github.com/duvrdx/grauthz/pkg/config"
	"github.com/duvrdx/grauthz/pkg/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func BuildClause(clause, objectType string, auditable bool, policy *models.Policy) (string, error) {
	clause = strings.TrimSpace(clause)
	query := ""

	if !strings.Contains(clause, " from ") {
		relation := policy.GetRelation(objectType, clause)

		if relation == nil {
			return "", fmt.Errorf("relation %s not found for type %s", clause, objectType)
		}

		query = fmt.Sprintf("\tEXISTS((s)-[:%s]->(o))", strings.ToUpper(relation.Name))

		if auditable {
			query = fmt.Sprintf("%s AS %s", query, strings.ReplaceAll(clause, " ", "_"))
		}
	} else {
		splitClause := strings.Split(clause, " from ")

		if len(splitClause) != 2 {
			return "", fmt.Errorf("invalid clause format: %s", clause)
		}

		relationName := strings.TrimSpace(splitClause[0])
		relatedType := strings.TrimSpace(splitClause[1])

		relation := policy.GetRelation(objectType, relatedType)
		if relation == nil {
			return "", fmt.Errorf("relation %s not found for type %s", relationName, objectType)
		}

		query = fmt.Sprintf("\tEXISTS((s)-[:%s]-(:%s)-[:%s*..1000]->(o))", strings.ToUpper(relationName), capitalizeFL(relation.SubjectType),
			strings.ToUpper(relatedType))

		if auditable {
			query = fmt.Sprintf("%s AS %s", query, strings.ReplaceAll(clause, " ", "_"))
		}
	}

	return query, nil
}

func CountQuery(subjectType, subjectID, action, objectType string, policy *models.Policy, session neo4j.Session) (int, error) {
	actionEntity := policy.GetAction(objectType, action)
	if actionEntity == nil {
		return 0, fmt.Errorf("action %s not found", action)
	}

	clauses := strings.Split(actionEntity.Rule, " or")

	var query strings.Builder

	// Inicia a construção da consulta
	query.WriteString(fmt.Sprintf("MATCH (s:%s {id: '%s'})\n", capitalizeFL(subjectType), subjectID))
	query.WriteString(fmt.Sprintf("MATCH (o:%s)\n", capitalizeFL(objectType)))
	query.WriteString("WHERE ")

	var results []string
	for _, clause := range clauses {
		nClause, err := BuildClause(clause, objectType, false, policy)
		if err != nil {
			return 0, err
		}
		results = append(results, nClause)
	}

	// Constrói a parte WHERE da consulta
	query.WriteString(strings.Join(results, " OR \n"))

	// Adiciona a contagem
	query.WriteString("\nRETURN COUNT(DISTINCT o.id) AS total")

	countQuery := query.String()

	result, err := config.RunTransaction(session, countQuery)
	if err != nil {
		return 0, err
	}

	// Verifica o resultado da contagem
	if result.Next() {
		record := result.Record()
		if record == nil {
			return 0, fmt.Errorf("no count result found")
		}

		total, ok := record.Get("total")
		if !ok {
			return 0, fmt.Errorf("count result not found in record")
		}

		return int(total.(int64)), nil
	}

	return 0, fmt.Errorf("no records found")
}

func AccessQuery(subjectType, subjectID, action, objectType, objectID string, policy *models.Policy) (string, error) {
	actionEntity := policy.GetAction(objectType, action)
	if actionEntity == nil {
		return "", fmt.Errorf("action %s not found", action)
	}

	clauses := strings.Split(actionEntity.Rule, " or")

	var query strings.Builder

	query.WriteString(fmt.Sprintf("MATCH (s:%s {id: '%s'})\n", capitalizeFL(subjectType), subjectID))
	query.WriteString(fmt.Sprintf("MATCH (o:%s {id: '%s'})\n", capitalizeFL(objectType), objectID))
	query.WriteString("RETURN\n")

	var results []string
	for _, clause := range clauses {
		nClause, err := BuildClause(clause, objectType, true, policy)

		if err != nil {
			return "", err
		}

		results = append(results, nClause)
	}

	query.WriteString(strings.Join(results, ",\n"))

	return query.String(), nil
}

func FilterQuery(subjectType, subjectID, action, objectType string, policy *models.Policy) (string, error) {
	actionEntity := policy.GetAction(objectType, action)
	if actionEntity == nil {
		return "", fmt.Errorf("action %s not found", action)
	}

	clauses := strings.Split(actionEntity.Rule, " or")

	var query strings.Builder

	query.WriteString(fmt.Sprintf("MATCH (s:%s {id: '%s'})\n", capitalizeFL(subjectType), subjectID))
	query.WriteString(fmt.Sprintf("MATCH (o:%s)\n", capitalizeFL(objectType)))
	query.WriteString("WHERE ")

	var results []string
	for _, clause := range clauses {
		nClause, err := BuildClause(clause, objectType, false, policy)

		if err != nil {
			return "", err
		}

		results = append(results, nClause)
	}

	query.WriteString(strings.Join(results, " OR\n"))

	query.WriteString("\nRETURN DISTINCT o.id\n")

	return query.String(), nil
}

func FilterQueryPaginated(subjectType, subjectID, action, objectType string, policy *models.Policy, session neo4j.Session, limit, actualPage int) (string, int, error) {
	if actualPage <= 0 {
		return "", 0, fmt.Errorf("page must be greater than 0")
	}

	cacheKey := fmt.Sprintf("%s_%s_%s_%s_total", subjectType, subjectID, action, objectType)

	cacheValue, found := config.GetCache(cacheKey)
	var totalInt int
	if found {
		var ok bool
		totalInt, ok = cacheValue.(int)
		if !ok {
			return "", 0, fmt.Errorf("cache value for key %s is not an int", cacheKey)
		}
	} else {
		var err error
		totalInt, err = CountQuery(subjectType, subjectID, action, objectType, policy, session)
		if err != nil {
			return "", 0, err
		}

		config.SetCache(cacheKey, totalInt)
	}

	totalPages := (totalInt + limit - 1) / limit

	if actualPage > totalPages {
		return "", 0, fmt.Errorf("page %d out of range", actualPage)
	}

	filterQuery, err := FilterQuery(subjectType, subjectID, action, objectType, policy)
	if err != nil {
		return "", 0, err
	}

	paginatedQuery := fmt.Sprintf("%s SKIP %d LIMIT %d", filterQuery, (actualPage-1)*limit, limit)

	return paginatedQuery, totalPages, nil
}
