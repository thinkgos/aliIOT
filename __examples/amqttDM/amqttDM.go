package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/thinkgos/aliIOT"
	"github.com/thinkgos/aliIOT/infra"
	"github.com/thinkgos/aliIOT/model"
	"github.com/thinkgos/aliIOT/sign"
)

const (
	productKey    = "a1QR3GD1Db3"
	productSecret = ""
	deviceName    = "MPA19GT010070140"
	deviceSecret  = "CsC7Gmb6EvDLOm8V40HLOQwFPdc3KCHT"
)

func main() {
	signs, err := sign.NewMQTTSign().
		SetSDKVersion(infra.IOTSDKVersion).
		Generate(&sign.MetaInfo{
			ProductKey:    productKey,
			ProductSecret: productSecret,
			DeviceName:    deviceName,
			DeviceSecret:  deviceSecret}, sign.CloudRegionShangHai)
	if err != nil {
		panic(err)
	}
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("%s:%d", signs.HostName, signs.Port)).
		SetClientID(signs.ClientID).
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
	client := mqtt.NewClient(opts)

	dmopt := model.NewOption(productKey, deviceName, deviceSecret).
		SetEnableCache(true).
		Valid()
	manage := aliIOT.NewWithMQTT(dmopt, client)
	manage.LogMode(true)

	client.Connect().Wait()

	_ = manage.Subscribe(manage.URIServiceSelf(model.URISysPrefix, model.URIThingEventPostReplySingleWildcard), model.ProcThingEventPostReply)
	//_ = manage.Subscribe(manage.URIServiceSelf(model.URISysPrefix, model.URIThingServiceRequestMultiWildcard2), model.ProcThingServiceRequest)
	_ = manage.Subscribe(manage.URIServiceSelf(model.URISysPrefix, model.URIRRPCRequestSingleWildcard), model.ProcRRPCRequest)
	go func() {
		for {
			err := manage.UpstreamThingEventPost(model.DevItself, "tempAlarm", map[string]interface{}{
				"high": 1,
			})
			if err != nil {
				log.Printf("error: %#v", err)
			} else {
				log.Printf("success")
			}
			time.Sleep(time.Second * 10)
		}

	}()

	for {
		err = manage.UpstreamThingEventPropertyPost(model.DevItself, map[string]interface{}{
			"Temp":         rand.Intn(200),
			"Humi":         rand.Intn(100),
			"switchStatus": rand.Intn(1),
		})
		if err != nil {
			log.Printf("error: %#v", err)
		} else {
			log.Printf("success")
		}
		time.Sleep(time.Second * 30)
	}
}

//
//type UserProc struct {
//	model.DevUserProc
//}
