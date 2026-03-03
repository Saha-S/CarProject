package main

func manufacturerByID(id int) *Manufacturer {
	for i := range db.Manufacturers {
		if db.Manufacturers[i].ID == id {
			return &db.Manufacturers[i]
		}
	}
	return nil
}

func categoryByID(id int) *Category {
	for i := range db.Categories {
		if db.Categories[i].ID == id {
			return &db.Categories[i]
		}
	}
	return nil
}

func carByID(id int) *CarModel {
	for i := range db.CarModels {
		if db.CarModels[i].ID == id {
			return &db.CarModels[i]
		}
	}
	return nil
}
