package utils

func ValidParams(url, token, rfc, aplicacion, cltid, perid string) bool {
	if url == "" || token == "" || rfc == "" || aplicacion == "" || cltid == "" || perid == "" {
		return false
	}
	return true
}
