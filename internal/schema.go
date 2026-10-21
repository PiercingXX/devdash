package internal

import (
	"encoding/json"
	"slices"
	"strings"
)

// SchemaURL where the JSON schema of the dashboards is published.
const SchemaURL = "https://raw.githubusercontent.com/PiercingXX/devdash/master/doc/devdash.schema.json"

type object = map[string]any

// optionSchema of the value of an option.
func optionSchema(name string) object {
	numbers := object{"anyOf": []object{{"type": "integer", "minimum": 0}, {"type": "string", "pattern": "^[0-9]+$"}}}
	booleans := object{"anyOf": []object{{"type": "boolean"}, {"enum": []string{"true", "false"}}}}

	switch {
	case name == optionColor || strings.HasSuffix(name, "_color"):
		return object{"$ref": "#/definitions/color"}
	case slices.Contains([]string{optionHeight, optionBarGap, optionBarWidth, optionRowLimit, optionCharLimit}, name):
		return numbers
	case slices.Contains([]string{optionBold, optionMultiline, optionGlobal, optionRate}, name):
		return booleans
	case name == optionTimePeriod:
		return object{"enum": []string{"day", "month", "year"}}
	case name == optionScope:
		return object{"enum": []string{"owner", "all"}}
	case name == optionUnit:
		return object{"enum": units}
	}

	return object{"type": []string{"string", "number", "boolean"}}
}

// optionsSchema of an object with the options, defined once in the definitions.
func optionsSchema(names []string) object {
	props := object{}
	for _, n := range names {
		props[n] = object{"$ref": "#/definitions/option_" + n}
	}

	return object{"type": "object", "additionalProperties": false, "properties": props}
}

func stringsObject(keys []string) object {
	props := object{}
	for _, k := range keys {
		props[k] = object{"type": []string{"string", "number"}}
	}

	return object{"type": "object", "additionalProperties": false, "properties": props}
}

// Schema of the dashboards, in JSON.
func Schema() ([]byte, error) {
	services := object{}
	for name, keys := range serviceKeys {
		services[name] = stringsObject(keys)
	}

	nameOptions := optionsSchema(nameOptionsKeys)

	conditions := []object{}
	for _, n := range WidgetNames() {
		o, _ := WidgetOptions(n)
		conditions = append(conditions, object{
			"if":   object{"properties": object{"name": object{"const": n}}},
			"then": object{"properties": object{"options": optionsSchema(o)}},
		})
	}

	keys := object{}
	for _, k := range keyNames {
		keys[k] = object{"type": "string"}
	}

	colors := sortedKeys(colorLookUp)
	definitions := object{}
	for _, o := range slices.Concat(AllOptions(), nameOptionsKeys) {
		definitions["option_"+o] = optionSchema(o)
	}
	definitions["option_"+optionSize] = object{"$ref": "#/definitions/size"}

	schema := object{
		"$schema":              "http://json-schema.org/draft-07/schema#",
		"$id":                  SchemaURL,
		"title":                "DevDash dashboard",
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"projects"},
		"properties": object{
			"general": object{
				"type":                 "object",
				"additionalProperties": false,
				"properties": object{
					"refresh": object{"type": "integer", "minimum": 1, "description": "Seconds between two refreshes"},
					"editor":  object{"type": "string"},
					"keys":    object{"type": "object", "additionalProperties": false, "properties": keys},
				},
			},
			"projects": object{"type": "array", "minItems": 1, "items": object{"$ref": "#/definitions/project"}},
		},
		"definitions": object{
			"color": object{"enum": colors},
			"size": object{"anyOf": []object{
				{"type": "integer", "minimum": 1, "maximum": 12},
				{"type": "string", "pattern": "^([1-9]|1[0-2]|[Xx]{0,2}[SsLl]|[Mm])$"},
			}},
			"project": object{
				"type":                 "object",
				"additionalProperties": false,
				"properties": object{
					"name":         object{"type": "string"},
					"name_options": nameOptions,
					"services":     object{"type": "object", "additionalProperties": false, "properties": services},
					"themes": object{
						"type":                 "object",
						"additionalProperties": optionsSchema(AllOptions()),
					},
					"widgets": object{"type": "array", "items": object{"$ref": "#/definitions/row"}},
				},
			},
			"row": object{
				"type":                 "object",
				"additionalProperties": false,
				"properties": object{
					"row": object{"type": "array", "items": object{
						"type":                 "object",
						"additionalProperties": false,
						"properties": object{"col": object{"anyOf": []object{
							{"$ref": "#/definitions/column"},
							{"type": "array", "items": object{"$ref": "#/definitions/column"}},
						}}},
					}},
				},
			},
			"column": object{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"size"},
				"properties": object{
					"size":     object{"$ref": "#/definitions/size"},
					"elements": object{"type": "array", "items": object{"$ref": "#/definitions/widget"}},
				},
			},
			"widget": object{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"name"},
				"properties": object{
					"name":    object{"enum": WidgetNames()},
					"theme":   object{"type": "string"},
					"options": object{"type": "object"},
				},
				"allOf": conditions,
			},
		},
	}

	for k, v := range definitions {
		schema["definitions"].(object)[k] = v
	}

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	return append(b, '\n'), nil
}
