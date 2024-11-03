package services

import (
	"fmt"

	"github.com/duvrdx/grauthz/pkg/config"
	"github.com/duvrdx/grauthz/pkg/models"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func CheckAccess(subjectType, subjectID, action, objectType, objectID string, policy *models.Policy, session neo4j.Session) ([]models.AccessResult, error) {
	query, err := AccessQuery(subjectType, subjectID, action, objectType, objectID, policy)
	audits := make([]models.AccessResult, 0)

	if err != nil {
		return audits, err
	}

	result, err := config.RunTransaction(session, query)

	if err != nil {
		return audits, err
	}

	if result.Next() {
		record := result.Record()
		if record == nil {
			return audits, err
		}

		keys := record.Keys()

		for _, key := range keys {
			value, _ := record.Get(key)

			audits = append(audits, models.AccessResult{Clause: key, Ok: value.(bool)})
		}
		return audits, nil
	}

	return audits, fmt.Errorf("any record found")
}

func FilterAccess(subjectType, subjectID, action, objectType string, policy *models.Policy, session neo4j.Session) (models.FilterResult, error) {
	query, err := FilterQuery(subjectType, subjectID, action, objectType, policy)
	identifiers := make([]string, 0)

	if err != nil {
		return models.FilterResult{}, err
	}

	result, err := config.RunTransaction(session, query)

	if err != nil {
		return models.FilterResult{}, err
	}

	for result.Next() {
		record := result.Record()
		if record == nil {
			return models.FilterResult{}, err
		}

		identifier, _ := record.Get("o.id")

		if identifier == nil {
			return models.FilterResult{}, fmt.Errorf("no identifier found")
		}

		identifiers = append(identifiers, identifier.(string))
	}

	return models.FilterResult{Identifiers: identifiers}, nil
}

func FilterAccessPaginated(subjectType, subjectID, action, objectType string, policy *models.Policy, session neo4j.Session, limit, actualPage int) (models.PaginatedFilterResult, error) {
	query, totalPages, err := FilterQueryPaginated(subjectType, subjectID, action, objectType, policy, session, limit, actualPage)

	if err != nil {
		return models.PaginatedFilterResult{}, err
	}

	result, err := config.RunTransaction(session, query)
	if err != nil {
		return models.PaginatedFilterResult{}, err
	}

	identifiers := make([]string, 0)
	for result.Next() {
		record := result.Record()
		if record == nil {
			return models.PaginatedFilterResult{}, fmt.Errorf("error reading record")
		}

		identifier, _ := record.Get("o.id")
		identifiers = append(identifiers, identifier.(string))
	}

	return models.PaginatedFilterResult{Total: totalPages, CurrentPage: actualPage, Data: []models.FilterResult{{Identifiers: identifiers}}}, nil
}
