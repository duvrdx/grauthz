package config

import "github.com/neo4j/neo4j-go-driver/neo4j"

func SetupDatabase(uri string, token neo4j.AuthToken) (neo4j.Driver, error) {
	driver, err := neo4j.NewDriver(uri, token)
	if err != nil {
		return nil, err
	}

	return driver, nil
}

func NewSession(driver neo4j.Driver) (neo4j.Session, error) {
	session, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func RunTransaction(session neo4j.Session, query string) (neo4j.Result, error) {
	result, err := session.Run(query, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}
