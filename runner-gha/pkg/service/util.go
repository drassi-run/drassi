package service

import "time"

func renewAt(t time.Time) time.Duration {
	d := time.Until(t)
	if d <= 0 {
		return 0
	}

	// Renew when 3/4 time pass or 1 minute before expire, whichever later
	d1 := d * 3 / 4
	d2 := d - time.Minute
	if d1 < d2 {
		return d2
	} else {
		return d1
	}
}
