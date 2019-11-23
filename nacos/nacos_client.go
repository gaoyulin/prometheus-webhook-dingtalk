package nacos

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var applicationMap = make(map[string]string) ;

func main() {
	var namespaceId = "2afe13ee-d5e4-4154-8146-15f973b27a63"
	configClient, error := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": []constant.ServerConfig{
			{
				IpAddr: "ops.ximalaya.com",
				Port:   80,
			},
		},
		"clientConfig": constant.ClientConfig{
			NamespaceId:         namespaceId,
			TimeoutMs:           20000,
			ListenInterval:      100000,
			NotLoadCacheAtStart: true,
			LogDir:              "data/nacos/log",
		},
	})
	println("start")
	println(error)

	var dataId = "dingding.json"
	var group = "APP"
	// 获取配置
	content, error := configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group})
	if error != nil {
		print(error)
	}
	fmt.Println(content)
	byteContent := []byte(content)
	m := make(map[string]interface{})
	_ = json.Unmarshal(byteContent, &m)
	if error != nil {
		print(error)
	}
	jsonToMap(m)
	// 监听配置
	configClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("ListenConfig group:" + group + ", dataId:" + dataId + ", data:" + data)
			_ = json.Unmarshal(byteContent, &m)
			if error != nil {
				print(error)
			}
			jsonToMap(m)
		},
	})
}

func jsonToMap(m map[string]interface{})  {
	if v, ok := m["applications"]; ok {
		ws := v.([]interface{})
		for _, wsItem := range ws {
			wsMap := wsItem.(map[string]interface{})
			var key string = ""
			var value string = ""
			if w, ok := wsMap["app"]; ok {
				key = w.(string)
			}
			if w, ok := wsMap["mobile"]; ok {
				value = w.(string)
			}
			applicationMap[key] = value
			println()
		}
	}
}
