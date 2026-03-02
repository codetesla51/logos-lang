package interpreter

// GoToObject converts a native Go value to a Logos Object.
// Supported types: int, int64, float64, string, bool, []interface{}, map[string]interface{}, nil
func GoToObject(v interface{}) Object {
	if v == nil {
		return NULL
	}
	switch val := v.(type) {
	case int:
		return &Integer{Value: int64(val)}
	case int64:
		return &Integer{Value: val}
	case float64:
		return &Float{Value: val}
	case string:
		return &String{Value: val}
	case bool:
		if val {
			return TRUE
		}
		return FALSE
	case []interface{}:
		elements := make([]Object, len(val))
		for i, el := range val {
			elements[i] = GoToObject(el)
		}
		return &Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[string]Object)
		for k, v2 := range val {
			pairs["STRING:"+k] = GoToObject(v2)
		}
		return &Table{Pairs: pairs}
	default:
		// For unsupported types, return null
		return NULL
	}
}

// ObjectToGo converts a Logos Object back to a native Go value.
// Returns nil for unsupported types or NULL objects.
func ObjectToGo(obj Object) interface{} {
	if obj == nil {
		return nil
	}
	switch val := obj.(type) {
	case *Integer:
		return val.Value
	case *Float:
		return val.Value
	case *String:
		return val.Value
	case *Bool:
		return val.Value
	case *Null:
		return nil
	case *Array:
		result := make([]interface{}, len(val.Elements))
		for i, el := range val.Elements {
			result[i] = ObjectToGo(el)
		}
		return result
	case *Table:
		result := make(map[string]interface{})
		for k, v := range val.Pairs {
			// Keys are stored as "TYPE:value", extract just the value part
			parts := splitKey(k)
			result[parts] = ObjectToGo(v)
		}
		return result
	case *ReturnValue:
		return ObjectToGo(val.Value)
	default:
		return nil
	}
}

// splitKey extracts the key value from a table key format "TYPE:value"
func splitKey(k string) string {
	for i := 0; i < len(k); i++ {
		if k[i] == ':' {
			return k[i+1:]
		}
	}
	return k
}
