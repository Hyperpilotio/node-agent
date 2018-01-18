package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	configPath := flag.String("config", "", "The file path to a config file")
	flag.Parse()

	setDefault()
	config, err := ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Unable to read configure file: %s", err.Error())

	}

	//if *configPath == "" {
	//	*configPath = "/etc/node_agent/tasks.json"
	//}

	nodeAgent, err := NewNodeAgent(config)
	if err != nil {
		log.Errorf("create agent fail: %s", err.Error())
		return
	}

	if err := nodeAgent.Init(); err != nil {
		log.Error("Agent Init() fail: %s", err.Error())
	}

	log.Infof("Node Agent start running...")
	nodeAgent.Run()
	sig := make(chan os.Signal, 1)
	<-sig
	log.Infof("Node Agent receives OS signal to shutdown")
}

func setDefault() {
	//Default
	viper.SetDefault("APIServerPort", 7000)
	viper.SetDefault("TaskConfiguration", "/etc/node_agent/tasks.json")
	viper.SetDefault("PublisherQueueSize", 100)
	viper.SetDefault("PublisherTimeOut", "3m")
	viper.SetDefault("PublisherBatchSize", 10)
}

func ReadConfig(fileConfig string) (*viper.Viper, error) {
	viper := viper.New()
	viper.SetConfigType("json")

	if fileConfig == "" {
		viper.SetConfigName("agent_config")
		viper.AddConfigPath("/etc/node_agent")
	} else {
		viper.SetConfigFile(fileConfig)
	}

	// overwrite by file
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return viper, nil
}
