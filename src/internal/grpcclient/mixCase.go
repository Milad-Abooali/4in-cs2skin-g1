package grpcclient

import (
	"google.golang.org/protobuf/types/known/structpb"
	"log"
)

type CaseWithItems map[string]interface{}

func MergeCasesAndItems(
	cases []*structpb.Struct,
	items []*structpb.Struct,
) map[int]CaseWithItems {

	result := make(map[int]CaseWithItems)

	// Make Cases
	for _, c := range cases {
		idField, ok := c.Fields["id"]
		if !ok {
			// log.Printf("[MergeCasesAndItems] case #%d has no 'id' field", i)
			continue
		}

		id := int(idField.GetNumberValue())

		caseMap := make(CaseWithItems)
		for k, v := range c.Fields {
			caseMap[k] = getProtoValue(v)
		}

		caseMap["items"] = make(map[int]map[string]interface{})
		result[id] = caseMap
	}

	// Add Items to Cases
	for _, it := range items {
		if it.Fields["case_id"].GetNumberValue() == 508 {
			log.Println(it)
		}

		caseIDField, ok := it.Fields["case_id"]
		if !ok {
			// log.Printf("[MergeCasesAndItems] item #%d has no 'case_id' field", j)
			continue
		}
		itemIDField, ok := it.Fields["id"]
		if !ok {
			// log.Printf("[MergeCasesAndItems] item #%d has no 'id' field", j)
			continue
		}

		caseID := int(caseIDField.GetNumberValue())
		itemID := int(itemIDField.GetNumberValue())

		caseMap, ok := result[caseID]
		if !ok {
			// log.Printf("[MergeCasesAndItems] item #%d references missing case_id=%d", j, caseID)
			continue
		}

		itemsVal, ok := caseMap["items"]
		if !ok {
			// log.Printf("[MergeCasesAndItems] case_id=%d has no 'items' field", caseID)
			continue
		}

		itemsMap, ok := itemsVal.(map[int]map[string]interface{})
		if !ok {
			// log.Printf("[MergeCasesAndItems] case_id=%d 'items' type assertion failed", caseID)
			continue
		}

		itemMap := make(map[string]interface{})
		for k, v := range it.Fields {
			itemMap[k] = getProtoValue(v)
		}

		itemsMap[itemID] = itemMap
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
		var arr []interface{}
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
