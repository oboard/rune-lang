package tsdecl

// Flattened is an interface view with inherited members merged. Bases are
// applied breadth-first; earlier declarations win so a subtype can shadow a
// base member with a narrower signature.
type Flattened struct {
	Name    string
	Members []Member
}

// FlattenAll resolves every interface's inheritance chain, returning the
// flat member view keyed by interface name. Unknown bases are skipped; cycles
// are truncated so helper return is total.
func FlattenAll(dom *DOM) map[string]*Flattened {
	out := make(map[string]*Flattened, len(dom.Interfaces))
	visiting := map[string]bool{}
	var resolve func(name string) *Flattened
	resolve = func(name string) *Flattened {
		if flat := out[name]; flat != nil {
			return flat
		}
		iface := dom.Interfaces[name]
		if iface == nil {
			return nil
		}
		if visiting[name] {
			return &Flattened{Name: name, Members: append([]Member(nil), iface.Members...)}
		}
		visiting[name] = true
		flat := &Flattened{Name: name}
		seen := map[string]bool{}
		// Bases first so the subtype can shadow inherited members.
		for _, base := range iface.Bases {
			baseFlat := resolve(base)
			if baseFlat == nil {
				continue
			}
			for _, mem := range baseFlat.Members {
				if seen[mem.Name] {
					continue
				}
				seen[mem.Name] = true
				flat.Members = append(flat.Members, mem)
			}
		}
		// Own members come after bases so the subtype shadow wins. For
		// overloaded methods we keep all sibling declarations so the caller
		// can pick the best overload; only field/property duplicates are
		// dedup'd.
		for _, mem := range iface.Members {
			if !mem.IsMethod {
				if seen[mem.Name] {
					continue
				}
				seen[mem.Name] = true
			}
			flat.Members = append(flat.Members, mem)
		}
		delete(visiting, name)
		out[name] = flat
		return flat
	}
	for name := range dom.Interfaces {
		resolve(name)
	}
	return out
}

// FlattenedMembers returns the merged member list for one interface name.
func FlattenedMembers(dom *DOM, name string) []Member {
	flat := FlattenAll(dom)[name]
	if flat == nil {
		return nil
	}
	return flat.Members
}
