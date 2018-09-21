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

## InfluxData Compponents

![InfluxData TICK Stack Overview](img/Tick-Stack-Complete.png "The InfluxData Tick Stack")

This tuttorial will assume that you have an instance of InfluxDB running that is accessible. 

We will be also be building and installing a version of Telegraf which supports the Google Core IoT Pub/Sub
client. That instance of Telegraf *must* be running on a publicly-accessible internet server that is 
available with an SSL encryption certificate. **Google Pub/Sub can *only* write to SSL-secured instances
of Telegraf**


### InfluxData Telegraf

Right now, the Google Core IoT Plugin for Telegraf has not been released in a 'released' version of Telegraf. This is a preview release and hence must be built from source before runninng. 

#### Prerequisites:

* Golang

Installing and configuring Go is beyond the scope of this tutorial, but you can find detailed instructions 
for your operating system [here](https://golang.org/doc/install). 

**Note**: It is possible to build Telegraf on your local machine, and run it on your internet-accessible server
by setting some GO environment variables. 

#### Building Telegraf

First, you will want to clone the GoogleIoTCore branch of Telegaf to your machine: 

```bash
$ git clone https://github.com/davidgs/telegraf/tree/GoogleIoT
```
Will get you the Google IoT branch. 

```bash
$ cd GoogleIoT
$ make
```
Will build the Telegraf agent for your platform. If you are building on a Mac OS X machine, but plan to run it
on a Linux machine, you can still build it locally:

```bash
$ export GOOS=linux
$ export GOARCH=amd64
$ make
```
Will build a 64-bit linux version of Telegraf.

#### Configuring Telegraf

Once you have built telegraf, you will need to create, and then edit, a configuration file. 

```bash
$ telegraf config > telegraf.conf
```
You will output a default configuration file. You'll beed to edit the google-specific section

```toml
  ## Server listens on <server name>:port
  ## Address and port to host HTTP listener on
  service_address = ":9999"

  ## Path to serve
  ## default is /write
  ## path = "/write"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  # precision of the time stamps. can be one of the following:
  # second, millisecond, microsecond, nanosecond
  # Default is nanosecond

  precision = "nanosecond"

  # Data Format is either line protocol or json
  protocol="line protocol"

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

```

You **must** provide the certificate and key files! 

Once this file is complete, save it to ```/etc/telegraf/telegraf.conf``` and then you can start telegraf:

```bash
$ sudo telegraf -config /etc/telegraf/telegraf.conf
```

You can verify that your telegraf agent is listening with ```$ netstat -nlp``` and look for the output line: 

```
tcp6     129      0 :::9999                 :::*                    LISTEN      -
```

### Google Core IoT Components

You will need to login to your Google Cloud Platform account. There's a great [Getting Started Guide](https://cloud.google.com/iot/docs/quickstart) to get you started. 

#### Create a Registry

You will need to create a Google Core IoT Registry:

![Create A Registry](img/Cloud.gif "Creating a Google Core IoT Registry")

Once you've created your Public/Private key pair and cert for your device, you can create a device 