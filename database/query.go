package database

import (
	"strconv"
	"strings"
)

const COUNT_DEFAULT_EXPRESSION = "count(*)"

func buildCountExpression(columnName string) string {
	return "count(" + columnName + ")"
}

func buildSelectTableExpression(fieldsList string, tableName string) string {
	return strings.Join([]string{
		"SELECT",
		fieldsList,
		"FROM",
		tableName,
	}, " ")
}

func buildWhereExpression(conditions []string) string {
	if len(conditions) > 0 {
		encapsulatedConditions := []string{}
		for _, singleCondition := range conditions {
			encapsulatedConditions = append(encapsulatedConditions, "("+singleCondition+")")
		}

		return "WHERE " + strings.Join(encapsulatedConditions, " AND ")
	}

	return ""
}

func buildLimitExpression(limit int) string {
	return "LIMIT " + strconv.Itoa(limit)
}

func buildLimitOffsetExpression(limit int, offset int) string {
	return "LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)
}

func joinAllExpressions(expressions []string) string {
	return strings.Join(expressions, " ")
}
