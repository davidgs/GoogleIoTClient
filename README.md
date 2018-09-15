# Google Core IoT Example App

## Overview

* Entirely written in Go
* Designed to run on a Raspberry Pi
* Designed to read data from an [Adafruit](http://adafruit.com) BME280 Temperature/Humidity/Pressure sensor -- though it ignores the pressure portion
* Uses Go driver for BME280 from [davidgs](https://github.com/davidgs/bme280_go)
* Also uses the SenseAir K30 CO2 sensor with Go driver from [davidgs](https://github.com/davidgs/SenseAir_K30_go) 
* Writes 4 BME readings and 1 K30 reading to Google Core IoT per iteration (~5 seconds)
    * Temperature readings are every ~1 sec
    * CO2 readings are every ~4 seconds

## Usage

### Build:

```bash
$ go build GoogleIoT.go
```

### Run

```bash
$ ./GoogleIoT -ca_certs roots.pem -device <device_name> -private_key <key>.pem -project <project> -region <region> -registry <registry> -influx_db <measurement> -format [ line | json ]
```

### Output

Sample output on device without a K30 sensor: 

```bash
2018/09/15 12:45:58 [main] Humidity: 66.24 Temperature: 22.56
2018/09/15 12:45:59 [main] Humidity: 66.26 Temperature: 22.56
2018/09/15 12:46:01 [main] Humidity: 66.24 Temperature: 22.57
2018/09/15 12:46:02 [main] Humidity: 66.23 Temperature: 22.56
2018/09/15 12:46:04 [main] testingGoogle,sensor=bme_280 temp_c=22.56,humidity=66.24 1537029958025504878
testingGoogle,sensor=bme_280 temp_c=22.56,humidity=66.26 1537029959527147539
testingGoogle,sensor=bme_280 temp_c=22.57,humidity=66.24 1537029961028724629
testingGoogle,sensor=bme_280 temp_c=22.56,humidity=66.23 1537029962530306772
```

Note that the output is 4 lines of Influx Line Protocol, so 4 temp and humidity points are written with each iteration. 

This writes data to the Google Core IoT MQTT broker, which then sends it to the Google Pub/Sub agent, which pushes to Telegraf (using GoogleIoT branch of Telegraf from [davidgs](https://github.com/davidgs/telegraf/tree/GoogleIoT))

## Acknowledgements 

Thanks to article from [Daz Wilkins](https://medium.com/google-cloud/google-cloud-iot-core-golang-b130f65951ba) for 90% of the code! Couldn't find it on his Github, so ... good developers copy, great developers paste. :-) 