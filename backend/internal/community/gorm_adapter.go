package community

import "gorm.io/gorm"

// GormAdapter adapts *gorm.DB to implement the Database interface
type GormAdapter struct {
	db *gorm.DB
}

// NewGormAdapter creates a new GORM adapter
func NewGormAdapter(db *gorm.DB) Database {
	return &GormAdapter{db: db}
}

func (g *GormAdapter) Create(value interface{}) *gorm.DB {
	return g.db.Create(value)
}

func (g *GormAdapter) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return g.db.Find(dest, conds...)
}

func (g *GormAdapter) Where(query interface{}, args ...interface{}) Database {
	return &GormAdapter{db: g.db.Where(query, args...)}
}

func (g *GormAdapter) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return g.db.First(dest, conds...)
}

func (g *GormAdapter) Save(value interface{}) *gorm.DB {
	return g.db.Save(value)
}

func (g *GormAdapter) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	return g.db.Delete(value, conds...)
}

func (g *GormAdapter) Order(value interface{}) Database {
	return &GormAdapter{db: g.db.Order(value)}
}

func (g *GormAdapter) Limit(limit int) Database {
	return &GormAdapter{db: g.db.Limit(limit)}
}

func (g *GormAdapter) Offset(offset int) Database {
	return &GormAdapter{db: g.db.Offset(offset)}
}
