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
    "encoding/json"
    "strconv"
)

const (
    url_csrf = "https://xueqiu.com/service/csrf?api=/user/login"
    url_login = "https://xueqiu.com/user/login"
    url_stockList = "https://xueqiu.com/stock/cata/stocklist.json"
    url_stockDetail = "https://xueqiu.com/v4/stock/quote.json"
    url_stockDataForChart = "https://xueqiu.com/stock/forchart/stocklist.json"

    url_events = "https://xueqiu.com/calendar/cal/events.json"
)

type CodeList struct {
    Count map[string]int
    Success bool
    Stocks [](map[string]string)
}

type StockInfo map[string]string

type EventItem map[string]interface{}
type EventSet map[string][]EventItem

//"id": 20974531,
//"author_id": -1,
//"calendar_id": 20194324,
//"title": "盘前 披露财报，预期EPS -0.02",
//"timezone": "US/Eastern",
//"color": "blue",
//"start_date": 1487088000000,
//"end_date": null,
//"location": "",
//"description": "盘前 披露财报，预期EPS -0.02 http://www.nasdaq.com/earnings/report/grpn",
//"url": null,
//"stock": "GRPN",
//"stock_event_type": 1,
//"best_editor_id": -1,
//"last_modified": 1486339203000,
//"created_at": 1486339203000,
//"all_day": false,
//"share_id": 0,
//"sequence": 0,
//"privacy_read": "0",
//"privacy_write": "1",
//"is_stock_event": true,
//"stat": null,
//"stock_name": "Groupon"


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
        codeList := CodeList{}
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

func (me *Controller) GetEvents(code string) (eventSet EventSet) {
    params := url.Values{}
    params.Set("symbol", code)
    params.Set("end_date", "INF")
    params.Set("page", "1")
    params.Set("count", "3")
    req, _ := http.NewRequest("GET", url_events + "?" + params.Encode(), nil)
    resp, _ := me.client.Do(req)
    body, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    json.Unmarshal(body, & eventSet)
    return
}

func (me *Controller) GetDetail(codes string) (list []interface{}) {
    req, _ := http.NewRequest("GET", url_stockDetail + "?code=" + codes, nil)
    resp, _ := me.client.Do(req)
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    stocks := map[string](map[string]string){}
    json.Unmarshal(body, &stocks)
    for _, item := range stocks {
        itemCast := map[string]interface{}{}
        for k, v := range item {
            switch k {
            case "volume", "current", "instOwn", "low52week", "high52week",
                "marketCapital", "pe_ttm", "pe_lyr", "net_assets",
                "moving_avg_200_day", "chg_from_200_day_moving_avg", "pct_chg_from_200_day_moving_avg",
                "moving_avg_50_day", "chg_from_50_day_moving_avg", "pct_chg_from_50_day_moving_avg":
                    val, err := strconv.ParseFloat(v, 64)
                    if err != nil {
                        itemCast[k] = 0
                    } else {
                        itemCast[k] = val
                    }
            default:
                    itemCast[k] = v
            }
        }
        list = append(list, itemCast)
    }
    return list
}
