package vmware

func CleanupTags() error {
	tags, e := LoadTags()
	if e != nil {
		return e
	}

	vms, e := All()
	if e != nil {
		return e
	}
	m := map[string]struct{}{}

	for _, vm := range vms {
		m[vm.Id()] = struct{}{}
	}

	filtered := Tags{}

	for _, tag := range tags {
		if _, ok := m[tag.VmId]; ok {
			filtered = append(filtered, tag)
		}
	}
	return StoreTags(filtered)
}
