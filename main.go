package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shopspring/decimal"
)

var urltmp = "https://webcast3-normal-c-lf.amemv.com/webcast/ranklist/room/%s/contributor/" +
	"?room_id=%s&sec_anchor_id=MS4wLjABAAAAvttHFccCWP24gPnnZd68FJY4nojgNaIyGpKuaGW9F9c&" +
	"sec_user_id=MS4wLjABAAAAVGcxEZhWezHABIs0mRFR1BTMGpT39-8tLYNJmKXgXGeBf7PLBEC32VC3mWqsS6fl&rank_type=30&" +
	"webcast_sdk_version=1560&webcast_language=zh&webcast_locale=zh_CN&manifest_version_code=110501&" +
	"_rticket=%d" + //毫秒
	"&app_type=normal&iid=738487158705304&channel=gdt_growth14_big_yybwz&device_type=M2002J9E&language=zh&" +
	"cpu_support64=true&host_abi=armeabi-v7a&uuid=249315281287458&resolution=1080*2287&openudid=9762155d7de24d4f&update_version_code=11509900&" +
	"cdid=998a9881-3dae-4813-aa12-a7f4f3f44cf0&os_api=29&mac_address=d8%3A5a%3Ad9%3A52%3Ade%3Aba&dpi=440&oaid=58bcf74415b90a9a&ac=wifi&" +
	"device_id=2585652498273976&mcc_mnc=46001&os_version=10&version_code=110500&app_name=aweme&version_name=11.5.0&" +
	"device_brand=Xiaomi&ssmix=a&device_platform=android&aid=1128&" +
	"ts=%d" //秒

type Info struct {
	Times     int
	LastScore int
}

type InfoID struct {
	Info
	ID string
}

var (
	serverHost string = ""
	client            = http.DefaultClient
	hasMap            = make(map[string]*Info)

	roomID    string
	currentID *InfoID = new(InfoID)
	firstID   string
	f         *os.File
)

func main() {
	roomID = getRoomID()
	var err error
	f, err = os.OpenFile("./1.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	b, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(b, &hasMap)
	if err != nil && len(b) > 0 {
		fmt.Println("json map err: ", err)
		os.Exit(2)
	}
	defer f.Close()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	table := widgets.NewTable()
	x, y := ui.TerminalDimensions()
	table.SetRect(0, 0, x, y)
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.TextAlignment = ui.AlignCenter
	getRows(table)
	ui.Render(table)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var row []string
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				b, _ := json.Marshal(hasMap)
				f.Truncate(0)
				f.Seek(0, 0)
				f.Write(b)
				return

			case "<Escape>":
				currentID.ID = ""
				ticker.Reset(10 * time.Second)
				getRows(table)
				ui.Render(table)

			case "<Enter>":
				if len(table.Rows) > 2 {
					row = table.Rows[1]
				} else {
					continue
				}
				if currentID.ID == "" {
					currentID.ID = firstID
					currentID.LastScore, _ = strconv.Atoi(row[2])
					row[4] = "测骰子ing..."
					ticker.Reset(10 * time.Second)
					ui.Render(table)

				} else {
					if info, ok := hasMap[currentID.ID]; ok {
						info.LastScore += currentID.LastScore //这里加上相对音浪
						info.Times++
					} else {
						info = new(Info)
						hasMap[currentID.ID] = info
						info.LastScore = currentID.LastScore
						info.Times = 1
					}

					currentID.ID = ""
					ticker.Reset(10 * time.Second)
					getRows(table)
					ui.Render(table)
				}
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				num, err := strconv.Atoi(e.ID)
				if err != nil || currentID.ID != "" {
					continue
				}
				row := table.Rows[num]
				currentID.ID = row[5]
				currentID.LastScore, _ = strconv.Atoi(row[2])
				row[4] = "测骰子ing..."
				ticker.Reset(10 * time.Second)
				ui.Render(table)
			}

		case <-ticker.C:
			getRows(table)
			ui.Render(table)
		}
	}
}

func getRows(table *widgets.Table) {
	rows := table.Rows
	ret := getData()
	lret := len(ret.Data.Ranks)
	ltab := len(table.Rows)

	i := lret - ltab
	if ltab == 0 { //初始化header
		rows = make([][]string, 0, lret+1)
		for ; lret >= 0; lret-- {
			rows = append(rows, make([]string, 6, 6))
		}
		row := rows[0]
		row[0] = "排行"
		row[1] = "昵称"
		row[2] = "相对音浪"
		row[3] = "测骰子次数"
		row[4] = "状态"
		row[5] = "UID"
	} else if i >= 0 { //如果返回的len大于现在的row，添加剩下的
		for ; i >= 0; i-- {
			rows = append(rows, make([]string, 6, 6))
		}
	}

	//根据数据填充rows
	index := 1
	for _, row := range ret.Data.Ranks {
		rows[index][0] = ""
		rows[index][1] = row.User.Name
		if info, ok := hasMap[row.User.ID]; ok {
			rows[index][2] = strconv.Itoa(row.Score - info.LastScore)
			rows[index][3] = strconv.Itoa(info.Times)
		} else {
			rows[index][2] = strconv.Itoa(row.Score)
			rows[index][3] = strconv.Itoa(0)
		}

		if currentID.ID == row.User.ID {
			rows[index][4] = "测骰子ing..."
		} else {
			rows[index][4] = "排队中"
		}
		rows[index][5] = row.User.ID
		index++
	}

	//根据相对音浪排序
	r := rows[1:]
	sort.Slice(r, func(i, j int) bool {
		x, err := strconv.Atoi(r[i][2])
		if err != nil {
			panic(err)
		}
		y, err := strconv.Atoi(r[j][2])
		if err != nil {
			panic(err)
		}

		return x > y
	})

	if len(rows) >= 2 {
		firstID = rows[1][5]
	}
	//重新计算排名
	for i = 1; i < len(rows); i++ {
		rows[i][0] = strconv.Itoa(i)
	}

	table.Rows = rows
}

