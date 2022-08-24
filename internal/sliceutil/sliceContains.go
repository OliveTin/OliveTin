package stringslice

func Contains(slice []string, search string) bool {
	for _, candidate := range slice {
		if search == candidate {
			return true
		}
	}

	return false
}

func Append(slice []string, element string) {
	slice = append(slice, element)
}
