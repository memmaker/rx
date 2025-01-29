package ui_graphics

import (
    "contractor/foundation"
    "fmt"
)

func (u *UIController) OpenTransferEnchantmentMenu() {
    // step one, choose an item
    state := u.gameState.GameState()
    player := state.GetPlayer()
    playerInventory := player.GetInventory()
    msg := u.msg
    if playerInventory.IsEmpty() {
        msg(fmt.Sprintf("You have no items"))
        return
    }
    itemsWithEnchantments := make([]foundation.Item, 0)
    for _, item := range playerInventory.Items() {
        gItem := item
        if len(gItem.GetAllEnchantments()) > 0 {
            itemsWithEnchantments = append(itemsWithEnchantments, item)
        }
    }
    if len(itemsWithEnchantments) == 0 {
        msg(fmt.Sprintf("You have no items with enchantments"))
        return
    }
    var menuItems []MenuItem
    for _, item := range itemsWithEnchantments {
        gItem := item
        menuItems = append(menuItems, MenuItem{
            MainText: gItem.Name(),
            Action: func() {
                u.openChooseEnchantmentMenu(gItem)
            },
        })
    }

    u.OpenMenu(menuItems)
}

func (u *UIController) openChooseEnchantmentMenu(source foundation.Item) {
    var menuItems []MenuItem
    for _, e := range source.GetAllEnchantments() {
        enchantment := e
        menuItems = append(menuItems, MenuItem{
            MainText: enchantment.ToString(),
            Action: func() {
                u.openChooseTransferTargetMenu(source, enchantment)
            },
        })
    }

    u.OpenMenu(menuItems)
}

func (u *UIController) openChooseTransferTargetMenu(source foundation.Item, enchantment common.Enchantment) {
    state := u.gameState.GameState()
    player := state.GetPlayer()
    playerInventory := player.GetInventory()
    if playerInventory.IsEmpty() {
        return
    }
    var menuItems []MenuItem
    for _, item := range playerInventory.Items() {
        gItem := item
        if gItem == source {
            continue
        }

        menuItems = append(menuItems, MenuItem{
            MainText: gItem.Name(),
            Action: func() {
                u.transferEnchantment(source, gItem, enchantment)
            },
        })
    }

    u.OpenMenu(menuItems)
}

func (u *UIController) transferEnchantment(source foundation.Item, item foundation.Item, enchantment common.Enchantment) {
    msg := u.msg
    source.RemoveEnchantment(enchantment)
    msg(fmt.Sprintf("%s lost %s", source.Name(), enchantment.ToString()))
    item.AddEnchantment(enchantment)
    msg(fmt.Sprintf("%s gained %s", item.Name(), enchantment.ToString()))
    u.PopAll()
}
