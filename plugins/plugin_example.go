package plugins

import (
	"fmt"
	"github.com/Steins-Lab/Amadeus-SDK/entity"
)

// Linux go build -buildmode=plugin -o myplugin.so myplugin.go
// Windows go build -buildmode=plugin -o myplugin.dll myplugin.go

type PluginInterface struct{}

func (p *PluginInterface) SetCommunication(comm entity.Communication) {
	go func() {
		for msg := range comm.ReceiveMessage() {
			fmt.Println(msg)

		}
	}()
}

func (p *PluginInterface) Install() {

}

func (p *PluginInterface) Uninstall() {

}

func (p *PluginInterface) Name() string {
	return "修仙"
}

func (p *PluginInterface) Version() string {
	return "1.0.0"
}

// 必须导出的初始化函数
func NewPlugin() entity.Plugin {
	return &PluginInterface{}
}
