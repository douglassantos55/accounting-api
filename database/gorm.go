package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

type GormQueryResult struct {
	db *gorm.DB
}

func (g *GormQueryResult) Get(dest interface{}) error {
	result := g.db.Find(dest)
	return result.Error
}

func (g *GormQueryResult) With(relation string) QueryResult {
	g.db = g.db.Preload(relation)
	return g
}

func (g *GormQueryResult) Where(column string, value interface{}) QueryResult {
	g.db = g.db.Where(map[string]interface{}{
		column: value,
	})
	return g
}

func (g *GormRepository) Find(model interface{}) QueryResult {
	return &GormQueryResult{
		db: g.db.Model(model),
	}
}

func (g *GormRepository) Create(model interface{}) error {
	result := g.db.Create(model)
	return result.Error
}

func (g *GormRepository) Update(model interface{}) error {
	result := g.db.Updates(model)
	return result.Error
}

func (g *GormRepository) Delete(model interface{}) error {
	result := g.db.Delete(model)
	return result.Error
}

func (g *GormRepository) Migrate(model interface{}) error {
	return g.db.AutoMigrate(model)
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
