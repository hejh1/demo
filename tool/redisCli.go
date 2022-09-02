package tool

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var REDIS_ADDRESS string = "34.134.141.170:6379"
var REDIS_PEXPIRE time.Duration = time.Hour * 24
var REDIS_PASSWORD string = "cmVkaXNhZG1pbnBhc3N3b3JkCg=="

type RedisCli struct {
	Client   *redis.Client
	ctx      context.Context
	isExists bool
	pexpire  time.Duration // data expire, milliseconds
}

func (rdb *RedisCli) Init(address, password string, db int) {
	rdb.Client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db, // default DB is 0
	})
	rdb.ctx = context.Background()
	rdb.isExists = false
	rdb.pexpire = 0 // data will not expire if set 0
}

func (rdb *RedisCli) InitWithContext(address, password string, db int, ctx context.Context) {
	rdb.Client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db, // default DB is 0
	})
	rdb.ctx = ctx
	rdb.isExists = false
	rdb.pexpire = 0 // data will not expire if set 0
}

func (rdb *RedisCli) SetPexpire(t time.Duration) {
	rdb.pexpire = t
}

func (rdb *RedisCli) GetPage(UUID string, page int) (interface{}, error) {
	val, err := rdb.Client.Get(rdb.ctx, UUID+strconv.Itoa(page)).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (rdb *RedisCli) SavePage(UUID string, page int, data interface{}) error {

	var err error = nil
	var keyPageId string = UUID + strconv.Itoa(page)

	// insert page data
	err = rdb.Client.Set(rdb.ctx, keyPageId, data, rdb.pexpire).Err()
	if err != nil {
		panic(err)
	}

	// insert page index
	member := redis.Z{Score: float64(page), Member: keyPageId}
	if rdb.isExists {
		_, err = rdb.Client.ZAdd(rdb.ctx, UUID, &member).Result()
		if err != nil {
			return err
		}
	} else {
		// check key UUID exists or not
		isExists, err := rdb.Client.Exists(rdb.ctx, UUID).Result()
		if err != nil {
			return err
		}
		if isExists == 0 {
			// insert page index
			_, err = rdb.Client.ZAdd(rdb.ctx, UUID, &member).Result()
			if err != nil {
				return err
			}
		} else {
			// insert page index and set pexpire
			_, err = rdb.Client.ZAdd(rdb.ctx, UUID, &member).Result()
			if err != nil {
				return err
			}
			_, err = rdb.Client.PExpire(rdb.ctx, UUID, rdb.pexpire).Result()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// return max page number
func (rdb *RedisCli) GetMaxPage(UUID string) (int, error) {
	// get pages count
	count, err := rdb.Client.ZCard(rdb.ctx, UUID).Result()
	if err != nil || count == 0 {
		return 0, err
	}
	// get the max page index, the menber scope
	members, err := rdb.Client.ZRangeWithScores(rdb.ctx, UUID, count-1, count).Result()
	if err != nil {
		return 0, err
	}
	return int(members[0].Score), nil
}

// return how many pages
func (rdb *RedisCli) GetPageCount(UUID string) (int, error) {
	count, err := rdb.Client.ZCard(rdb.ctx, UUID).Result()
	if err != nil {
		return 0, err
	}
	return int(count), err
}

func (rdb *RedisCli) Close() error {
	err := rdb.Client.Close()
	if err != nil {
		return err
	}
	return nil
}

func GetPage(UUID string, page int) (interface{}, error) {
	var rdb RedisCli
	rdb.Init(REDIS_ADDRESS, REDIS_PASSWORD, 0)
	defer rdb.Close()
	val, err := rdb.GetPage(UUID, page)
	return val, err
}

func SavePage(UUID string, page int, data interface{}) error {
	var rdb RedisCli
	rdb.Init(REDIS_ADDRESS, REDIS_PASSWORD, 0)
	defer rdb.Close()
	rdb.SetPexpire(REDIS_PEXPIRE)
	err := rdb.SavePage(UUID, page, data)
	return err
}

func GetMaxPage(UUID string) (int, error) {
	var rdb RedisCli
	rdb.Init(REDIS_ADDRESS, REDIS_PASSWORD, 0)
	defer rdb.Close()
	val, err := rdb.GetMaxPage(UUID)
	return val, err
}

func GetPageCount(UUID string) (int, error) {
	var rdb RedisCli
	rdb.Init(REDIS_ADDRESS, REDIS_PASSWORD, 0)
	defer rdb.Close()
	val, err := rdb.GetPageCount(UUID)
	return val, err
}
