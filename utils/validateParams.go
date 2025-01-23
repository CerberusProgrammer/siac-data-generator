package utils

func ValidParams(url, token, rfc, aplicacion string, cltid, perid int) bool {
	if url == "" || token == "" || rfc == "" || aplicacion == "" || cltid == 0 || perid == 0 {
		return false
	}
	return true
}
