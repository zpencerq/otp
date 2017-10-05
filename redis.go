package otp

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type RedisStore struct {
	sync.Mutex
	conn    net.Conn
	scanner *bufio.Scanner
}

func NewRedisStore(host string) *RedisStore {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		panic(err)
	}

	return &RedisStore{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}
}

func (rs *RedisStore) runCmd(args ...string) string {
	chunks := []string{
		fmt.Sprintf("*%d", len(args)),
	}

	for _, arg := range args {
		chunks = append(chunks,
			fmt.Sprintf("$%d", len(arg)), arg)
	}

	chunks = append(chunks, "")

	result := strings.Join(chunks, "\r\n")

	rs.conn.Write([]byte(result))

	rs.scanner.Scan()
	return rs.scanner.Text()
}

func (rs *RedisStore) Get(key string) *string {
	rs.Lock()
	defer rs.Unlock()

	rs.runCmd("HGET", key, "content")

	rs.scanner.Scan()
	value := rs.scanner.Text()

	rs.runCmd("HINCRBY", key, "views", "-1")
	rs.runCmd("HGET", key, "views")

	rs.scanner.Scan()
	raw_views := rs.scanner.Text()

	if views, _ := strconv.Atoi(strings.TrimSpace(string(raw_views))); views <= 0 {
		rs.runCmd("DEL", key)
	}

	return &value

}

func (rs *RedisStore) Exists(key string) bool {
	rs.Lock()
	defer rs.Unlock()

	response := rs.runCmd("EXISTS", key)
	return response[1] == '1'
}

func (rs *RedisStore) Set(content string, views int, expire int) string {
	rs.Lock()
	defer rs.Unlock()

	key := generateUUID()
	rs.runCmd("HMSET", key, "content", content, "views", strconv.Itoa(views))
	rs.runCmd("EXPIRE", key, strconv.Itoa(expire))

	return key
}
