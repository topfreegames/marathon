package util

// SliceContains tells if a slice contains a string
func SliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// SliceRemove removes the first found element from a slice and returns if the element was found
func SliceRemove(s []string, e string) bool {
	for i, a := range s {
		if a == e {
			s = append(s[:i], s[i+1:]...)
			return true
		}
	}
	return false
}
