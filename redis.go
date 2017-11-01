package otp

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RedisStore struct {
	host string
}

type RedisConn struct {
	sync.Mutex
	net.Conn
	reader *bufio.Reader
}

func NewRedisStore(host string) OneTimeStore {
	conn, err := net.DialTimeout("tcp", host, time.Duration(1)*time.Second)
	if err != nil {
		panic(err)
	}
	conn.Close()

	return &RedisStore{
		host: host,
	}
}

func (rs *RedisStore) NewConn() OneTimeStoreConn {
	conn, err := net.DialTimeout("tcp", rs.host, time.Duration(10)*time.Second)
	if err != nil {
		panic(err)
	}

	return &RedisConn{
		Conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (rc *RedisConn) runCmd(args ...string) string {
	chunks := []string{
		fmt.Sprintf("*%d", len(args)),
	}

	for _, arg := range args {
		chunks = append(chunks,
			fmt.Sprintf("$%d", len(arg)), arg)
	}

	chunks = append(chunks, "")

	result := strings.Join(chunks, "\r\n")

	rc.Conn.Write([]byte(result))

	raw, _ := rc.reader.ReadBytes('\n')
	return string(raw[:len(raw)-2])
}

func (rc *RedisConn) Get(key string) *string {
	rc.Lock()
	defer rc.Unlock()

	resp := rc.runCmd("HGET", key, "content")

	n, err := strconv.Atoi(resp[1:])
	if err != nil {
		panic(err)
	}

	bytes := make([]byte, n)
	rc.reader.Read(bytes)
	rc.reader.Discard(2) // throw away next \r\n
	value := string(bytes)

	rc.runCmd("HINCRBY", key, "views", "-1")
	rc.runCmd("HGET", key, "views")

	raw_views, _ := rc.reader.ReadBytes('\n')
	if views, _ := strconv.Atoi(string(raw_views[:len(raw_views)-2])); views <= 0 {
		rc.runCmd("DEL", key)
	}

	return &value

}

func (rc *RedisConn) Exists(key string) bool {
	rc.Lock()
	defer rc.Unlock()

	response := rc.runCmd("EXISTS", key)
	return response[1] == '1'
}

func (rc *RedisConn) Set(content string, views int, expire int) string {
	rc.Lock()
	defer rc.Unlock()

	key := generateUUID()
	rc.runCmd("HMSET", key, "content", content, "views", strconv.Itoa(views))
	rc.runCmd("EXPIRE", key, strconv.Itoa(expire))

	return key
}

func (rc *RedisConn) Close() error {
	return rc.Conn.Close()
}
