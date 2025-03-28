package plugins

import (
	"errors"
	"fmt"
	"github.com/Steins-Lab/Amadeus-SDK/entity"
	"os"
	"plugin"
)

type PluginCommunication struct {
	entity.PluginCommunication
	// 可以在这里添加额外字段
}

type PluginManager struct {
	*entity.PluginManager
}

func (pc *PluginCommunication) SendMessage(to string, message interface{}) {
	// 这里可以扩展为定向消息
	pc.SendCh <- message
}

func (pc *PluginCommunication) ReceiveMessage() <-chan interface{} {
	return pc.ReceiveCh
}

// 修改 PluginManager 添加通信支持
func (pm *PluginManager) SetCommunication(name string, comm entity.Communication) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()

	if lp, exists := pm.Plugins[name]; exists {
		lp.Instance.SetCommunication(comm)
	}
}

// 创建新插件管理器
func NewPluginManager() *entity.PluginManager {
	return &entity.PluginManager{
		Plugins: make(map[string]*entity.LoadedPlugin),
	}
}

// 加载插件
func (pm *PluginManager) LoadPlugin(path string) error {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()

	// 打开插件文件
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开插件文件失败: %w", err)
	}

	// 创建插件实例
	p, err := plugin.Open(path)
	if err != nil {
		err := file.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("加载插件失败: %w", err)
	}

	// 查找初始化函数
	initFunc, err := p.Lookup("NewPlugin")
	if err != nil {
		err := file.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("找不到插件初始化函数: %w", err)
	}

	// 类型断言
	newPlugin, ok := initFunc.(func() entity.Plugin)
	if !ok {
		err := file.Close()
		if err != nil {
			return err
		}
		return errors.New("无效的插件初始化函数类型")
	}

	// 创建插件实例
	pluginInstance := newPlugin()
	name := pluginInstance.Name()

	// 注册插件
	pm.Plugins[name] = &entity.LoadedPlugin{
		Instance: pluginInstance,
		File:     file,
		Handle:   p,
	}
	// 初始化通信通道
	comm := &entity.PluginCommunication{
		SendCh:    make(chan interface{}, 10),
		ReceiveCh: make(chan interface{}, 10),
	}
	pluginInstance.SetCommunication(comm)

	fmt.Printf("插件 %s (v%s) 加载成功\n", name, pluginInstance.Version())
	return nil
}

// 卸载插件
func (pm *PluginManager) UnloadPlugin(name string) error {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()

	lp, exists := pm.Plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}

	// 关闭文件句柄
	if err := lp.File.Close(); err != nil {
		return fmt.Errorf("关闭插件文件失败: %w", err)
	}

	delete(pm.Plugins, name)
	fmt.Printf("插件 %s 已卸载\n", name)
	return nil
}

// 添加热更新方法
func (pm *PluginManager) ReloadPlugin(name string, newPath string) error {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()

	// 获取旧实例
	oldLp, exists := pm.Plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}

	// 加载新实例
	file, err := os.Open(newPath)
	if err != nil {
		return fmt.Errorf("打开新插件文件失败: %w", err)
	}

	p, err := plugin.Open(newPath)
	if err != nil {
		err := file.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("加载新插件失败: %w", err)
	}

	initFunc, err := p.Lookup("NewPlugin")
	if err != nil {
		err := file.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("找不到新插件初始化函数: %w", err)
	}

	newPlugin, ok := initFunc.(func() entity.Plugin)
	if !ok {
		err := file.Close()
		if err != nil {
			return err
		}
		return errors.New("无效的新插件初始化函数类型")
	}

	newInstance := newPlugin()

	// 替换实例（这里需要根据具体需求处理状态迁移）
	oldLp.Instance = newInstance
	oldLp.File = file
	oldLp.Handle = p

	fmt.Printf("插件 %s 已热更新\n", name)
	return nil
}

// 插件列表
func (pm *PluginManager) ListPlugins() []string {
	pm.Mu.RLock()
	defer pm.Mu.RUnlock()

	names := make([]string, 0, len(pm.Plugins))
	for name := range pm.Plugins {
		names = append(names, name)
	}
	return names
}
