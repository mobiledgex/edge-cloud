package util

// Make a copy of a string map
// if the source map is nil, the result is an empty map
func CopyStringMap(srcM map[string]string) map[string]string {
	mapCopy := map[string]string{}
	if srcM == nil {
		return mapCopy
	}
	for k, v := range srcM {
		mapCopy[k] = v
	}
	return mapCopy
}

func AddMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
