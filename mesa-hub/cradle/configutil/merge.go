package configutil

// MergeConfigLayers 将多个配置层合并，后者覆盖前者
func MergeConfigLayers(layers ...ConfigLayer) ConfigLayer {
	result := ConfigLayer{
		Env:   make(map[string]string),
		Files: []ConfigFile{},
	}
	fileMap := make(map[string]string)

	for _, layer := range layers {
		for k, v := range layer.Env {
			if v != "" {
				result.Env[k] = v
			}
		}
		for _, f := range layer.Files {
			if f.Content != "" {
				fileMap[f.Path] = f.Content
			}
		}
	}

	for path, content := range fileMap {
		result.Files = append(result.Files, ConfigFile{Path: path, Content: content})
	}
	return result
}
