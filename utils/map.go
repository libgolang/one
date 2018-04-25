package utils

func CopyStringStringMap(m map[string]string) map[string]string {
	target := make(map[string]string)
	for k, v := range m {
		target[k] = v
	}
	return target
}
