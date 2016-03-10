package history

import (
	"github.com/garyburd/redigo/redis"
)

type HistoryDb struct {
	redis redis.Conn
}

func NewHistoryDb(address string) (*HistoryDb, error) {
	h := &HistoryDb{}
	c, err := redis.Dial("tcp", address)
	if err != nil {
		return h, err
	}
	h.redis = c
	return h, nil
}

func (h *HistoryDb) AddStartHistory(id uint64) error {
	_, err := h.redis.Do("MSET", "start-history", id, "current-history", id)
	if err != nil {
		return err
	}
	return nil
}

func (h *HistoryDb) SetCurrentHistory(id uint64) error {
	_, err := h.redis.Do("SET", "current-history", id)
	if err != nil {
		return err
	}
	return nil
}

func (h *HistoryDb) HasStartHistory() (bool, error) {
	return redis.Bool(h.redis.Do("EXISTS", "start-history"))
}

func (h *HistoryDb) HasCurrentHistory() (bool, error) {
	return redis.Bool(h.redis.Do("EXISTS", "current-history"))
}

func (h *HistoryDb) GetCurrentHistory() (uint64, error) {
	return redis.Uint64(h.redis.Do("GET", "current-history"))
}

func (h *HistoryDb) GetStartHistory() (uint64, error) {
	return redis.Uint64(h.redis.Do("GET", "start-history"))
}
