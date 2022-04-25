package main

func main() {
	RegisterFunction("keep_keys", "GetSetter", "[]string")
}

func keepKeys(target GetSetter, keys []string) ExprFunc {
	keySet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	return func(ctx TransformContext) interface{} {
		val := target.Get(ctx)
		if val == nil {
			return nil
		}

		if attrs, ok := val.(Map); ok {
			attrs.RemoveIf(func(key string, _ Value) bool {
				if _, ok := keySet[key]; ok {
					return false
				}
				return true
			})
		}
		return nil
	}
}
