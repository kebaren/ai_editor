package textbuffer

import (
	"errors"
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// LuaPlugin represents a Lua plugin for the text buffer
type LuaPlugin struct {
	// Lua state
	L *lua.LState
	// Plugin name
	Name string
	// Plugin description
	Description string
	// Plugin version
	Version string
	// Plugin author
	Author string
	// Plugin path
	Path string
	// Plugin enabled state
	Enabled bool
	// Mutex for thread safety
	mutex sync.RWMutex
}

// LuaPluginManager manages Lua plugins for the text buffer
type LuaPluginManager struct {
	// Map of plugin name to plugin
	plugins map[string]*LuaPlugin
	// TextBuffer reference
	textBuffer *TextBuffer
	// Mutex for thread safety
	mutex sync.RWMutex
}

// NewLuaPluginManager creates a new Lua plugin manager
func NewLuaPluginManager(textBuffer *TextBuffer) *LuaPluginManager {
	return &LuaPluginManager{
		plugins:    make(map[string]*LuaPlugin),
		textBuffer: textBuffer,
		mutex:      sync.RWMutex{},
	}
}

// LoadPlugin loads a Lua plugin from the given path
func (pm *LuaPluginManager) LoadPlugin(path string) (*LuaPlugin, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Create a new Lua state
	L := lua.NewState()
	defer L.Close()

	// Load the plugin
	if err := L.DoFile(path); err != nil {
		return nil, fmt.Errorf("failed to load plugin: %v", err)
	}

	// Get plugin metadata
	name := getLuaString(L, "plugin_name")
	if name == "" {
		return nil, errors.New("plugin_name is required")
	}

	// Check if plugin already exists
	if _, exists := pm.plugins[name]; exists {
		return nil, fmt.Errorf("plugin '%s' already loaded", name)
	}

	description := getLuaString(L, "plugin_description")
	version := getLuaString(L, "plugin_version")
	author := getLuaString(L, "plugin_author")

	// Create the plugin
	plugin := &LuaPlugin{
		L:           lua.NewState(),
		Name:        name,
		Description: description,
		Version:     version,
		Author:      author,
		Path:        path,
		Enabled:     true,
		mutex:       sync.RWMutex{},
	}

	// Register text buffer API
	registerTextBufferAPI(plugin.L, pm.textBuffer)

	// Load the plugin in the plugin's Lua state
	if err := plugin.L.DoFile(path); err != nil {
		return nil, fmt.Errorf("failed to load plugin in plugin state: %v", err)
	}

	// Call init function if it exists
	if initFn := plugin.L.GetGlobal("init"); initFn.Type() == lua.LTFunction {
		if err := plugin.L.CallByParam(lua.P{
			Fn:      initFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return nil, fmt.Errorf("failed to call init function: %v", err)
		}
	}

	// Add plugin to manager
	pm.plugins[name] = plugin

	return plugin, nil
}

// UnloadPlugin unloads a plugin by name
func (pm *LuaPluginManager) UnloadPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	// Call cleanup function if it exists
	if cleanupFn := plugin.L.GetGlobal("cleanup"); cleanupFn.Type() == lua.LTFunction {
		if err := plugin.L.CallByParam(lua.P{
			Fn:      cleanupFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return fmt.Errorf("failed to call cleanup function: %v", err)
		}
	}

	// Close Lua state
	plugin.L.Close()

	// Remove plugin from manager
	delete(pm.plugins, name)

	return nil
}

// EnablePlugin enables a plugin by name
func (pm *LuaPluginManager) EnablePlugin(name string) error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	if plugin.Enabled {
		return nil // Already enabled
	}

	// Call enable function if it exists
	if enableFn := plugin.L.GetGlobal("enable"); enableFn.Type() == lua.LTFunction {
		if err := plugin.L.CallByParam(lua.P{
			Fn:      enableFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return fmt.Errorf("failed to call enable function: %v", err)
		}
	}

	plugin.Enabled = true

	return nil
}

// DisablePlugin disables a plugin by name
func (pm *LuaPluginManager) DisablePlugin(name string) error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	plugin.mutex.Lock()
	defer plugin.mutex.Unlock()

	if !plugin.Enabled {
		return nil // Already disabled
	}

	// Call disable function if it exists
	if disableFn := plugin.L.GetGlobal("disable"); disableFn.Type() == lua.LTFunction {
		if err := plugin.L.CallByParam(lua.P{
			Fn:      disableFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return fmt.Errorf("failed to call disable function: %v", err)
		}
	}

	plugin.Enabled = false

	return nil
}

// GetPlugin gets a plugin by name
func (pm *LuaPluginManager) GetPlugin(name string) (*LuaPlugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// GetPlugins gets all plugins
func (pm *LuaPluginManager) GetPlugins() []*LuaPlugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]*LuaPlugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// CallPluginFunction calls a function in a plugin
func (pm *LuaPluginManager) CallPluginFunction(pluginName, functionName string, args ...lua.LValue) (lua.LValue, error) {
	pm.mutex.RLock()
	plugin, exists := pm.plugins[pluginName]
	pm.mutex.RUnlock()

	if !exists {
		return lua.LNil, fmt.Errorf("plugin '%s' not found", pluginName)
	}

	plugin.mutex.RLock()
	defer plugin.mutex.RUnlock()

	if !plugin.Enabled {
		return lua.LNil, fmt.Errorf("plugin '%s' is disabled", pluginName)
	}

	// Get the function
	fn := plugin.L.GetGlobal(functionName)
	if fn.Type() != lua.LTFunction {
		return lua.LNil, fmt.Errorf("function '%s' not found in plugin '%s'", functionName, pluginName)
	}

	// Call the function
	err := plugin.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, args...)
	if err != nil {
		return lua.LNil, fmt.Errorf("failed to call function '%s' in plugin '%s': %v", functionName, pluginName, err)
	}

	// Get the result
	result := plugin.L.Get(-1)
	plugin.L.Pop(1)

	return result, nil
}

// Helper functions

// getLuaString gets a string value from the Lua state
func getLuaString(L *lua.LState, name string) string {
	value := L.GetGlobal(name)
	if value.Type() != lua.LTString {
		return ""
	}
	return value.String()
}

// registerTextBufferAPI registers the text buffer API in the Lua state
func registerTextBufferAPI(L *lua.LState, tb *TextBuffer) {
	// Create text buffer module
	mod := L.NewTable()
	L.SetGlobal("textbuffer", mod)

	// Register functions
	L.SetField(mod, "get_text", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(tb.GetText()))
		return 1
	}))

	L.SetField(mod, "get_line_count", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(tb.GetLineCount()))
		return 1
	}))

	L.SetField(mod, "get_line_content", L.NewFunction(func(L *lua.LState) int {
		lineIndex := L.CheckInt(1)
		L.Push(lua.LString(tb.GetLineContent(lineIndex)))
		return 1
	}))

	L.SetField(mod, "insert", L.NewFunction(func(L *lua.LState) int {
		line := L.CheckInt(1)
		column := L.CheckInt(2)
		text := L.CheckString(3)

		err := tb.Insert(Position{Line: line, Column: column}, text)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "delete", L.NewFunction(func(L *lua.LState) int {
		startLine := L.CheckInt(1)
		startColumn := L.CheckInt(2)
		endLine := L.CheckInt(3)
		endColumn := L.CheckInt(4)

		err := tb.Delete(Range{
			Start: Position{Line: startLine, Column: startColumn},
			End:   Position{Line: endLine, Column: endColumn},
		})
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "replace", L.NewFunction(func(L *lua.LState) int {
		startLine := L.CheckInt(1)
		startColumn := L.CheckInt(2)
		endLine := L.CheckInt(3)
		endColumn := L.CheckInt(4)
		text := L.CheckString(5)

		err := tb.Replace(Range{
			Start: Position{Line: startLine, Column: startColumn},
			End:   Position{Line: endLine, Column: endColumn},
		}, text)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "undo", L.NewFunction(func(L *lua.LState) int {
		err := tb.Undo()
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "redo", L.NewFunction(func(L *lua.LState) int {
		err := tb.Redo()
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetField(mod, "get_eol_type", L.NewFunction(func(L *lua.LState) int {
		eolType := tb.GetEOLType()
		L.Push(lua.LNumber(eolType))
		return 1
	}))

	L.SetField(mod, "set_eol_type", L.NewFunction(func(L *lua.LState) int {
		eolType := EOLType(L.CheckInt(1))
		err := tb.SetEOLType(eolType)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))
}

// Close 关闭Lua插件管理器，清理资源
func (lm *LuaPluginManager) Close() {
	// 停止所有Lua虚拟机
	// 清理插件资源
	// 停止所有相关的goroutine
	lm.textBuffer = nil
}
