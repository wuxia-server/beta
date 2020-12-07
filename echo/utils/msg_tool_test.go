package utils

import "testing"

func TestRetMsg_PackError(t *testing.T) {
	ret := NewRetMsg(nil)
	ret.PackError(0, "test", 0, map[string]interface{}{"ok": 1}, []int{1, 3, 4})
	t.Logf("%+v", ret)

}

func TestRetMsg_PackResult(t *testing.T) {
	ret := NewRetMsg(nil)
	//ret.PackResult(map[string]interface{}{"ok": 1})
	t.Logf("%+v", ret)
}
