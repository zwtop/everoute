/*
Copyright 2021 The Lynx Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/smartxworks/lynx/plugin/tower/pkg/utils"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type Client struct {
	Url      string
	AuthInfo AuthInformation

	token string
}

const (
	responseChanLenth = 10
)

// subscription subscribe change of objects, subscribe will stop when get response error, subscribe
// also could be stopped by run return function stopWatch().
func (c *Client) Subscription(req *Request) (respCh <-chan Response, stopWatch func(), err error) {
	var respChan = make(chan Response, responseChanLenth)

	msg := Message{
		ID:   string(uuid.NewUUID()),
		Type: StartMsg,
	}

	msg.PayLoad, err = json.Marshal(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %s", err)
	}

	conn, err := c.newWebsocketConn()
	if err != nil {
		return nil, nil, err
	}

	if err = conn.WriteJSON(msg); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to write message %v: %s", msg, err)
	}

	var stopChan = make(chan struct{})
	go lookReadMessage(conn, respChan, stopChan)

	return respChan, closeChanFunc(stopChan), nil
}

// query send query request to tower
func (c *Client) Query(req *Request) (*Response, error) {
	var reqBody, respBody bytes.Buffer
	var resp Response

	if err := json.NewEncoder(&reqBody).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode request: %s", err)
	}

	r, err := http.NewRequest(http.MethodPost, c.Url, &reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed call http.NewRequest: %s", err)
	}
	c.setScheme(r.URL, false)
	c.setHeader(r.Header, false)

	httpResp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if _, err := io.Copy(&respBody, httpResp.Body); err != nil {
		return nil, fmt.Errorf("failed to read reponse body: %s", err)
	}

	if err := json.NewDecoder(&respBody).Decode(&resp); err != nil {
		return nil, fmt.Errorf("server response code: %d, err: %s", httpResp.StatusCode, err)
	}

	return &resp, nil
}

// query send login request to tower, and save token
func (c *Client) Auth() error {
	var authRequest = &Request{
		Query:     "mutation($data: LoginInput!) {login(data: $data) {token}}",
		Variables: map[string]interface{}{"data": c.AuthInfo},
	}

	resp, err := c.Query(authRequest)
	if err != nil {
		return fmt.Errorf("failed to login tower: %s", err)
	}

	if len(resp.Errors) != 0 {
		return fmt.Errorf("receive unexpected errors: %v", resp.Errors)
	}

	tokenRaw := utils.LookupJsonRaw(resp.Data, "login", "token")
	err = json.Unmarshal(tokenRaw, &c.token)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s to token: %s", tokenRaw, err)
	}

	return nil
}

func (c *Client) newWebsocketConn() (*websocket.Conn, error) {
	header := http.Header{}
	u, err := url.Parse(c.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %s", c.Url, err)
	}

	c.setScheme(u, true)
	c.setHeader(header, true)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, fmt.Errorf("failed to dialer %s: %s", u, err)
	}

	return conn, nil
}

func (c *Client) setHeader(header http.Header, websocket bool) {
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")

	if c.token != "" {
		header.Set("Authorization", c.token)
	}

	if websocket {
		header.Set("Sec-Websocket-Protocol", "graphql-ws")
	}
}

// when use http, set scheme https/http; when use ws, set scheme wss/ws.
func (c *Client) setScheme(u *url.URL, websocket bool) {
	var secure bool

	switch u.Scheme {
	case "https", "wss":
		secure = true
	}

	u.Scheme = "http"
	if websocket {
		u.Scheme = "ws"
	}

	if secure {
		u.Scheme = fmt.Sprintf("%ss", u.Scheme)
	}
}

// lookReadMessage loop read message from conn until read error or get signal from stopChan
func lookReadMessage(conn *websocket.Conn, respChan chan<- Response, stopChan chan struct{}) {
	defer close(respChan)
	defer conn.Close()

	go func() {
		for {
			resp := readConnResponse(conn)
			select {
			case <-stopChan:
				// check if already stop before send to response chan
				return
			default:
				respChan <- resp
				if len(resp.Errors) != 0 {
					// stop watch if get response error
					closeChanFunc(stopChan)()
					return
				}
			}
		}
	}()

	<-stopChan
}

func readConnResponse(conn *websocket.Conn) Response {
	var msg Message
	var resp Response

	if err := conn.ReadJSON(&msg); err != nil {
		return connectErrorMessage("error read response message: %s", err)
	}

	switch msg.Type {
	case DataMsg:
	case ErrorMsg:
		return connectErrorMessage(string(msg.PayLoad))
	case CompleteMsg:
		return connectErrorMessage("unexpect complete msg, payload: %+v", msg.PayLoad)
	default:
		return connectErrorMessage("unknow message type %s, payload: %+v", msg.Type, msg.PayLoad)
	}

	if err := json.Unmarshal(msg.PayLoad, &resp); err != nil {
		return connectErrorMessage("error unmarshal json message: %s", err)
	}

	return resp
}

func connectErrorMessage(format string, a ...interface{}) Response {
	var resp Response
	resp.Errors = append(resp.Errors, ResponseError{
		Message: fmt.Sprintf(format, a...),
		Code:    WebsocketConnectError,
	})
	return resp
}

// closeChanFunc close chan once, prevent panic of multiple close chan.
func closeChanFunc(ch chan struct{}) func() {
	return func() {
		select {
		case <-ch:
			// skipped when chan already closed
		default:
			close(ch)
		}
	}
}
