package lolesport

import "time"

const dateLayout = "2006-01-02"

type Date struct{ time.Time }

func (d *Date) UnmarshalJSON(b []byte) error {
	// Trim surrounding quotes.
	s := string(b)
	s = s[1 : len(s)-1]

	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return err
	}

	*d = Date{t}

	return nil
}
