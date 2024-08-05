package database

import (
	"strconv"
	"strings"
)

const COUNT_DEFAULT_EXPRESSION = "count(*)"

func buildCountExpression(columnName string) string {
	return "count(" + columnName + ")"
}

func buildSelectQuery(selectedFieldsList string, tableName string, conditions []string) string {
	clauses := []string{
		"SELECT",
		selectedFieldsList,
		"FROM",
		tableName,
	}

	if len(conditions) > 0 {
		encapsulatedConditions := []string{}
		for _, singleCondition := range conditions {
			encapsulatedConditions = append(encapsulatedConditions, "("+singleCondition+")")
		}

		clauses = append(clauses, "WHERE", strings.Join(encapsulatedConditions, " AND "))
	}

	return strings.Join(clauses, " ") + ";"
}

func buildSelectQueryWithEndClauses(selectedFieldsList string, tableName string, conditions []string, endClauses []string) string {
	clauses := []string{
		"SELECT",
		selectedFieldsList,
		"FROM",
		tableName,
	}

	if len(conditions) > 0 {
		encapsulatedConditions := []string{}
		for _, singleCondition := range conditions {
			encapsulatedConditions = append(encapsulatedConditions, "("+singleCondition+")")
		}

		clauses = append(clauses, "WHERE", strings.Join(encapsulatedConditions, " AND "))
	}

	if len(endClauses) > 0 {
		clauses = append(clauses, endClauses...)
	}

	return strings.Join(clauses, " ") + ";"
}

func buildLimitExpression(limit int) string {
	return "LIMIT " + strconv.Itoa(limit)
}

func buildLimitOffsetExpression(limit int, offset int) string {
	return "LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
}
