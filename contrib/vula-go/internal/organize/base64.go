package organize

import (
	"encoding/base64"
)

/*
func base64ToShortString(b []byte) string {
	s := ""
	if len(b) > 6 {
		s = base64.StdEncoding.EncodeToString(b[:6])[:6]
	} else {
		s = base64.StdEncoding.EncodeToString(b)
	}
	return fmt.Sprintf("<b64:%s...(%d)>", s, len(b))
}
*/

func toBase64String(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
