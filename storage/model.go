package storage

import (
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

type NewsItem struct {
	gorm.Model
	TimeStamp time.Time `json:"time_stamp" validate:"nonzero,time"`
	Header    string    `json:"header" validate:"nonzero" gorm:"unique"`
}

type NewsDataStore struct {
	Conn *gorm.DB
}

func (d *NewsDataStore) GetByObject(n *NewsItem) error {
	return d.Conn.First(n).Error
}

func (d *NewsDataStore) GetAll() (ni *[]NewsItem, err error) {
	err = d.Conn.Find(ni).Error
	return
}

func (d *NewsDataStore) GetByHeader(h string) (ni *NewsItem, err error) {
	err = d.Conn.Where("lower(header) = ?", strings.ToLower(h)).First(ni).Error
	return
}

func (d *NewsDataStore) GetById(i int) (ni *NewsItem, err error) {
	err = d.Conn.Where("id = ?", i).First(ni).Error
	return
}

func (d *NewsDataStore) GetByTime(begin time.Time, end time.Time) (ni *[]NewsItem, err error) {
	err = d.Conn.Where("time_stamp > ? and time_stamp < ?", begin, end).Find(ni).Error
	return
}

func (d *NewsDataStore) AddNews(ni *NewsItem) error {
	return d.Conn.Create(ni).Error
}

func (d *NewsDataStore) ChangeNews(ni *NewsItem) error {
	return d.Conn.Update(ni).Error
}
