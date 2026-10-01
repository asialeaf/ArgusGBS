package manscdp

import "testing"

func TestDecodeKeepalive(t *testing.T) {
	raw := `<?xml version="1.0" encoding="UTF-8"?>
<Notify>
<CmdType>Keepalive</CmdType>
<SN>43</SN>
<DeviceID>34020000001320000001</DeviceID>
<Status>OK</Status>
</Notify>`
	env, err := Decode([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if env.CmdType != "Keepalive" || env.DeviceID != "34020000001320000001" {
		t.Fatalf("%+v", env)
	}
}

func TestDecodeCatalog2022(t *testing.T) {
	raw := `<Response><CmdType>Catalog</CmdType><SN>1</SN><DeviceID>34020000001180000001</DeviceID><SumNum>1</SumNum>
<DeviceList Num="1"><Item><DeviceID>34020000001320000001</DeviceID><Name>cam</Name><Status>ON</Status><DownloadSpeed>1/2/4</DownloadSpeed></Item></DeviceList></Response>`
	env, err := Decode([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !env.Is2022() {
		t.Fatal("expected 2022")
	}
	if env.DeviceList.Item[0].Name != "cam" {
		t.Fatal(env.DeviceList.Item[0])
	}
}

func TestPTZStop(t *testing.T) {
	if s := PTZCmd("stop", 0); s != "A50F0100000000B5" {
		t.Fatal(s)
	}
	if s := PTZCmd("left", 129); s != "A50F010281000038" {
		t.Fatal(s)
	}
	if s := PTZCmd("right", 129); s != "A50F010181000037" {
		t.Fatal(s)
	}
	if s := PTZCmd("up", 129); s != "A50F01080081003E" {
		t.Fatal(s)
	}
	if s := PTZCmd("down", 129); s != "A50F01040081003A" {
		t.Fatal(s)
	}
}
