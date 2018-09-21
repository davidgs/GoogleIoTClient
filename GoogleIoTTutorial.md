# Integrating Google Core IoT with InfluxData for IoT Data collection
 This tutorial will walk through the steps necessary to integrat your Google Core IoT device with an InfluxData
 data collection, visualization and alerting platform to fully enable your IoT application. 

 What you'll need to complete this tutorial

 * [Google Cloud](https://console.cloud.google.com/) Account
 * [InfluxDB](https://influxdata.com) instance available on the internet
 * [Raspberry Pi 3](http://www.raspberrypi.org) device
 * [Bosch BME 280](https://www.adafruit.com/product/2652) sensor board from [Adafruit](https://adafruit.com)

All of the code required for this tutorial is available in the following Github repositories:

* [Rasperry Pi Code](https://github.com/davidgs/telegraf/tree/GoogleIoT)
* [Telegraf Plugin](https://github.com/davidgs/telegraf/tree/GoogleIoT/plugins/inputs/googlecoreiot)

## Compponents

![InfluxData TICK Stack Overview](file://./img/Tick-Stack-Complete.png "The InfluxData Tick Stack")


### InfluxData Telegraf

Right now, the Google Core IoT Plugin for Telegraf has 