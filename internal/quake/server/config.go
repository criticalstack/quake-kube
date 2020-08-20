package server

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	FragLimit int             `name:"fraglimit"`
	TimeLimit metav1.Duration `name:"timelimit"`

	BotConfig        `json:"bot"`
	GameConfig       `json:"game"`
	FileServerConfig `json:"fs"`
	ServerConfig     `json:"server"`
	Commands         []string `json:"commands"`

	Maps
}

type BotConfig struct {
	MinPlayers int  `name:"bot_minplayers"`
	NoChat     bool `name:"bot_nochat"`
}

type GameConfig struct {
	ForceRespawn      bool            `name:"g_forcerespawn"`
	GameType          GameType        `json:"type" name:"g_gametype"`
	Inactivity        metav1.Duration `name:"g_inactivity"`
	Log               string          `name:"g_log"`
	MOTD              string          `name:"g_motd"`
	Password          string          `name:"g_password"`
	QuadFactor        int             `name:"g_quadfactor"`
	SinglePlayerSkill int             `name:"g_spSkill"`
	WeaponRespawn     int             `name:"g_weaponrespawn"`
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
		case reflect.Slice:
			switch val := fv.Interface().(type) {
			case Maps:
				data, _ := val.Marshal()
				b.Write(data)
			case []string:
			default:
				panic(fmt.Errorf("received unknown type %T", val))
			}
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
	for i := 0; i < v.Type().NumField(); i++ {
		if v.Type().Field(i).Name == "Commands" {
			cmds := v.Field(i).Interface().([]string)
			for _, cmd := range cmds {
				b.WriteString(cmd)
				b.WriteString("\n")
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
	case metav1.Duration:
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
	case Maps:
		data, _ := val.Marshal()
		return string(data)
	default:
		panic(fmt.Errorf("received unknown type %T", v.Interface()))
	}
}

func Default() *Config {
	return &Config{
		FragLimit: 25,
		TimeLimit: metav1.Duration{Duration: 15 * time.Minute},
		Commands:  []string{},
		BotConfig: BotConfig{
			NoChat: true,
		},
		GameConfig: GameConfig{
			Log:               "",
			MOTD:              "Welcome to Critical Stack",
			QuadFactor:        3,
			GameType:          FreeForAll,
			WeaponRespawn:     3,
			Inactivity:        metav1.Duration{Duration: 10 * time.Minute},
			SinglePlayerSkill: 2,
			ForceRespawn:      false,
		},
		ServerConfig: ServerConfig{
			MaxClients: 12,
			Hostname:   "quakekube",
			Password:   "changeme",
		},
		Maps: Maps{
			{Name: "q3dm7", Type: FreeForAll},
			{Name: "q3dm17", Type: FreeForAll},
		},
	}
}

type Map struct {
	Name string   `json:"name"`
	Type GameType `json:"type"`

	CaptureLimit int             `json:"captureLimit"`
	FragLimit    int             `json:"fragLimit"`
	TimeLimit    metav1.Duration `json:"timeLimit"`
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
		if m.TimeLimit.Duration != 0 {
			cmds = append(cmds, fmt.Sprintf("timelimit %s", toString("TimeLimit", reflect.ValueOf(m.TimeLimit))))
		}
		cmds = append(cmds, fmt.Sprintf("map %s", m.Name))
		nextmap := "d0"
		if i < len(maps)-1 {
			nextmap = fmt.Sprintf("d%d", i+1)
		}
		cmds = append(cmds, fmt.Sprintf("set nextmap vstr %s", nextmap))
		b.WriteString(fmt.Sprintf("set d%d \"seta %s\"\n", i, strings.Join(cmds, " ; ")))
	}
	b.WriteString("vstr d0\n")
	return b.Bytes(), nil
}
