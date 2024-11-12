package storage

import (
	"context"

	"github.com/chazari-x/training-api-v1/model"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *pg.DB
}

func NewStorage(ctx context.Context, Host string, Port string, User string, Pass string, Name string) (*Storage, error) {
	db := pg.Connect(&pg.Options{
		Addr:     Host + ":" + Port,
		User:     User,
		Password: Pass,
		Database: Name,
	})

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	s := &Storage{db: db}

	if err := s.create(db); err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

func (s *Storage) Close() {
	_ = s.db.Close()
}

func (s *Storage) create(db *pg.DB) error {
	models := []interface{}{
		(*model.User)(nil),
	}

	for _, m := range models {
		if err := db.Model(m).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) SelectById(id int) (model.User, error) {
	user := model.User{AccountID: id}
	err := s.db.Model(&user).WherePK().Select()
	return user, err
}

func (s *Storage) SearchByPageAndLimit(search string, limit, offset int, orderBy string) ([]model.ShortUser, error) {
	var users []model.User
	err := s.db.Model(&users).Where("account_name ILIKE ? OR account_id::text ILIKE ?", "%"+search+"%", "%"+search+"%").Limit(limit).Offset(offset).Order(orderBy).Select()
	if err != nil {
		return nil, err
	}

	var shortUsers []model.ShortUser
	for _, user := range users {
		shortUsers = append(shortUsers, model.ShortUser{ID: user.AccountID, Login: user.AccountName})
	}

	return shortUsers, nil
}
