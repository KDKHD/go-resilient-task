package postgres

const (
	PostgresUniqueViolation     = "23505"
	PostgresForeignKeyViolation = "23503"
	PostgresCheckViolation      = "23514"
	PostgresNotNullViolation    = "23502"
	PostgresSyntaxError         = "42601"
	PostgresInvalidTextRep      = "22P02"
	PostgresInternalError       = "XX000"
)
