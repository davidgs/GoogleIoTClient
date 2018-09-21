package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/davidgs/SenseAir_K30_go"
	"github.com/davidgs/bme280_go"
	jwt "github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	deviceID = flag.String("device", "", "Cloud IoT Core Device ID")
	bridge   = struct {
		host *string
		port *string
	}{
		flag.String("mqtt_host", "mqtt.googleapis.com", "MQTT Bridge Host"),
		flag.String("mqtt_port", "8883", "MQTT Bridge Port"),
	}
	projectID   = flag.String("project", "", "GCP Project ID")
	registryID  = flag.String("registry", "", "Cloud IoT Registry ID (short form)")
	region      = flag.String("region", "", "GCP Region")
	certsCA     = flag.String("ca_certs", "", "Download https://pki.google.com/roots.pem")
	privateKey  = flag.String("private_key", "", "Path to private key file")
	measurement = flag.String("influx_db", "googleCoreIOT", "InfluxDB Measurement to store into")
	format      = flag.String("format", "line", "Data format: line or json")
)

func main() {
	log.Println("[main] Entered")

	log.Println("[main] Flags")
	flag.Parse()

	log.Println("[main] Loading Google's roots")
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(*certsCA)
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	log.Println("[main] Creating TLS Config")

	config := &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{},
		MinVersion:         tls.VersionTLS12,
	}

	clientID := fmt.Sprintf("projects/%v/locations/%v/registries/%v/devices/%v",
		*projectID,
		*region,
		*registryID,
		*deviceID,
	)

	log.Println("[main] Creating MQTT Client Options")
	opts := MQTT.NewClientOptions()

	broker := fmt.Sprintf("ssl://%v:%v", *bridge.host, *bridge.port)
	log.Printf("[main] Broker '%v'", broker)

	opts.AddBroker(broker)
	opts.SetClientID(clientID).SetTLSConfig(config)

	opts.SetUsername("unused")

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = jwt.StandardClaims{
		Audience:  *projectID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	log.Println("[main] Load Private Key")
	keyBytes, err := ioutil.ReadFile(*privateKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[main] Parse Private Key")
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[main] Sign String")
	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Fatal(err)
	}

	opts.SetPassword(tokenString)

	// Incoming
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("[handler] Topic: %v\n", msg.Topic())
		fmt.Printf("[handler] Payload: %v\n", msg.Payload())
	})

	log.Println("[main] MQTT Client Connecting")
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	topic := struct {
		config    string
		telemetry string
	}{
		config:    fmt.Sprintf("/devices/%v/config", *deviceID),
		telemetry: fmt.Sprintf("/devices/%v/events", *deviceID),
	}

	log.Println("[main] Creating Subscription")
	client.Subscribe(topic.config, 0, nil)
	dev := "/dev/i2c-1"
	bme := bme280_go.BME280{}
	r := bme.BME280Init(dev)
	if r < 0 {
		log.Println("[main] BME Init Error")
	}
	defer bme.Dev.Close()
	k30 := SenseAir_K30_go.K30{}
	r = k30.K30Init(dev)
	if r < 0 {
		log.Println("[main] K30 Init Error")
	}
	defer k30.Dev.Close()
	log.Println("[main] Publishing Messages")
	defer client.Disconnect(250)
	for 1 > 0 {
		mString := ""
		count := 0
		for count < 4 {
			rets := bme.BME280ReadValues()
			t := float64(float64(rets.Temperature) / 100.00)
			h := float64(rets.Humidity) / 1024.00
			if rets.Humidity != -1 && rets.Temperature > -100 {
				if *format == "json" {
					if mString == "" {
						mString = fmt.Sprintf("{\"measurement\": \"%s\", \"tags\": {\"sensor\": \"bme_280\"}, \"fields\": {\"temp_c\": %.2f, \"humidity\": %.2f}, \"time\": %d}", *measurement, t, h, time.Now().UnixNano())
					} else {
						mString = fmt.Sprintf("%s\n{\"measurement\": \"%s\", \"tags\": {\"sensor\": \"bme_280\"}, \"fields\": {\"temp_c\": %.2f, \"humidity\": %.2f}, \"time\": %d}", mString, *measurement, t, h, time.Now().UnixNano())
					}
				} else {
					if mString == "" {
						mString = fmt.Sprintf("%s,sensor=bme_280 temp_c=%.2f,humidity=%.2f %d", *measurement, t, h, time.Now().UnixNano())
					} else {
						mString = fmt.Sprintf("%s\n%s,sensor=bme_280 temp_c=%.2f,humidity=%.2f %d", mString, *measurement, t, h, time.Now().UnixNano())
					}
				}
				log.Printf("[main] Humidity: %.2f Temperature: %.2f", h, t)
			} else {
				log.Printf("[main] Temperature Reading Error")
			}
			time.Sleep(1500 * time.Millisecond)
			count++
		}
		co2Value := k30.K30ReadValue()
		if co2Value > 0 {
			if *format == "json" {
				if mString == "" {
					mString = fmt.Sprintf("{\"measurement\": \"%s\", \"tags\": {\"sensor\": \"k_30\"}, \"fields\": {\"ppm\": %d}, \"time\": %d}", *measurement, co2Value, time.Now().UnixNano())
				} else {
					mString = fmt.Sprintf("%s\n{\"measurement\": \"%s\", \"tags\": {\"sensor\": \"k_30\"}, \"fields\": {\"ppm\": %d}, \"time\": %d}", mString, *measurement, co2Value, time.Now().UnixNano())
				}
			} else {
				if mString == "" {
					mString = fmt.Sprintf("%s,sensor=k_30 ppm=%.2f %d", *measurement, co2Value, time.Now().UnixNano())
				} else {
					mString = fmt.Sprintf("%s\n%s,sensor=k_30 ppm=%.2f %d", mString, *measurement, co2Value, time.Now().UnixNano())
				}
			}

			log.Printf("[main] CO2: %d ", co2Value)
		} else {
			log.Println("[main] CO2 Reading Error")
		}
		token := client.Publish(
			topic.telemetry,
			0,
			false,
			mString)

		token.WaitTimeout(5 * time.Second)
		log.Printf("[main] %s\n", mString) //Publishing Message %s,sensor=bme_280 temp_c=%.2f,humidity=%.2f %d", *measurement, t, h, time.Now().UnixNano())

	}

}
