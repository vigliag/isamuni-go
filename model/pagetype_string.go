// Code generated by "stringer -type PageType"; DO NOT EDIT.

package model

import "strconv"

const _PageType_name = "PageUser"

var _PageType_index = [...]uint8{0, 8}

func (i PageType) String() string {
	if i < 0 || i >= PageType(len(_PageType_index)-1) {
		return "PageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PageType_name[_PageType_index[i]:_PageType_index[i+1]]
}