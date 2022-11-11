package snow

import (
	"fmt"
	"time"
)

func ToUnixString(date string) string {
	t, err := time.Parse("02/01/2006 15:04:05", date)

	if err != nil {
		return ""
	}

	return fmt.Sprintf("%d", t.Unix())
}
