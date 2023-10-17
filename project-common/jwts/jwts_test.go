package jwts

import "testing"

func TestParseToken(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTgwODUzNDQsInRva2VuIjoiMTAzNCJ9.vgvu4fXqvcyC-4fQG2QEbi-3zYg8eBHZKDC8L3mhxVM"
	ParseToken(tokenString, "msproject")
}
