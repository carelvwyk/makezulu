## Note
This is not a complete project but it will provide some reference for those
attending OfferZen Make Days and want to control Sphero Spark+ bots using 
Go on OSX. Also includes a quickly hacked together lib for doing AWS IoT PubSub, check
"iotPubSubDemo" in main.go for example usage. 

## Set up
Reference: https://gobot.io/documentation/platforms/sprkplus/

go get -d -u gobot.io/x/gobot/...

go get -d -u github.com/nsf/termbox-go

go get -d -u github.com/eclipse/paho.mqtt.golang

## Example invocation on OSX
GODEBUG=cgocheck=0 go run main.go --iot_privkey="$(< keys/a9b12b50d5-private.pem.key)" --iot_cert="$(< keys/a9b12b50d5-certificate.pem.crt)" --iot_thingname=GoGoBot
