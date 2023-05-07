package util

import (
	companyModel "main-server/pkg/model/company"
	"reflect"
)

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func RemoveDuplicate(sliceList []companyModel.ManagerDataEx) []companyModel.ManagerDataEx {
	allKeys := make(map[string]bool)
	list := []companyModel.ManagerDataEx{}

	for _, item := range sliceList {
		if _, value := allKeys[item.Uuid]; !value {
			allKeys[item.Uuid] = true
			list = append(list, item)
		}
	}

	return list
}
