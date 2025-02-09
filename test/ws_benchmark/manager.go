package main

import (
	"encoding/json"
	"fmt"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	httpURL       = "http://localhost:8080"
	websocketAddr = "ws://localhost:8081"
	contentType   = "application/x-www-form-urlencoded"
)

var (
	msgTimer             *timer
	isStart              bool
	receivedMessageCount int64
)

type Manager struct {
	clients     sync.Map // 用了一个并发安全的map
	PhoneNum    int64    // 起始电话号码
	OnlineCount int64    // 在线成员数量
	SendCount   int64    // 每次发送消息数量
	MemberCount int64    // 群成员总数
	TestCount   int64    // 发送消息次数
	chatId      int64    // 群聊id
	tokens      []string // 分配给 client 进行 Login 的 token
	userIds     []string
}

func NewManager(phoneNum, onlineCount, sendCount, memberCount, testCount int64) *Manager {
	return &Manager{
		PhoneNum:    phoneNum,
		OnlineCount: onlineCount,
		SendCount:   sendCount,
		MemberCount: memberCount,
		TestCount:   testCount,
	}
}

// 这个函数用于测试单机最大连接数
func (m *Manager) RunMaxOnlineCount() {
	// 启动计时器，每隔一段时间打印在线人数和接收消息总数
	m.debug()

	// 批量注册 MemberCount 个用户
	m.registerMember()
	fmt.Println("创建用户完成..")

	// 新建 websocket 连接
	m.batchCreate()
}

// 用于压测群聊
func (m *Manager) Run() {
	// 启动计时器，每隔一段时间打印在线人数和接收消息总数
	m.debug()

	// 批量注册 MemberCount 个用户，并创建群组
	m.registerAndCreateGroup()
	fmt.Println("创建群组完成..")

	// 新建 websocket 连接
	m.batchCreate()

	// 循环发送消息
	m.loopSend()
}

// 批量注册用户
func (m *Manager) registerMember() {
	var (
		start = m.PhoneNum
		end   = start + m.MemberCount
	)
	type respStrut struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token string `json:"token"`
			Id    string `json:"id"`
		} `json:"data"`
	}

	var mutex sync.Mutex

	// 批量注册用户
	var wg sync.WaitGroup
	ch := make(chan struct{}, 10) // 限制并发数
	for i := start; i < end; i++ {
		ch <- struct{}{}
		wg.Add(1)

		go func(i int64) {
			defer func() {
				<-ch
				wg.Done()
			}()
			client := &http.Client{}

			// 准备要发送的表单数据
			data := url.Values{}
			data.Set("phone_number", utils.Int64ToStr(i))
			data.Set("nickname", "test")
			data.Set("password", "123")

			// 创建一个 POST 请求
			req, err := http.NewRequest("POST", httpURL+"/register", strings.NewReader(data.Encode()))
			if err != nil {
				panic(err)
			}
			req.Header.Set("Content-Type", contentType)
			// 发送请求并获取响应
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			// 读取数据，并解析返回值
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var respData respStrut
			err = json.Unmarshal(responseBody, &respData)
			if err != nil {
				panic(err)
			}

			mutex.Lock()
			m.userIds = append(m.userIds, respData.Data.Id)
			m.tokens = append(m.tokens, respData.Data.Token)
			mutex.Unlock()
		}(i)
	}
	close(ch)
	wg.Wait()

	if int64(len(m.tokens)) != m.MemberCount {
		panic("用户注册失败")
	}

	return
}

