package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func (g *GormRepository) Create(value interface{}) error {
	result := g.db.Create(value)
	return result.Error
}

func (g *GormRepository) Update(value interface{}) error {
	result := g.db.Updates(value)
	return result.Error
}

func (g *GormRepository) Delete(value interface{}) error {
	result := g.db.Delete(value)
	return result.Error
}

func (g *GormRepository) Migrate(value interface{}) error {
	return g.db.AutoMigrate(value)
}

func (g *GormRepository) CleanUp() {
	migrator := g.db.Migrator()
	tables, _ := migrator.GetTables()

	for _, table := range tables {
		migrator.DropTable(table)
	}
}

func CreateGormRepository(driver string, dns string) (Repository, error) {
	db, err := gorm.Open(GetDialector(driver, dns))
	if err != nil {
		return nil, err
	}
	return &GormRepository{
		db: db,
	}, nil
}

func GetDialector(driver string, dns string) gorm.Dialector {
	switch driver {
	case "sqlite":
		return sqlite.Open(dns)
	default:
		return nil
	}
}
