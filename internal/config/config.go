package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/joscherrer/dofus-manager/internal/win32"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"golang.design/x/hotkey"
	"golang.org/x/sys/windows"
)

var (
	k           *koanf.Koanf
	f           *file.File
	configPath  string
	once        sync.Once
	changed     []chan bool
	changedLock sync.Mutex
)

func SubscribeToChanged() chan bool {
	ch := make(chan bool)
	changedLock.Lock()
	changed = append(changed, ch)
	changedLock.Unlock()
	return ch
}

func UnsubscribeFromChanged(ch chan bool) {
	changedLock.Lock()
	for i, c := range changed {
		if c != ch {
			continue
		}
		changed = slices.Delete(changed, i, i+1)
	}
	changedLock.Unlock()
}

func init() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	appConfigDir := path.Join(userConfigDir, "/dofus-companion")
	fmt.Println("appConfigDir: ", appConfigDir)

	if stat, err := os.Stat(appConfigDir); os.IsNotExist(err) || !stat.IsDir() {
		os.Mkdir(appConfigDir, os.ModePerm)
	}
	configPath = path.Join(appConfigDir, "settings.yaml")
}

func reloadFromFile() {
	// Lock during config reload
	changedLock.Lock()
	defer changedLock.Unlock()
	k = koanf.New("yaml")
	if err := k.Load(f, yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}
}

func koanfInit() {
	k = koanf.New("yaml")
	f = file.Provider(configPath)

	reloadFromFile()

	f.Watch(func(event interface{}, err error) {
		if err != nil {
			log.Fatalf("error watching config: %v", err)
		}
		fmt.Println("Config file changed")
		reloadFromFile()
		changedLock.Lock()
		for _, ch := range changed {
			ch <- true
		}
		changedLock.Unlock()
	})
}

func GetConfig() *Config {
	once.Do(koanfInit)
	changedLock.Lock()
	defer changedLock.Unlock()
	c := Config{}
	k.Unmarshal("", &c)
	return &c
}

func SetConfig(nc *Config) {
	changedLock.Lock()
	defer changedLock.Unlock()

	k = koanf.New("yaml")
	k.Load(structs.Provider(nc, "koanf"), nil)

	b, err := k.Marshal(yaml.Parser())
	if err != nil {
		log.Fatalf("error marshalling config: %v", err)
	}

	f, err := os.Create(configPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Panic(err)
		}
	}()

	w := bufio.NewWriter(f)
	n, err := w.Write(b)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("wrote %d bytes\n", n)
	if err = w.Flush(); err != nil {
		log.Panic(err)
	}
}

type WindowConfig struct {
	Order int    `json:"order"`
	Name  string `json:"name"`
}

type KeyBinds struct {
	Previous hotkey.Key `koanf:"previous"`
	Next     hotkey.Key `koanf:"next"`
}

type Config struct {
	Keys    map[string]string         `koanf:"keys"`
	Windows map[string][]WindowConfig `koanf:"windows"`
}

func (c *Config) GetKey(name string) hotkey.Key {
	if c.Keys == nil {
		log.Panic("no keys")
	}
	k, ok := c.Keys[name]
	if !ok {
		log.Panic("no key " + name)
	}
	parts := strings.Split(k, "+")
	if len(parts) == 1 {
		return getKeyFromName(parts[0])
	}
	mods := parts[:len(parts)-1]
	fmt.Println("mods: ", mods)
	key := parts[len(parts)-1]
	return getKeyFromName(key)
}

func (c *Config) GetKeyName(name string) string {
	if c.Keys == nil {
		log.Panic("no keys")
	}
	k, ok := c.Keys[name]
	if !ok {
		log.Panic("no key " + name)
	}
	return k
}

func (c *Config) SetKey(name string, key string) {
	for k, v := range c.Keys {
		if k != name && v == key {
			log.Printf("key %s already bound to %s\n", key, k)
			return
		}
	}
	c.Keys[name] = key
}

// func KeyToVk(k string) hotkey.Key {
// 	return getKeyFromName(k)
// }
//
// func VkToKey(vk hotkey.Key) string {
// 	return getNameFromKey(vk)
// }

