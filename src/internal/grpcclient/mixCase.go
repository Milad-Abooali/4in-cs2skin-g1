package grpcclient

import (
	"google.golang.org/protobuf/types/known/structpb"
)

type CaseWithItems map[string]interface{}

func MergeCasesAndItems(cases []*structpb.Struct, items []*structpb.Struct) map[int]CaseWithItems {
	result := make(map[int]CaseWithItems)

	for _, c := range cases {
		id := int(c.Fields["id"].GetNumberValue())

		caseMap := make(CaseWithItems)
		for k, v := range c.Fields {
			caseMap[k] = getProtoValue(v)
		}

		caseMap["items"] = make(map[int]map[string]interface{})

		result[id] = caseMap
	}

	for _, it := range items {
		caseID := int(it.Fields["case_id"].GetNumberValue())
		itemID := int(it.Fields["id"].GetNumberValue())

		if caseMap, ok := result[caseID]; ok {
			itemsMap := caseMap["items"].(map[int]map[string]interface{})

			itemMap := make(map[string]interface{})
			for k, v := range it.Fields {
				itemMap[k] = getProtoValue(v)
			}

			itemsMap[itemID] = itemMap
		}
	}

	return result
}

func getProtoValue(v *structpb.Value) interface{} {
	switch kind := v.Kind.(type) {
	case *structpb.Value_StringValue:
		return kind.StringValue
	case *structpb.Value_NumberValue:
		return kind.NumberValue
	case *structpb.Value_BoolValue:
		return kind.BoolValue
	case *structpb.Value_StructValue:
		m := make(map[string]interface{})
		for k, val := range kind.StructValue.Fields {
			m[k] = getProtoValue(val)
		}
		return m
	case *structpb.Value_ListValue:
		arr := []interface{}{}
		for _, val := range kind.ListValue.Values {
			arr = append(arr, getProtoValue(val))
		}
		return arr
	default:
		return nil
	}
}

func ListValueToStructs(list *structpb.ListValue) []*structpb.Struct {
	var out []*structpb.Struct
	for _, v := range list.Values {
		if s := v.GetStructValue(); s != nil {
			out = append(out, s)
		}
	}
	return out
}
