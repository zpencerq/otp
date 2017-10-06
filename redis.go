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
	conn   net.Conn
	reader *bufio.Reader
}

func NewRedisStore(host string) *RedisStore {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		panic(err)
	}

	return &RedisStore{
		conn:   conn,
		reader: bufio.NewReader(conn),
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

	raw, _ := rs.reader.ReadBytes('\n')
	return string(raw[:len(raw)-2])
}

func (rs *RedisStore) Get(key string) *string {
	rs.Lock()
	defer rs.Unlock()

	resp := rs.runCmd("HGET", key, "content")

	n, err := strconv.Atoi(resp[1:])
	if err != nil {
		panic(err)
	}

	bytes := make([]byte, n)
	rs.reader.Read(bytes)
	rs.reader.Discard(2) // throw away next \r\n
	value := string(bytes)

	rs.runCmd("HINCRBY", key, "views", "-1")
	rs.runCmd("HGET", key, "views")

	raw_views, _ := rs.reader.ReadBytes('\n')
	if views, _ := strconv.Atoi(string(raw_views[:len(raw_views)-2])); views <= 0 {
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