func getKeyFromName(keyName string) hotkey.Key {
	if len(keyName) == 1 {
		char := []byte(keyName)[0]
		vkCode, _ := win32.VkKeyScanExA(char, windows.GetKeyboardLayout(0))
		if vkCode != -1 {
			return hotkey.Key(vkCode)
		}
	}

	keyMap := map[string]hotkey.Key{
		"Space":       hotkey.KeySpace,
		"Return":      hotkey.KeyReturn,
		"Escape":      hotkey.KeyEscape,
		"Delete":      hotkey.KeyDelete,
		"Tab":         hotkey.KeyTab,
		"Left Arrow":  hotkey.KeyLeft,
		"Right Arrow": hotkey.KeyRight,
		"Up Arrow":    hotkey.KeyUp,
		"Down Arrow":  hotkey.KeyDown,
		"F1":          hotkey.KeyF1,
		"F2":          hotkey.KeyF2,
		"F3":          hotkey.KeyF3,
		"F4":          hotkey.KeyF4,
		"F5":          hotkey.KeyF5,
		"F6":          hotkey.KeyF6,
		"F7":          hotkey.KeyF7,
		"F8":          hotkey.KeyF8,
		"F9":          hotkey.KeyF9,
		"F10":         hotkey.KeyF10,
		"F11":         hotkey.KeyF11,
		"F12":         hotkey.KeyF12,
		"F13":         hotkey.KeyF13,
		"F14":         hotkey.KeyF14,
		"F15":         hotkey.KeyF15,
		"F16":         hotkey.KeyF16,
		"F17":         hotkey.KeyF17,
		"F18":         hotkey.KeyF18,
		"F19":         hotkey.KeyF19,
		"F20":         hotkey.KeyF20,
	}

	if vkCode, exists := keyMap[keyName]; exists {
		return vkCode
	}

	panic("unsupported key")
}

func getNameFromKey(vk hotkey.Key) string {
	reverseKeyMap := map[hotkey.Key]string{
		hotkey.KeySpace:  "Space",
		hotkey.KeyReturn: "Return",
		hotkey.KeyEscape: "Escape",
		hotkey.KeyDelete: "Delete",
		hotkey.KeyTab:    "Tab",
		hotkey.KeyLeft:   "Left Arrow",
		hotkey.KeyRight:  "Right Arrow",
		hotkey.KeyUp:     "Up Arrow",
		hotkey.KeyDown:   "Down Arrow",
		hotkey.KeyF1:     "F1",
		hotkey.KeyF2:     "F2",
		hotkey.KeyF3:     "F3",
		hotkey.KeyF4:     "F4",
		hotkey.KeyF5:     "F5",
		hotkey.KeyF6:     "F6",
		hotkey.KeyF7:     "F7",
		hotkey.KeyF8:     "F8",
		hotkey.KeyF9:     "F9",
		hotkey.KeyF10:    "F10",
		hotkey.KeyF11:    "F11",
		hotkey.KeyF12:    "F12",
		hotkey.KeyF13:    "F13",
		hotkey.KeyF14:    "F14",
		hotkey.KeyF15:    "F15",
		hotkey.KeyF16:    "F16",
		hotkey.KeyF17:    "F17",
		hotkey.KeyF18:    "F18",
		hotkey.KeyF19:    "F19",
		hotkey.KeyF20:    "F20",
	}

	if keyName, exists := reverseKeyMap[hotkey.Key(vk)]; exists {
		return keyName
	}

	scanCode := win32.MapVirtualKeyA(uint32(vk), win32.MAPVK_VK_TO_VSC)

	stateBuf := make([]byte, 256)
	buf := make([]uint16, 256)
	ret := windows.ToUnicodeEx(uint32(vk), scanCode, &stateBuf[0], &buf[0], 256, 0, windows.GetKeyboardLayout(0))

	if ret <= 0 {
		log.Panic("unsupported key")
	}

	return syscall.UTF16ToString(buf)
}
