package configutil

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ApplyEnvOverrides 基于 yaml tag 递归遍历配置结构，并按 PREFIX_FIELD_NAME 规则自动应用环境变量。
// 之所以使用反射而不是手写映射，是为了在新增配置字段时默认具备 env override 能力，减少重复样板代码。
func ApplyEnvOverrides(prefix string, cfg any) error {
	if cfg == nil {
		return errors.New("config target is nil")
	}

	value := reflect.ValueOf(cfg)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("config target must be a non-nil pointer")
	}

	return applyEnvReflect(strings.ToUpper(prefix), value.Elem(), nil)
}

// ApplyStringOverride 在环境变量存在时覆盖目标字符串。
func ApplyStringOverride(target *string, envKey string) {
	if target == nil {
		return
	}
	if value := os.Getenv(envKey); value != "" {
		*target = value
	}
}

// ApplyBoolOverride 在环境变量可被解析时覆盖目标布尔值。
func ApplyBoolOverride(target *bool, envKey string) {
	if target == nil {
		return
	}
	if value := os.Getenv(envKey); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			*target = parsed
		}
	}
}

// ApplyIntOverride 在环境变量可被解析为正整数时覆盖目标值。
func ApplyIntOverride(target *int, envKey string) {
	if target == nil {
		return
	}
	if value := os.Getenv(envKey); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			*target = parsed
		}
	}
}

// ApplyCSVOverride 将逗号分隔的环境变量值拆成切片，并去除空白与空项。
func ApplyCSVOverride(target *[]string, envKey string) {
	if target == nil {
		return
	}
	if value := os.Getenv(envKey); value != "" {
		*target = SplitCSV(value)
	}
}

// SplitCSV 将逗号分隔字符串拆成去空白后的切片。
func SplitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func applyEnvReflect(prefix string, value reflect.Value, parents []string) error {
	if value.Kind() != reflect.Struct {
		return errors.New("config target must point to a struct")
	}

	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := valueType.Field(i)
		if !field.CanSet() {
			continue
		}

		tagName := yamlTagName(fieldType)
		if tagName == "-" {
			continue
		}

		fieldParents := append([]string{}, parents...)
		if tagName != "" {
			fieldParents = append(fieldParents, strings.ToUpper(tagName))
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := applyEnvReflect(prefix, field, fieldParents); err != nil {
				return err
			}
		case reflect.String:
			if value := os.Getenv(joinEnvKey(prefix, fieldParents)); value != "" {
				field.SetString(value)
			}
		case reflect.Bool:
			if value := os.Getenv(joinEnvKey(prefix, fieldParents)); value != "" {
				parsed, err := strconv.ParseBool(value)
				if err != nil {
					return fmt.Errorf("parse bool env %s: %w", joinEnvKey(prefix, fieldParents), err)
				}
				field.SetBool(parsed)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value := os.Getenv(joinEnvKey(prefix, fieldParents)); value != "" {
				parsed, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return fmt.Errorf("parse int env %s: %w", joinEnvKey(prefix, fieldParents), err)
				}
				field.SetInt(parsed)
			}
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				if value := os.Getenv(joinEnvKey(prefix, fieldParents)); value != "" {
					field.Set(reflect.ValueOf(SplitCSV(value)))
				}
			}
		}
	}

	return nil
}

func yamlTagName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return strings.ToUpper(field.Name)
	}
	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		for _, part := range parts[1:] {
			if part == "inline" {
				return ""
			}
		}
		return strings.ToUpper(field.Name)
	}
	return name
}

func joinEnvKey(prefix string, parts []string) string {
	allParts := append([]string{prefix}, parts...)
	return strings.Join(allParts, "_")
}
