package services

import (
	"testing"

	"github.com/example/table-order/models"
)

func newDeliveryOrder(t *testing.T, quoteNo string) (models.Order, models.OrderDelivery) {
	t.Helper()
	shop := models.Shop{Name: "Dispatch Shop", Latitude: 39.9, Longitude: 116.4, Address: "门店地址"}
	testDB.Create(&shop)
	user := setupUser(t, nil)
	order := models.Order{OrderNo: "DSP" + quoteNo, UserID: user.ID, ShopID: shop.ID, OrderType: "delivery", Amount: 100, Status: 2}
	testDB.Create(&order)
	od := models.OrderDelivery{
		OrderID: order.ID, ShansongQuoteNo: quoteNo, DeliveryFee: 8.5,
		RecipientName: "张三", RecipientPhone: "13800000000",
		Province: "北京市", City: "北京市", County: "朝阳区", DetailAddress: "某路1号",
		RecipientLat: 39.91, RecipientLng: 116.41,
	}
	testDB.Create(&od)
	return order, od
}

func TestDispatchShansong_PersistsOrderNoOnSuccess(t *testing.T) {
	cleanTables(t)
	order, od := newDeliveryOrder(t, "Q-OK")

	prev := Shansong
	Shansong = &ShansongClient{ClientID: "c", AppSecret: "s", BaseURL: "http://stub",
		HTTP: &stubDoer{resp: `{"status":200,"data":{"orderNumber":"SS-ORD-77"}}`}}
	defer func() { Shansong = prev }()

	DispatchShansong(order.ID)

	var got models.OrderDelivery
	testDB.First(&got, od.ID)
	if got.ShansongOrderNo != "SS-ORD-77" {
		t.Errorf("expected shansong_order_no persisted, got %q", got.ShansongOrderNo)
	}
}

func TestDispatchShansong_FailureLeavesOrderNoEmpty(t *testing.T) {
	cleanTables(t)
	order, od := newDeliveryOrder(t, "Q-FAIL")

	prev := Shansong
	Shansong = &ShansongClient{ClientID: "c", AppSecret: "s", BaseURL: "http://stub",
		HTTP: &stubDoer{resp: `{"status":500,"msg":"boom","data":null}`}}
	defer func() { Shansong = prev }()

	DispatchShansong(order.ID) // must not panic

	var got models.OrderDelivery
	testDB.First(&got, od.ID)
	if got.ShansongOrderNo != "" {
		t.Errorf("expected empty shansong_order_no on failure, got %q", got.ShansongOrderNo)
	}
}

// A non-delivery order (no OrderDelivery row) is a safe no-op.
func TestDispatchShansong_NoOpForNonDelivery(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)
	shop := models.Shop{Name: "Dinein"}
	testDB.Create(&shop)
	order := models.Order{OrderNo: "DINE1", UserID: user.ID, ShopID: shop.ID, OrderType: "dine_in", Amount: 50, Status: 2}
	testDB.Create(&order)

	prev := Shansong
	Shansong = &ShansongClient{ClientID: "c", AppSecret: "s", BaseURL: "http://stub",
		HTTP: &stubDoer{resp: `{"status":200,"data":{"orderNumber":"SHOULD-NOT-USE"}}`}}
	defer func() { Shansong = prev }()

	DispatchShansong(order.ID) // no OrderDelivery → returns silently, no panic
}

// Idempotent: a second dispatch (e.g. duplicate pay notify) does not re-create.
func TestDispatchShansong_IdempotentWhenAlreadyDispatched(t *testing.T) {
	cleanTables(t)
	order, od := newDeliveryOrder(t, "Q-IDEM")
	testDB.Model(&od).Update("shansong_order_no", "SS-EXISTING")

	prev := Shansong
	Shansong = &ShansongClient{ClientID: "c", AppSecret: "s", BaseURL: "http://stub",
		HTTP: &stubDoer{resp: `{"status":200,"data":{"orderNumber":"SS-NEW"}}`}}
	defer func() { Shansong = prev }()

	DispatchShansong(order.ID)

	var got models.OrderDelivery
	testDB.First(&got, od.ID)
	if got.ShansongOrderNo != "SS-EXISTING" {
		t.Errorf("expected dispatch to be idempotent, got %q", got.ShansongOrderNo)
	}
}
