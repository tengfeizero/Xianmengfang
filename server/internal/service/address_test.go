package service

import (
	"errors"
	"fmt"
	"testing"

	"xmf/internal/model"
)

func addrSvc(t *testing.T) *AddressService {
	t.Helper()
	return &AddressService{DB: newTestDB(t)}
}

func validInput(name string) AddressInput {
	return AddressInput{Name: name, Phone: "13800001111", Province: "广东省", City: "深圳市", District: "南山区", Detail: "科技园 1 号楼"}
}

func TestAddressFirstIsDefault(t *testing.T) {
	svc := addrSvc(t)
	a1, err := svc.Create(1, validInput("张三"))
	if err != nil {
		t.Fatalf("创建地址失败: %v", err)
	}
	if a1.IsDefault != 1 {
		t.Fatalf("第一个地址应自动成为默认，got %d", a1.IsDefault)
	}
	a2, _ := svc.Create(1, validInput("李四"))
	if a2.IsDefault != 0 {
		t.Fatalf("第二个地址不应默认，got %d", a2.IsDefault)
	}
}

func TestAddressDefaultUnique(t *testing.T) {
	svc := addrSvc(t)
	a1, _ := svc.Create(1, validInput("张三"))
	a2, _ := svc.Create(1, validInput("李四"))
	if err := svc.SetDefault(1, a2.ID); err != nil {
		t.Fatalf("设默认失败: %v", err)
	}
	list, _ := svc.List(1)
	defaults := 0
	for _, a := range list {
		if a.IsDefault == 1 {
			defaults++
		}
	}
	if defaults != 1 || list[0].ID != a2.ID {
		t.Fatalf("默认地址应唯一且为 a2，defaults=%d first=%d", defaults, list[0].ID)
	}
	// a1 不再是默认
	var check model.Address
	if err := svc.DB.First(&check, a1.ID).Error; err != nil || check.IsDefault != 0 {
		t.Fatalf("旧默认应被清掉，got %d (err=%v)", check.IsDefault, err)
	}
}

func TestAddressForbiddenAndNotFound(t *testing.T) {
	svc := addrSvc(t)
	a, _ := svc.Create(1, validInput("张三"))
	if _, err := svc.GetForOrder(2, a.ID); !errors.Is(err, ErrAddressForbidden) {
		t.Fatalf("他人地址应拒绝，got %v", err)
	}
	if _, err := svc.GetForOrder(1, 999); !errors.Is(err, ErrAddressNotFound) {
		t.Fatalf("不存在地址应 NotFound，got %v", err)
	}
}

func TestAddressDeletePromotesNewest(t *testing.T) {
	svc := addrSvc(t)
	a1, _ := svc.Create(1, validInput("张三")) // 默认
	a2, _ := svc.Create(1, validInput("李四"))
	if err := svc.Delete(1, a1.ID); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	list, _ := svc.List(1)
	if len(list) != 1 || list[0].ID != a2.ID || list[0].IsDefault != 1 {
		t.Fatalf("删除默认后应剩 a2 且自动成为默认，got %+v", list)
	}
}

func TestAddressLimit(t *testing.T) {
	svc := addrSvc(t)
	for i := 0; i < maxAddressesPerUser; i++ {
		in := validInput("用户")
		in.Phone = fmt.Sprintf("1380000%04d", i) // 11 位合法手机号
		if _, err := svc.Create(1, in); err != nil {
			t.Fatalf("第 %d 个地址不应失败: %v", i+1, err)
		}
	}
	if _, err := svc.Create(1, validInput("超额")); !errors.Is(err, ErrAddressLimit) {
		t.Fatalf("超过 20 条应拒绝，got %v", err)
	}
}

func TestAddressInvalidInput(t *testing.T) {
	svc := addrSvc(t)
	in := validInput("张三")
	in.Phone = "12345" // 非法手机号
	if _, err := svc.Create(1, in); err == nil {
		t.Fatal("非法手机号应报错")
	}
	in2 := validInput("张三")
	in2.Detail = ""
	if _, err := svc.Create(1, in2); !errors.Is(err, ErrAddressInvalid) {
		t.Fatalf("缺详址应 ErrAddressInvalid，got %v", err)
	}
}

func TestOrderSnapshotAddress(t *testing.T) {
	db := newTestDB(t)
	osvc := &OrderService{DB: db}
	asvc := &AddressService{DB: db}
	a, _ := asvc.Create(1, validInput("王五"))
	res, err := osvc.Create(1, []ItemInput{{SKUID: 1, Qty: 1}}, a)
	if err != nil {
		t.Fatalf("下单失败: %v", err)
	}
	o := res.Order
	if o.ReceiverName != "王五" || o.ReceiverPhone != "13800001111" {
		t.Fatalf("快照姓名电话不符: %+v", o)
	}
	want := "广东省深圳市南山区科技园 1 号楼"
	if o.ReceiverAddress != want {
		t.Fatalf("快照地址不符: got %q want %q", o.ReceiverAddress, want)
	}
	// 地址后续修改不影响已下单的快照
	upd := validInput("王五改")
	upd.Phone = "13900002222"
	if err := asvc.Update(1, a.ID, upd); err != nil {
		t.Fatalf("改地址失败: %v", err)
	}
	var check model.Order
	osvc.DB.First(&check, o.ID)
	if check.ReceiverPhone != "13800001111" {
		t.Fatalf("历史订单快照不应被地址修改影响，got %s", check.ReceiverPhone)
	}
}
