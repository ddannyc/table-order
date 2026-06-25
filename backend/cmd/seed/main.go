// Command seed populates the local dev database with a test merchant, shop,
// and a menu (mirroring ui-add-food.png) for previewing the mini-program UI.
//
// Usage (from backend/):  go run ./cmd/seed
// DSN: DATABASE_URL env, else localhost:5432 postgres/postgres db=table_order.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type specSeed struct {
	Name  string
	Price float64
}

type prodSeed struct {
	Name     string
	Category string
	Desc     string
	Price    float64
	Specs    []specSeed
}

// Menu mirrors ui-add-food.png: spec products show 选规格, no-spec products use +/-.
var menu = []prodSeed{
	{"酸奶青提", "新品上市", "香甜青提 + 醇厚酸奶", 15, []specSeed{{"600ml", 15}, {"800ml", 18}}},
	{"酸奶芒果", "新品上市", "新鲜芒果 + 酸奶", 15, []specSeed{{"600ml", 15}, {"800ml", 18}}},
	{"酸奶草莓", "新品上市", "当季草莓 + 酸奶", 18, []specSeed{{"600ml", 16}, {"800ml", 18}}},
	{"蜜汁桃桃柠檬珍珠茶", "新品上市", "蜜桃 + 柠檬 + 珍珠", 18, []specSeed{{"600ml", 18}}},
	{"杨枝甘露奶芙", "新品上市", "杨枝甘露 + 奶芙", 19, []specSeed{{"700ml", 19}}},
	{"手打柠檬茶", "新品上市", "手打鲜柠檬", 12, nil},
	{"芝士葡萄", "芝士奶盖", "芝士奶盖 + 葡萄", 16, []specSeed{{"中杯", 16}, {"大杯", 19}}},
	{"芝士奶绿", "芝士奶盖", "芝士奶盖 + 奶绿", 15, []specSeed{{"中杯", 15}, {"大杯", 18}}},
	{"招牌奶茶", "奶茶牛乳", "经典招牌奶茶", 13, []specSeed{{"中杯", 13}, {"大杯", 16}}},
	{"厚乳波波", "奶茶牛乳", "厚乳 + 黑糖波波", 14, nil},
	{"西柚气泡水", "气泡水", "西柚 + 气泡", 14, nil},
	{"青提气泡水", "气泡水", "青提 + 气泡", 14, nil},
}

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "host=localhost port=5432 user=postgres password=postgres dbname=table_order sslmode=disable"
}

func main() {
	db, err := gorm.Open(postgres.Open(dsn()), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	config.DB = db
	if err := config.MigrateDB(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Merchant (idempotent by phone). Password: 123456
	var merchant models.Merchant
	if err := db.Where("phone = ?", "13800000000").First(&merchant).Error; err != nil {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		merchant = models.Merchant{Phone: "13800000000", Password: string(hashed), Name: "测试商户", Company: "测试餐饮", Status: 1}
		if err := db.Create(&merchant).Error; err != nil {
			log.Fatalf("create merchant: %v", err)
		}
	}

	// Shop (idempotent by merchant + name)
	const shopName = "世纪广场店1楼111号"
	var shop models.Shop
	if err := db.Where("merchant_id = ? AND name = ?", merchant.ID, shopName).First(&shop).Error; err != nil {
		shop = models.Shop{
			MerchantID: merchant.ID, Name: shopName, Description: "测试奶茶店",
			Address: "世纪广场1楼111号", Phone: "021-12345678", Hours: "10:00-22:00",
			Latitude: 31.230416, Longitude: 121.473701, Status: 1,
		}
		if err := db.Create(&shop).Error; err != nil {
			log.Fatalf("create shop: %v", err)
		}
	}

	// Re-seed this shop's menu: drop existing products + their specs, then insert fresh.
	var oldIDs []uint
	db.Model(&models.Product{}).Where("shop_id = ?", shop.ID).Pluck("id", &oldIDs)
	if len(oldIDs) > 0 {
		db.Where("product_id IN ?", oldIDs).Delete(&models.ProductSpec{})
		db.Where("shop_id = ?", shop.ID).Delete(&models.Product{})
	}

	nProducts, nSpecs := 0, 0
	for _, p := range menu {
		product := models.Product{
			ShopID: shop.ID, Name: p.Name, Price: p.Price,
			Description: p.Desc, Category: p.Category, Status: 1,
		}
		if err := db.Create(&product).Error; err != nil {
			log.Fatalf("create product %s: %v", p.Name, err)
		}
		nProducts++
		for _, s := range p.Specs {
			if err := db.Create(&models.ProductSpec{ProductID: product.ID, Name: s.Name, Price: s.Price, Status: 1}).Error; err != nil {
				log.Fatalf("create spec %s/%s: %v", p.Name, s.Name, err)
			}
			nSpecs++
		}
	}

	fmt.Println("✅ Seed done")
	fmt.Printf("   merchant: id=%d phone=13800000000 password=123456\n", merchant.ID)
	fmt.Printf("   shop:     id=%d  %s\n", shop.ID, shop.Name)
	fmt.Printf("   products: %d  specs: %d\n", nProducts, nSpecs)
	fmt.Printf("\n预览（微信开发者工具）:\n")
	fmt.Printf("   堂食菜单: 编译页面 pages/menu/index  参数  shop_id=%d&table_no=A01\n", shop.ID)
	fmt.Printf("   或在 Console: wx.setStorageSync('current_shop_id',%d); wx.setStorageSync('current_table_no','A01')\n", shop.ID)
}
