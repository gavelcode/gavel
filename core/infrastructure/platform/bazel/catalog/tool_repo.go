package catalog

func BinaryNames(languages []string) []string {
	return active().binaryNames(languages)
}
