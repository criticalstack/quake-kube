package server

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

const config = `
fragLimit: 25
timeLimit: 15m
game:
  motd: "Welcome to Critical Stack"
  type: FreeForAll
  forceRespawn: false
  inactivity: 10m
  quadFactor: 3
  weaponRespawn: 3
server:
  hostname: "quakekube"
  maxClients: 12
  password: "changeme"
commands:
- seta g_inactivity 600
- seta sv_timeout 120
maps:
- name: q3dm7
  type: FreeForAll
- name: q3dm17
  type: FreeForAll
- name: q3wctf1
  type: CaptureTheFlag
  captureLimit: 8
- name: q3tourney2
  type: Tournament
- name: q3wctf3
  type: CaptureTheFlag
  captureLimit: 8
- name: ztn3tourney1
  type: Tournament
`

const expectedConfig = `seta fraglimit "25"
seta bot_minplayers "0"
seta bot_nochat "0"
seta g_forcerespawn "0"
seta g_gametype "0"
seta g_log ""
seta g_motd "Welcome to Critical Stack"
seta g_password ""
seta g_quadfactor "3"
seta g_spSkill "0"
seta g_weaponrespawn "3"
seta fs_basegame ""
seta fs_basepath ""
seta fs_copyfiles "0"
seta fs_debug "0"
seta fs_game ""
seta fs_homepath ""
seta sv_allowDownload "0"
seta sv_hostname "quakekube"
seta sv_maxclients "12"
seta rconpassword "changeme"
set d0 "seta g_gametype 0 ; map q3dm7 ; set nextmap vstr d1"
set d1 "seta g_gametype 0 ; map q3dm17 ; set nextmap vstr d2"
set d2 "seta g_gametype 4 ; capturelimit 8 ; map q3wctf1 ; set nextmap vstr d3"
set d3 "seta g_gametype 1 ; map q3tourney2 ; set nextmap vstr d4"
set d4 "seta g_gametype 4 ; capturelimit 8 ; map q3wctf3 ; set nextmap vstr d5"
set d5 "seta g_gametype 1 ; map ztn3tourney1 ; set nextmap vstr d0"
vstr d0
seta g_inactivity 600
seta sv_timeout 120
`

func TestConfigMarshal(t *testing.T) {
	var cfg *Config
	if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
		t.Fatal(err)
	}
	data, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", data)
	if diff := cmp.Diff(string(data), expectedConfig); diff != "" {
		t.Fatalf(diff)
	}
}
