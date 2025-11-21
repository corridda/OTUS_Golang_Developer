package menu

type Menu struct {
	add    string
	view   string
	update string
	delete string
}

func NewMenu() Menu {
	return Menu{
		add:    "Добавить",
		view:   "Отобразить",
		update: "Обновить",
		delete: "Удалить",
	}
}
