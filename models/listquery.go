package models

// buildOrderBy resolves a caller-supplied sort key against a whitelist of
// column names, returning a safe "ORDER BY" fragment. This is the only
// place user input influences column names in a query, and it only ever
// selects a value already present in `allowed` -- never the raw input --
// so there's no SQL injection risk.
func buildOrderBy(sortKey, dirKey string, allowed map[string]string, defaultOrderBy string) string {
	col, ok := allowed[sortKey]
	if !ok {
		return defaultOrderBy
	}
	dir := "ASC"
	if dirKey == "desc" {
		dir = "DESC"
	}
	return col + " " + dir
}
