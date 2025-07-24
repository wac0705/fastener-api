// fastener-api/handler/menus.go

package handler

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/wac0705/fastener-api/db"
	"github.com/wac0705/fastener-api/models"
	"github.com/golang-jwt/jwt/v5"
)

// ... (您原本的 CreateMenu, GetMenus, GetMenu, UpdateMenu, DeleteMenu 函式保持不變) ...

// --- 新增的函式 ---

// buildMenuTree 將扁平的選單列表轉換為樹狀結構
func buildMenuTree(menus []models.Menu) []models.Menu {
	menuMap := make(map[uint]models.Menu)
	var rootMenus []models.Menu

	for _, m := range menus {
		menuMap[m.ID] = m
	}

	for _, m := range menus {
		if m.ParentID != nil {
			parent, ok := menuMap[*m.ParentID]
			if ok {
				if parent.Children == nil {
					parent.Children = []models.Menu{}
				}
				parent.Children = append(parent.Children, m)
				menuMap[*m.ParentID] = parent
			}
		} else {
			rootMenus = append(rootMenus, m)
		}
	}
    
    // 為了確保每次遍歷順序一致，需要對 map 的 key 進行排序
    var keys []uint
    for k := range menuMap {
        keys = append(keys, k)
    }
    sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

    var resultMenus []models.Menu
    // 從 rootMenus 開始構建最終結果，確保順序
    for _, root := range rootMenus {
        resultMenus = append(resultMenus, menuMap[root.ID])
    }


	// 對每一層的 children 進行排序
	var sortChildren func(menus []models.Menu)
	sortChildren = func(menus []models.Menu) {
		sort.Slice(menus, func(i, j int) bool {
			return menus[i].OrderNo < menus[j].OrderNo
		})
		for i := range menus {
			if len(menus[i].Children) > 0 {
				sortChildren(menus[i].Children)
			}
		}
	}

	sortChildren(resultMenus)
    sort.Slice(resultMenus, func(i, j int) bool {
        return resultMenus[i].OrderNo < resultMenus[j].OrderNo
    })

	return resultMenus
}


// GetUserMenus 根據 JWT 中的 role_id 獲取使用者可見的選單樹
func GetUserMenus(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	roleID := uint(claims["role_id"].(float64))

	var menus []models.Menu
	// 使用 GORM 進行 Join 查詢
	result := db.DB.
		Joins("JOIN role_menu_relations ON role_menu_relations.menu_id = menus.id").
		Where("role_menu_relations.role_id = ?", roleID).
        Where("menus.is_active = ?", true).
		Order("menus.order_no ASC").
		Find(&menus)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": result.Error.Error()})
	}

	menuTree := buildMenuTree(menus)

	return c.JSON(menuTree)
}

// GetAllMenusTree 獲取完整的選單樹（供後台管理使用）
func GetAllMenusTree(c *fiber.Ctx) error {
    var menus []models.Menu
    result := db.DB.Order("order_no ASC").Find(&menus)
    if result.Error != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": result.Error.Error()})
    }
    menuTree := buildMenuTree(menus)
    return c.JSON(menuTree)
}
