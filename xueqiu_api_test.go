package xueqiu_api

import (
    "testing"
    "fmt"
)

func TestNew(t *testing.T) {

    ctrl := New("username", "pasword")
    ctrl.Login()

    list := ctrl.GetCodeList()

    fmt.Println(list[0])
    fmt.Println(len(list))

    detailList := ctrl.GetDetail(list[0]);
    fmt.Println(detailList)

}