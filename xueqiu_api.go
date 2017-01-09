package xueqiu_api

import (
    "crypto/md5"
    "io"
    "encoding/hex"
    "strings"
    "net/http/cookiejar"
    "net/http"
    "time"
    "io/ioutil"
    "net/url"
    "xueqiu_api/model"
    "encoding/json"
    "strconv"
)

const (
    url_csrf = "https://xueqiu.com/service/csrf?api=/user/login"
    url_login = "https://xueqiu.com/user/login"
    url_stockList = "https://xueqiu.com/stock/cata/stocklist.json"
    url_stockDetail = "https://xueqiu.com/v4/stock/quote.json"
    url_stockDataForChart = "https://xueqiu.com/stock/forchart/stocklist.json"
)

func md5hex(str string) string {
    h := md5.New()
    io.WriteString(h, str)
    resBytes := h.Sum(nil)
    resStr := hex.EncodeToString(resBytes)
    return strings.ToUpper(resStr)
}

type Controller struct {
    jar      *cookiejar.Jar
    client   *http.Client
    username string
    password string
}

func New(username, password string) (*Controller) {
    jar, _ := cookiejar.New(nil)
    client := &http.Client{
        Jar: jar,
        Timeout: time.Second * 10,
    }
    return &Controller{
        jar: jar,
        client: client,
        username: username,
        password: md5hex(password),
    }
}

func (me *Controller) Login() error {
    // csrf request
    csrfReq, err := http.NewRequest("GET", url_csrf, nil)
    if err != nil {
        return err
    }
    resp, err := me.client.Do(csrfReq)
    if err != nil {
        return err
    }
    resp.Body.Close()

    // login request
    postData := url.Values{}
    postData.Set("telephone", me.username)
    postData.Set("remember_me", "on")
    postData.Set("areacode", "86")
    postData.Set("password", me.password)
    postDataReader := strings.NewReader(postData.Encode())
    loginReq, err := http.NewRequest("POST", url_login, postDataReader)
    loginReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    if err != nil {
        return err
    }

    resp, err = me.client.Do(loginReq)
    if err != nil {
        return err
    }
    resp.Body.Close()
    return nil
}

func (me *Controller) GetCodeList() (list []string) {
    pageNum := 0
    count := -1
    size := 100
    params := url.Values{}
    params.Set("size", strconv.Itoa(size))
    params.Set("order", "asc")
    params.Set("orderby", "code")
    params.Set("type", "0,1,2")
    for {
        codeList := model.CodeList{}
        pageNum += 1
        params.Set("page", strconv.Itoa(pageNum))
        req, _ := http.NewRequest("GET", url_stockList + "?" + params.Encode(), nil)
        resp, _ := me.client.Do(req)
        body, _ := ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        json.Unmarshal(body, &codeList)
        for _, item := range codeList.Stocks {
            if code, ok := item["code"]; ok {
                list = append(list, code)
            }
        }
        if count == -1 {
            if c, ok := codeList.Count["count"]; ok {
                count = c
            }
        }
        count -= len(codeList.Stocks)
        if size > count {
            params.Set("size", strconv.Itoa(count))
        }
        if count <= 0 {
            break
        }
    }
    return
}

func (me *Controller) GetDetail(codes string) (list []interface{}) {
    req, _ := http.NewRequest("GET", url_stockDetail + "?code=" + codes, nil)
    resp, _ := me.client.Do(req)
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    stocks := map[string]interface{}{}
    json.Unmarshal(body, &stocks)
    for _, item := range stocks {
        list = append(list, item)
    }
    return list
}