// 批量注册用户并创建群聊
func (m *Manager) registerAndCreateGroup() {
	var (
		start = m.PhoneNum
		end   = start + m.MemberCount
	)
	type respStrut struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token string `json:"token"`
			Id    string `json:"id"`
		} `json:"data"`
	}

	var mutex sync.Mutex

	// 批量注册用户
	var wg sync.WaitGroup
	ch := make(chan struct{}, 10) // 限制并发数
	for i := start; i < end; i++ {
		ch <- struct{}{}
		wg.Add(1)

		go func(i int64) {
			defer func() {
				<-ch
				wg.Done()
			}()
			client := &http.Client{}

			// 准备要发送的表单数据
			data := url.Values{}
			data.Set("phone_number", utils.Int64ToStr(i))
			data.Set("nickname", "test")
			data.Set("password", "123")

			// 创建一个 POST 请求
			req, err := http.NewRequest("POST", httpURL+"/register", strings.NewReader(data.Encode()))
			if err != nil {
				panic(err)
			}
			req.Header.Set("Content-Type", contentType)
			// 发送请求并获取响应
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			// 读取数据，并解析返回值
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var respData respStrut
			err = json.Unmarshal(responseBody, &respData)
			if err != nil {
				panic(err)
			}

			mutex.Lock()
			m.userIds = append(m.userIds, respData.Data.Id)
			m.tokens = append(m.tokens, respData.Data.Token)
			mutex.Unlock()
		}(i)
	}
	close(ch)
	wg.Wait()

	if int64(len(m.tokens)) != m.MemberCount {
		panic("用户注册失败")
	}

	// 创建群聊
	// 创建一个 http.Client
	client := &http.Client{}

	// 准备要发送的表单数据
	data := url.Values{}
	data.Set("name", "benchmark_test")
	for _, id := range m.userIds {
		data.Add("ids", id)
	}

	// 创建一个 POST 请求
	req, err := http.NewRequest("POST", httpURL+"/group/create", strings.NewReader(data.Encode()))
	if err != nil {
		// 处理错误
		return
	}
	req.Header.Set("Content-Type", contentType)
	// 设置 token
	req.Header.Set("token", m.tokens[0])

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var respData respStrut
	err = json.Unmarshal(responseBody, &respData)
	m.chatId = utils.StrToInt64(respData.Data.Id)
	return
}

// 创建在线成员数量的 websocket 连接
func (m *Manager) batchCreate() {
	var wg sync.WaitGroup
	ch := make(chan struct{}, 100)
	for i := 0; i < int(m.OnlineCount); i++ {
		wg.Add(1)
		ch <- struct{}{}
		go func(i int) {
			defer func() {
				wg.Done()
				<-ch
			}()
			userId := m.userIds[i]
			token := m.tokens[i]

			// client := NewClient(userId, token, websocketAddr)
			client := NewClientInPB(userId, token, websocketAddr)
			if client.conn != nil {
				m.clients.Store(userId, client)
			}
		}(i)
	}
	close(ch)
	wg.Wait()
}

// 每隔 5s，打印一次
func (m *Manager) debug() {
	go func() {
		allTicker := time.NewTicker(time.Second * 10)
		defer allTicker.Stop()
		for {
			select {
			case <-allTicker.C:
				fmt.Println("在线人数:", m.clientCount(), " 接收消息数量:", receivedMessageCount)
			}
		}
	}()
}

// 客户端总数
func (m *Manager) clientCount() int {
	j := 0
	m.clients.Range(func(k, v interface{}) bool {
		j++
		return true
	})
	return j
}

// 每秒循环发送消息
func (m *Manager) loopSend() {
	var (
		count  int64
		ticker = time.NewTicker(time.Second)
		start  int64
		end    = start + m.SendCount
	)
	defer func() {
		ticker.Stop()
		m.clients.Range(func(k, v interface{}) bool {
			client, ok := v.(*Client)
			if ok {
				return true
			}
			// 主动断开
			client.conn.Close()
			return true
		})
	}()
	for {
		select {
		case <-ticker.C:
			// 首次发送消息开始计时
			if !isStart {
				isStart = true
				msgTimer = newTimer(time.Second)
				msgTimer.run()
			}
			// 发送 SendCount 次消息
			// 500个人每次发一次消息
			for i := start; i < end; i++ {
				client, ok := m.clients.Load(m.userIds[i])
				if !ok {
					continue
				}
				// 发送信息
				client.(*Client).sendInPB(m.chatId)
				//client.(*Client).send(m.chatId)
			}
			count++
			if count >= m.TestCount {
				fmt.Println("测试完成")
				return
			}
		}
	}
}
