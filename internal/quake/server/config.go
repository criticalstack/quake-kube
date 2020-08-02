package server

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type GameType int

const (
	FreeForAll     GameType = 0
	Tournament     GameType = 1
	SinglePlayer   GameType = 2
	TeamDeathmatch GameType = 3
	CaptureTheFlag GameType = 4
)

func (gt GameType) String() string {
	switch gt {
	case FreeForAll:
		return "FreeForAll"
	case Tournament:
		return "Tournament"
	case SinglePlayer:
		return "SinglePlayer"
	case TeamDeathmatch:
		return "TeamDeathmatch"
	case CaptureTheFlag:
		return "CaptureTheFlag"
	default:
		return "Unknown"
	}
}

func (gt *GameType) UnmarshalText(data []byte) error {
	switch string(data) {
	case "FreeForAll", "FFA":
		*gt = FreeForAll
	case "Tournament":
		*gt = Tournament
	case "SinglePlayer":
		*gt = SinglePlayer
	case "TeamDeathmatch":
		*gt = TeamDeathmatch
	case "CaptureTheFlag", "CTF":
		*gt = CaptureTheFlag
	default:
		return errors.Errorf("unknown GameType: %s", data)
	}
	return nil
}

type Config struct {
	FragLimit int           `name:"fraglimit"`
	TimeLimit time.Duration `name:"timelimit"`

	GameConfig
	ServerConfig
}

type GameConfig struct {
	ForceRespawn  bool          `name:"g_forcerespawn"`
	GameType      GameType      `name:"g_gametype"`
	Inactivity    time.Duration `name:"g_inactivity"`
	Log           string        `name:"g_log"`
	MOTD          string        `name:"g_motd"`
	QuadFactor    int           `name:"g_quadfactor"`
	WeaponRespawn int           `name:"g_weaponrespawn"`
}

type FileServerConfig struct {
	// allows people to base mods upon mods syntax to follow
	BaseGame string `name:"fs_basegame"`
	// set base path root C:\Program Files\Quake III Arena for files to be
	// downloaded from this path may change for TC's and MOD's
	BasePath string `name:"fs_basepath"`
	// toggle if files can be copied from servers or if client will download
	CopyFiles bool `name:"fs_copyfiles"`
	// possibly enables file server debug mode for download/uploads or
	// something
	Debug bool `name:"fs_debug"`
	// set gamedir set the game folder/dir default is baseq3
	Game string `name:"fs_game"`
	// possibly for TC's and MODS the default is the path to quake3.exe
	HomePath string `name:"fs_homepath"`
}

type ServerConfig struct {
	AllowDownload bool   `name:"sv_allowDownload"`
	DownloadURL   string `name:"sv_dlURL"`
	Hostname      string `name:"sv_hostname"`
	MaxClients    int    `name:"sv_maxclients"`
	Password      string `name:"rconpassword"`
}

func (c *Config) Marshal() ([]byte, error) {
	return writeStruct(reflect.Indirect(reflect.ValueOf(c)))
}

func writeStruct(v reflect.Value) ([]byte, error) {
	if v.Kind() != reflect.Struct {
		return nil, errors.Errorf("expected struct, received %T", v.Kind())
	}
	var b bytes.Buffer
	for i := 0; i < v.Type().NumField(); i++ {
		fv := v.Field(i)
		switch fv.Kind() {
		case reflect.Struct:
			data, err := writeStruct(fv)
			if err != nil {
				return nil, err
			}
			b.Write(data)
		default:
			tv, ok := v.Type().Field(i).Tag.Lookup("name")
			if !ok {
				continue
			}
			s := toString(v.Type().Field(i).Name, fv)
			switch tv {
			case "sv_dlURL":
				if s != "" {
					b.WriteString(fmt.Sprintf("sets %s %s\n", tv, s))
				}
			default:
				b.WriteString(fmt.Sprintf("seta %s %s\n", tv, strconv.Quote(s)))
			}
		}
	}
	return b.Bytes(), nil
}

func toString(name string, v reflect.Value) string {
	switch val := v.Interface().(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case time.Duration:
		switch name {
		case "TimeLimit":
			return fmt.Sprintf("%d", int(val.Minutes()))
		default:
			return fmt.Sprintf("%d", int(val.Seconds()))
		}
	case bool:
		if val {
			return "1"
		}
		return "0"
	case GameType:
		return fmt.Sprintf("%d", val)
	default:
		panic(fmt.Errorf("received unknown type %T", v.Interface()))
	}
}

func Default() *Config {
	return &Config{
		TimeLimit: 15 * time.Minute,
		FragLimit: 25,
		GameConfig: GameConfig{
			Log:           "",
			MOTD:          "Welcome to Critical Stack",
			QuadFactor:    3,
			GameType:      FreeForAll,
			WeaponRespawn: 3,
			Inactivity:    10 * time.Minute,
			ForceRespawn:  false,
		},
		ServerConfig: ServerConfig{
			MaxClients: 12,
			Hostname:   "quakekube",
			Password:   "changeme",
		},
	}
}

type Maps []Map

func (maps Maps) Marshal() ([]byte, error) {
	var b bytes.Buffer
	for i, m := range maps {
		cmds := []string{
			fmt.Sprintf("g_gametype %d", m.Type),
		}
		if m.Type == CaptureTheFlag && m.CaptureLimit != 0 {
			cmds = append(cmds, fmt.Sprintf("capturelimit %d", m.CaptureLimit))
		}
		if m.FragLimit != 0 {
			cmds = append(cmds, fmt.Sprintf("fraglimit %d", m.FragLimit))
		}
		if m.TimeLimit != 0 {
			cmds = append(cmds, fmt.Sprintf("timelimit %d", int(m.TimeLimit.Minutes())))
		}
		cmds = append(cmds, fmt.Sprintf("map %s", m.Name))
		nextmap := "d0"
		if i < len(maps)-1 {
			nextmap = fmt.Sprintf("d%d", i+1)
		}
		cmds = append(cmds, fmt.Sprintf("set nextmap vstr %s", nextmap))
		b.WriteString(fmt.Sprintf("set d%d \"seta %s\"\n", i, strings.Join(cmds, " ; ")))
	}
	b.WriteString("vstr d0")
	return b.Bytes(), nil
}

type Map struct {
	Name string   `json:"name"`
	Type GameType `json:"type"`

	CaptureLimit int           `json:"captureLimit"`
	FragLimit    int           `json:"fragLimit"`
	TimeLimit    time.Duration `json:"timeLimit"`
}
