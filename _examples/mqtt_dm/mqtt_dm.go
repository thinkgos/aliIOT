package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/thinkgos/go-core-package/lib/logger"

	aiot "github.com/thinkgos/aliyun-iot"
	"github.com/thinkgos/aliyun-iot/_examples/mock"
	"github.com/thinkgos/aliyun-iot/dm"
	"github.com/thinkgos/aliyun-iot/dmd"
	"github.com/thinkgos/aliyun-iot/infra"
	"github.com/thinkgos/aliyun-iot/sign"
)

var dmClient *aiot.MQTTClient

func main() {
	signs, err := sign.Generate(mock.MetaTriad, infra.CloudRegionDomain{Region: infra.CloudRegionShangHai})
	if err != nil {
		panic(err)
	}
	opts :=
		mqtt.NewClientOptions().
			AddBroker(signs.Addr()).
			SetClientID(signs.ClientIDWithExt()).
			SetUsername(signs.UserName).
			SetPassword(signs.Password).
			SetCleanSession(true).
			SetAutoReconnect(true).
			SetOnConnectHandler(func(cli mqtt.Client) {
				log.Println("mqtt client connection success")
			}).
			SetConnectionLostHandler(func(cli mqtt.Client, err error) {
				log.Println("mqtt client connection lost, ", err)
			})

	dmClient = aiot.NewWithMQTT(
		mock.MetaTriad,
		mqtt.NewClient(opts),
		dm.WithEnableNTP(),
		dm.WithEnableDesired(),
		dm.WithCallback(mockCb{}),
		dm.WithLogger(logger.New(log.New(os.Stdout, "mqtt --> ", log.LstdFlags), logger.WithEnable(true))),
	)

	dmClient.Underlying().Connect().Wait()
	if err = dmClient.Connect(); err != nil {
		panic(err)
	}

	//go DslTemplateTest()
	// go DesiredGetTest() // done
	// go DesiredDeleteTest()
	// go ConfigTest() // done
	// NTPTest() // done
	// DeviceInfoTest()  // done
	// ThingEventPost() // done
	for {
		time.Sleep(time.Second * 15)
		err := dmClient.LinkThingEventPropertyPost(mock.ProductKey, mock.DeviceName,
			mock.Instance{
				rand.Intn(200),
				rand.Intn(100),
				rand.Intn(2),
			}, time.Second)
		if err != nil {
			log.Printf("error: %#v", err)
		}
	}
}

// done
func ThingEventPost() {
	for {
		err := dmClient.LinkThingEventPost(mock.ProductKey, mock.DeviceName, "tempAlarm",
			map[string]interface{}{
				"high": 1,
			}, time.Second)
		if err != nil {
			log.Printf("error: %#v", err)
		}
		time.Sleep(time.Second * 10)
	}
}

// done
func DeviceInfoTest() {
	err := dmClient.LinkThingDeviceInfoUpdate(mock.ProductKey, mock.DeviceName,
		[]dmd.DeviceInfoLabel{
			{AttrKey: "attrKey", AttrValue: "attrValue"},
		}, time.Second*5)
	if err != nil {
		log.Println(err)
		return
	}
	time.Sleep(time.Minute * 1)
	err = dmClient.LinkThingDeviceInfoDelete(mock.ProductKey, mock.DeviceName,
		[]dmd.DeviceLabelKey{{AttrKey: "attrKey"}}, time.Second*5)
	if err != nil {
		log.Println(err)
		return
	}
}

// done
func ConfigTest() {
	cpd, err := dmClient.LinkThingConfigGet(mock.ProductKey, mock.DeviceName, time.Second*5)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(cpd)
}

func DslTemplateTest() {
	_, err := dmClient.LinkThingDsltemplateGet(mock.ProductKey, mock.DeviceName, time.Second*5)
	if err != nil {
		log.Println(err)
		return
	}
}

func dynamictslTest() {
	_, err := dmClient.LinkThingDynamictslGet(mock.ProductKey, mock.DeviceName, time.Second*5)
	if err != nil {
		panic(err)
	}
}

// done
func NTPTest() {
	err := dmClient.ExtNtpRequest()
	if err != nil {
		log.Println(err)
		return
	}
}

func DesiredGetTest() {
	data, err := dmClient.LinkThingDesiredPropertyGet(mock.ProductKey, mock.DeviceName, []string{"temp", "humi"}, time.Second*5)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("%+v", string(data))
}

func DesiredDeleteTest() {
	err := dmClient.LinkThingDesiredPropertyDelete(mock.ProductKey, mock.DeviceName, "{}", time.Second*5)
	if err != nil {
		log.Println(err)
		return
	}
}

type mockCb struct {
	dm.NopCb
}

func (sf mockCb) RRPCRequest(c *dm.Client, messageID, productKey, deviceName string, payload []byte) error {
	log.Println(string(payload))
	return nil
}
