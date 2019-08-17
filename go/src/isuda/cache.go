package main

import (
	// "encoding/json"
	"fmt"
	// "strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

const (
	redisHost = "192.168.0.115"
	redisPort = "6379"

	key          = "KEY"
	REPLACER_KEY = "REP_KEY"
)

// func main() {}

// ===============================
//  既存のDBのすべてのキーを削除
//  必ず成功する
// ===============================

func FLUSH_ALL() error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Do("FLUSHALL")
	return nil
}

// ==============================
//
//  構造体を格納しやすくするため
//  			の関数
//
// ==============================

func getDataFromCache(key string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func setDataToCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("SET", key, data)
	if err != nil {
		return err
	}
	return nil
}

func incrementDataInCache(key string) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("INCR", key)
	if err != nil {
		return err
	}
	return nil
}

// =================================
//
//  構造体をKey: Valueの組み合わせで
//   格納しやすくするための関数
//
// =================================

// fieldが存在する場合更新
func setHashDataToCache(key, field string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("HSET", key, field, data)
	if err != nil {
		return err
	}
	return nil
}

func getHashDataFromCache(key, field string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("HGET", key, field))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func removeHashDataFromCache(key, field string) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("HDEL", key, field)
	if err != nil {
		return err
	}
	return nil
}

// 入力された順
func getAllHashDataFromCache(key string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	strs, err := redis.Strings(conn.Do("HVALS", key))
	if err != nil {
		return nil, err
	}
	str := strings.Join(strs[:], ",")
	str = "[" + str + "]"

	return []byte(str), err
}

// 複数のfieldsを一度にとってくる
func getMultiDataFromCache(key string, fields []string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// conn.Doの引数に合うように変換
	querys := make([]interface{}, 0, len(fields)+1)
	querys = append(querys, key)
	for i := range fields {
		querys = append(querys, fields[i])
	}

	fmt.Println(querys...)
	strs, err := redis.Strings((conn.Do("HMGET", querys...)))
	if err != nil {
		return nil, err
	}

	str := strings.Join(strs[:], ",")
	str = "[" + str + "]"

	return []byte(str), nil
}

// =================================
//
//   構造体をListで
//   格納しやすくするための関数
//
// =================================

func getListDataFromCache(key string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	strs, err := redis.Strings(conn.Do("LRANGE", key, 0, -1))
	if err != nil {
		return nil, err
	}
	str := strings.Join(strs[:], ",")
	str = "[" + str + "]"

	return []byte(str), err
}

// RPUSHは最後に追加
func pushListDataToCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("RPUSH", key, data)
	if err != nil {
		return err
	}
	return nil
}

func pushListDataToCacheWithConnection(key string, data []byte, conn redis.Conn) error {
	_, err := conn.Do("RPUSH", key, data)
	if err != nil {
		return err
	}
	return nil
}

// マッチするものを1つ削除
func removeListDataFromCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("LREM", key, 1, data)
	if err != nil {
		return err
	}
	return nil
}

func removeListDataFromCacheWithConnection(key string, data []byte, conn redis.Conn) error {
	_, err := conn.Do("LREM", key, data)
	if err != nil {
		return err
	}
	return nil
}

func getSetDataFromCache(key string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	strs, err := redis.Strings(conn.Do("SINTER", key))
	if err != nil {
		return nil, err
	}
	str := strings.Join(strs[:], ",")
	str = "[" + str + "]"

	return []byte(str), err
}

func pushSetDataToCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("SADD", key, data)
	if err != nil {
		return err
	}
	return nil
}

// マッチするものを1つ削除
func removeSetDataFromCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("SREM", key, data)
	if err != nil {
		return err
	}
	return nil
}

// ========================================
//          ソート済みセット
// ========================================

func getSortedSetDataFromCache(key string, desc bool) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	var strs []string

	if desc {
		strs, err = redis.Strings(conn.Do("ZRANGE", key, 0, -1))
	} else {
		strs, err = redis.Strings(conn.Do("ZREVRANGE", key, 0, -1))
	}
	if err != nil {
		return nil, err
	}
	str := strings.Join(strs[:], ",")
	str = "[" + str + "]"

	return []byte(str), err
}

func pushSortedSetDataToCache(key string, score int64, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	_, err = conn.Do("ZADD", key, score, data)
	if err != nil {
		return err
	}
	return nil
}

// マッチするものを1つ削除
func removeSortedSetDataFromCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	_, err = conn.Do("ZREM", key, data)
	if err != nil {
		return err
	}
	return nil
}

func pushSortedSetDataToCacheWithConnection(key string, score int64, data []byte, conn redis.Conn) error {
	_, err := conn.Do("ZADD", key, score, data)
	if err != nil {
		return err
	}
	return nil
}