func getGorgon(url string) string {
	res, err := client.Post(serverHost, "", strings.NewReader(url))
	if err != nil {
		fmt.Println("server shutdown")
		return ""
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	return string(b)
}

func getData() *resp {
	ts := time.Now().Unix()
	url := fmt.Sprintf(urltmp, roomID, roomID, ts*1000, ts)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("sdk-version", "1")
	req.Header.Add("Cookie", "install_id=4212964517742397; ttreq=1$d508cd56f0174ab5f688fc6f43e74fb6c92f10e6; odin_tt=4d20af5c89b0b3261975bec2c6e292918d8b655498c3218ecebe8c25488f05236932f32a4270267f5c9f2916ddfdf5e3c92de9e86c90f58b8383b3f13df1b972; SLARDAR_WEB_ID=d7fd01c7-5acd-4565-aa30-6ca7cf14f705")
	req.Header.Add("Host", "webcast3-normal-c-lf.amemv.com")
	req.Header.Add("User-Agent", "com.ss.android.ugc.aweme/110501 (Linux; U; Android 10; zh_CN; M2002J9E; Build/QKQ1.191222.002; Cronet/TTNetVersion:3c28619c 2020-05-19 QuicVersion:0144d358 2020-03-24)")
	req.Header.Add("X-Gorgon", "")
	req.Header.Add("X-Khronos", fmt.Sprintf("%d", ts))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("do req err: ", err)
		return nil
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)
	ret := new(resp)
	err = json.Unmarshal(b, ret)
	if err != nil {
		fmt.Println("json err")
		return nil
	}

	return ret
}

func getRoomID() string {
	uid := "3737967248287559"
	ut := "https://webcast3-normal-c-lf.amemv.com/webcast/room/live_room_id/?manifest_version_code=110501&_rticket=%d&app_type=normal&iid=738487158705304&channel=gdt_growth14_big_yybwz&device_type=M2002J9E&language=zh&cpu_support64=true&host_abi=armeabi-v7a&uuid=249315281287458&resolution=1080*2287&openudid=9762155d7de24d4f&update_version_code=11509900&cdid=998a9881-3dae-4813-aa12-a7f4f3f44cf0&os_api=29&mac_address=d8%3A5a%3Ad9%3A52%3Ade%3Aba&dpi=440&oaid=58bcf74415b90a9a&ac=wifi&device_id=2585652498273976&mcc_mnc=46001&os_version=10&version_code=110500&app_name=aweme&version_name=11.5.0&device_brand=Xiaomi&ssmix=a&device_platform=android&aid=1128&ts=%d"
	ts := time.Now().Unix()
	url := fmt.Sprintf(ut, ts*1000, ts) // uid 3737967248287559
	req, _ := http.NewRequest("POST", url, strings.NewReader("user_id="+uid))
	req.Header.Add("sdk-version", "1")
	req.Header.Add("Cookie", "install_id=4212964517742397; ttreq=1$d508cd56f0174ab5f688fc6f43e74fb6c92f10e6; odin_tt=4d20af5c89b0b3261975bec2c6e292918d8b655498c3218ecebe8c25488f05236932f32a4270267f5c9f2916ddfdf5e3c92de9e86c90f58b8383b3f13df1b972; SLARDAR_WEB_ID=d7fd01c7-5acd-4565-aa30-6ca7cf14f705")
	req.Header.Add("Host", "webcast3-normal-c-lf.amemv.com")
	req.Header.Add("User-Agent", "com.ss.android.ugc.aweme/110501 (Linux; U; Android 10; zh_CN; M2002J9E; Build/QKQ1.191222.002; Cronet/TTNetVersion:3c28619c 2020-05-19 QuicVersion:0144d358 2020-03-24)")
	req.Header.Add("X-Gorgon", "")
	req.Header.Add("X-Khronos", fmt.Sprintf("%d", ts))

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	ret := &struct {
		Data struct {
			RoomID map[string]decimal.Decimal `json:"room_id"`
		} `json:"data"`

		StatusCode int `json:"status_code"`
	}{}

	err = json.Unmarshal(b, ret)
	if err != nil {
		panic(err)
	}
	return ret.Data.RoomID[uid].String()
}
