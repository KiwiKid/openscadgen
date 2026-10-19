package pkg

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

// Report matching repeated values as places to check, rather than assuming
// every repetition in the configuration caused this particular path collision.
func duplicateParamDiagnostics(config *models.Config, first, duplicate models.InstanceConfig) string {
	var diagnostics []string
	inspect := func(location string, params map[string]interface{}, literalKeys []string) {
		keys := make([]string, 0, len(params))
		for key := range params {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if slices.Contains(literalKeys, key) {
				continue
			}
			firstValue, firstOK := first.Params[key]
			duplicateValue, duplicateOK := duplicate.Params[key]
			if !firstOK || !duplicateOK || fmt.Sprint(firstValue) != fmt.Sprint(duplicateValue) {
				continue
			}
			var values []interface{}
			value := reflect.ValueOf(params[key])
			if value.IsValid() && (value.Kind() == reflect.Slice || value.Kind() == reflect.Array) {
				for i := 0; i < value.Len(); i++ {
					values = append(values, value.Index(i).Interface())
				}
			} else if str, ok := params[key].(string); ok && strings.Contains(str, ",") {
				for _, part := range strings.Split(str, ",") {
					values = append(values, strings.TrimSpace(part))
				}
			}
			var positions []string
			for i, candidate := range values {
				if fmt.Sprint(candidate) == fmt.Sprint(firstValue) {
					positions = append(positions, fmt.Sprint(i+1))
				}
			}
			if len(positions) > 1 {
				diagnostics = append(diagnostics, fmt.Sprintf("- %s.%s: value %q is repeated at positions %s.", location, key, fmt.Sprint(firstValue), strings.Join(positions, ", ")))
			}
		}
	}
	// Global lists are expanded independently of instance lists.
	inspect("[openscadgen].global_params", config.Design.GlobalParams, nil)
	for i, instance := range config.Design.ConfiguredInstanceConfig {
		if instance.Name != first.Name && instance.Name != duplicate.Name {
			continue
		}
		location := fmt.Sprintf("[[openscadgen.instances]] #%d", i+1)
		if instance.Name != "" {
			location += fmt.Sprintf(" (name = %q)", instance.Name)
		}
		literalKeys := append([]string{}, instance.DirectArrayParams...)
		for _, key := range instance.IgnoreCommaInParams {
			if _, ok := instance.Params[key].(string); ok {
				literalKeys = append(literalKeys, key)
			}
		}
		inspect(location+".params", instance.Params, literalKeys)
	}
	for i, input := range config.GetInputPaths() {
		if input.Path != first.InputPath.Path && input.Path != duplicate.InputPath.Path {
			continue
		}
		inspect(fmt.Sprintf("[[openscadgen.input_paths]] #%d (path = %q).params", i+1, input.Path), input.Params, nil)
	}
	for i, paramSet := range config.Design.ParamSets {
		used := slices.Contains(strings.Split(first.InputPath.ParamSets, ","), paramSet.Name) || slices.Contains(strings.Split(duplicate.InputPath.ParamSets, ","), paramSet.Name)
		for _, instance := range config.Design.ConfiguredInstanceConfig {
			if (instance.Name == first.Name || instance.Name == duplicate.Name) && slices.Contains(strings.Split(instance.ParamSets, ","), paramSet.Name) {
				used = true
			}
		}
		if used {
			inspect(fmt.Sprintf("[[openscadgen.param_sets]] #%d (name = %q).params", i+1, paramSet.Name), paramSet.Params, nil)
		}
	}
	if len(diagnostics) == 0 {
		return ""
	}
	return "\n\nREPEATED PARAMETER VALUES TO CHECK:\n" + strings.Join(diagnostics, "\n") + "\nRemove repeated values from these lists if they should generate only one instance per value."
}
