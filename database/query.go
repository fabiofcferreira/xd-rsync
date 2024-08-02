package database

import "strings"

func BuildSelectQuery(selectedFieldsList string, tableName string, conditions []string) string {
	parts := []string{
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

		parts = append(parts, "WHERE", strings.Join(encapsulatedConditions, " AND "))
	}

	return strings.Join(parts, " ") + ";"
}
