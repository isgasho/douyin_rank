package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestMap(t *testing.T) {
	b := []byte(`{"54797587468":{"Times":1,"LastScore":270},"67941541702":{"Times":1,"LastScore":311},"92515313199":{"Times":1,"LastScore":263}}`)

	err := json.Unmarshal(b, &hasMap)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", hasMap)
}

func TestRoomID(t *testing.T) {
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
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
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
	t.Log(string(b))
	t.Log(ret.Data.RoomID[uid].String())
}
