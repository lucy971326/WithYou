package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Voice 是给前端展示的一个音色选项，预置和克隆共用。
type Voice struct {
	Voice       string `json:"voice"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Custom      bool   `json:"custom"`
	TargetModel string `json:"target_model,omitempty"`
}

// voicesResponse 是 GET /api/voices 的返回体。
type voicesResponse struct {
	DefaultVoice string  `json:"default_voice"`
	Preset       []Voice `json:"preset"`
	Custom       []Voice `json:"custom"`
	CustomError  string  `json:"custom_error,omitempty"`
}

// presetVoices 是 Qwen3.5-Omni / Realtime 系列的官方预置音色快照。
// 官方页面：https://help.aliyun.com/zh/model-studio/omni-voice-list （2026-05-22）
var presetVoices = []Voice{
	{Voice: "Tina", Name: "甜甜 Tina", Description: "我的声音像温热的奶茶，甜甜的、暖暖的，但解决问题可一点都不含糊哦！"},
	{Voice: "Cindy", Name: "林欣宜 Cindy", Description: "台湾说话嗲嗲的小姐姐"},
	{Voice: "Liora Mira", Name: "清欢 Liora Mira", Description: "用声音织就烟火人间的温柔"},
	{Voice: "Sunnybobi", Name: "知芝 Sunnybobi", Description: "大大咧咧的社恐邻家姑娘"},
	{Voice: "Raymond", Name: "林川野 Raymond", Description: "声音清亮，爱吃外/卖的宅男"},
	{Voice: "Ethan", Name: "晨煦 Ethan", Description: "标准普通话，带部分北方口音。阳光 温暖 活力 朝气"},
	{Voice: "Theo Calm", Name: "予安 Theo Calm", Description: "在静默处传递理解，在言语间疗愈人心。"},
	{Voice: "Serena", Name: "苏瑶 Serena", Description: "温柔小姐姐"},
	{Voice: "Harvey", Name: "厚 Harvey", Description: "我的声音来自岁月沉淀——低沉、温和，带着一点咖啡与旧书的气息。"},
	{Voice: "Maia", Name: "四月 Maia", Description: "知性与温柔的碰撞"},
	{Voice: "Evan", Name: "江晨 Evan", Description: "男大学生，年下奶狗"},
	{Voice: "Qiao", Name: "小乔妹 Qiao", Description: "超关键！她不是普通可爱，而是'表面甜妹，个性十足'"},
	{Voice: "Momo", Name: "茉兔 Momo", Description: "撒娇搞怪，逗你开心"},
	{Voice: "Wil", Name: "伟伦 Wil", Description: "在深圳长大的港台腔小哥哥"},
	{Voice: "Angel", Name: "台普 - 安琪 Angel", Description: "略带台式口音，她超甜的哦！"},
	{Voice: "Li Cassian", Name: "东厂 - 李公公 Li Cassian", Description: "话中三分留白、七分察言观色"},
	{Voice: "Mia", Name: "温柔生活博主 - 舒然 Mia", Description: "以细腻声音，传递慢生活美学与日常治愈力量的生活艺术家"},
	{Voice: "Joyner", Name: "喜剧担当 - 阿逗 Joyner", Description: "搞笑、夸张、接地气"},
	{Voice: "Gold", Name: "金爷 Gold", Description: "西海岸黑人 Rapper"},
	{Voice: "Katerina", Name: "卡捷琳娜 Katerina", Description: "御姐音色，韵律回味十足"},
	{Voice: "Ryan", Name: "甜茶 Ryan", Description: "节奏拉满，戏感炸裂，真实与张力共舞"},
	{Voice: "Jennifer", Name: "詹妮弗 Jennifer", Description: "品牌级、电影质感般美语女声"},
	{Voice: "Aiden", Name: "艾登 Aiden", Description: "精通厨艺的美语大男孩"},
	{Voice: "Mione", Name: "敏儿 Mione", Description: "成熟，知性英国邻家妹妹"},
	{Voice: "Sunny", Name: "四川 - 晴儿 Sunny", Description: "甜到你心里的川妹子"},
	{Voice: "Dylan", Name: "北京 - 晓东 Dylan", Description: "北京胡同里长大的少年"},
	{Voice: "Eric", Name: "四川 - 程川 Eric", Description: "一个跳脱市井的四川成都男子"},
	{Voice: "Peter", Name: "天津 - 李彼得 Peter", Description: "天津相声，专业捧哏"},
	{Voice: "Joseph Chen", Name: "阿樸伯 Joseph Chen", Description: "我是阿樸伯，本名陳志樸，南洋老華僑。"},
	{Voice: "Marcus", Name: "陕西 - 秦川 Marcus", Description: "面宽话短，心实声沉——老陕的味道。"},
	{Voice: "Li", Name: "南京 - 老李 Li", Description: "骂骂咧咧的伯伯"},
	{Voice: "Kiki", Name: "粤语-阿清", Description: "甜美的港妹闺蜜"},
	{Voice: "Rocky", Name: "粤语 - 阿强 Rocky", Description: "幽默风趣的阿强，在线陪聊"},
	{Voice: "Sohee", Name: "素熙 Sohee", Description: "温柔开朗，情绪丰富的韩国欧尼"},
	{Voice: "Lenn", Name: "莱恩 Lenn", Description: "理性是底色，叛逆藏在细节里——穿西装也听后朋克的德国青年。"},
	{Voice: "Ono Anna", Name: "小野杏 Ono Anna", Description: "鬼灵精怪的青梅竹马"},
	{Voice: "Sonrisa", Name: "索尼莎 Sonrisa", Description: "热情开朗的拉美大姐"},
	{Voice: "Bodega", Name: "博德加 Bodega", Description: "热情的西班牙大叔"},
	{Voice: "Emilien", Name: "埃米尔安 Emilien", Description: "浪漫的法国大哥哥"},
	{Voice: "Andre", Name: "安德雷 Andre", Description: "声音磁性，自然舒服、沉稳男生"},
	{Voice: "Radio Gol", Name: "拉迪奥·戈尔 Radio Gol", Description: "足球诗人 Rádio Gol！今天我要用名字为你们解说足球。"},
	{Voice: "Alek", Name: "阿列克 Alek", Description: "一开口，是战斗民族的冷，也是毛呢大衣下的暖"},
	{Voice: "Rizky", Name: "阿力 Rizky", Description: "印尼的青年小伙，声线个性"},
	{Voice: "Roya", Name: "萝雅 Roya", Description: "热爱运动的女孩，拥有一颗自由的心。"},
	{Voice: "Arda", Name: "阿尔达 Arda", Description: "不高亢，也不低沉，干净利落中带着温润的气质"},
	{Voice: "Hana", Name: "阿幸 Hana", Description: "爱狗狗的越南成熟姐姐"},
	{Voice: "Dolce", Name: "多尔切 Dolce", Description: "慵懒的意大利大叔"},
	{Voice: "Jakub", Name: "雅克 Jakub", Description: "波兰小镇文艺青年，声线磁性性感"},
	{Voice: "Griet", Name: "海娜 Griet", Description: "荷兰成熟又文艺的女性"},
	{Voice: "Eliška", Name: "艾莉卡 Eliška", Description: "每个单词都传递中欧的匠心与温度"},
	{Voice: "Marina", Name: "玛丽娜 Marina", Description: "一个在多元文化城市中长大的女孩。"},
	{Voice: "Siiri", Name: "西芮 Siiri", Description: "内敛温柔，语速舒缓如湖面微澜"},
	{Voice: "Ingrid", Name: "林恩 Ingrid", Description: "挪威乡村姑娘"},
	{Voice: "Sigga", Name: "海娜 Sigga", Description: "冰岛小镇的知性女青年"},
	{Voice: "Bea", Name: "雅娜 Bea", Description: "爱喝咖啡的菲律宾甜甜小姐姐"},
	{Voice: "Chloe", Name: "思怡 Chloe", Description: "马来西亚白领女生"},
}

type cloneListRequest struct {
	Model string `json:"model"`
	Input struct {
		Action    string `json:"action"`
		PageIndex int    `json:"page_index"`
		PageSize  int    `json:"page_size"`
	} `json:"input"`
}

type cloneListResponse struct {
	Output struct {
		TotalCount int `json:"total_count"`
		VoiceList  []struct {
			Voice       string `json:"voice"`
			TargetModel string `json:"target_model"`
			GMTCreate   string `json:"gmt_create"`
		} `json:"voice_list"`
	} `json:"output"`
}

// listClonedVoices 调官方声音复刻接口，分页拉当前账号的克隆音色。
// 缺 key 时返回空列表；接口失败由调用方记日志。
func listClonedVoices(ctx context.Context, apiKey, endpoint string) ([]Voice, error) {
	if apiKey == "" || endpoint == "" {
		return nil, nil
	}

	client := &http.Client{Timeout: 8 * time.Second}
	const pageSize = 100
	var out []Voice

	for page := 0; page < 10; page++ {
		reqBody := cloneListRequest{Model: "qwen-voice-enrollment"}
		reqBody.Input.Action = "list"
		reqBody.Input.PageIndex = page
		reqBody.Input.PageSize = pageSize

		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("voices: marshal list request: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("voices: new list request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("voices: list cloned voices: %w", err)
		}
		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("voices: read list response: %w", readErr)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("voices: list status %d: %s", resp.StatusCode, string(payload))
		}

		var parsed cloneListResponse
		if err := json.Unmarshal(payload, &parsed); err != nil {
			return nil, fmt.Errorf("voices: parse list response: %w", err)
		}
		for _, v := range parsed.Output.VoiceList {
			out = append(out, Voice{
				Voice:       v.Voice,
				Custom:      true,
				TargetModel: v.TargetModel,
			})
		}
		if len(parsed.Output.VoiceList) < pageSize || (parsed.Output.TotalCount > 0 && len(out) >= parsed.Output.TotalCount) {
			break
		}
	}
	return out, nil
}
