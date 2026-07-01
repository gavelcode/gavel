package catalog

// BinaryNames lists the tool-binary repos the given languages depend on, in
// catalog order and de-duplicated (cpd shares pmd's binary). Drives gavel init's
// informational output; the repos and versions are owned by gavel-tools.
func BinaryNames(languages []string) []string {
	return active().binaryNames(languages)
}
