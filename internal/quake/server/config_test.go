package server

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

const expectedConfig = `seta fraglimit "25"
seta timelimit "15"
seta g_forcerespawn "0"
seta g_gametype "0"
seta g_inactivity "600"
seta g_log ""
seta g_motd "Welcome to Critical Stack"
seta g_quadfactor "3"
seta g_weaponrespawn "3"
seta sv_allowDownload "0"
seta sv_hostname "quakekube"
seta sv_maxclients "12"
seta rconpassword "changeme"
`

func TestConfigMarshal(t *testing.T) {
	c := Default()

	data, err := c.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("data = %s\n", data)
	if diff := cmp.Diff(string(data), expectedConfig); diff != "" {
		t.Fatalf(diff)
	}
}

const mapConfig = `- name: q3dm7
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

const expectedMapConfig = `set d0 "seta g_gametype 0 ; map q3dm7 ; set nextmap vstr d1"
set d1 "seta g_gametype 0 ; map q3dm17 ; set nextmap vstr d2"
set d2 "seta g_gametype 4 ; capturelimit 8 ; map q3wctf1 ; set nextmap vstr d3"
set d3 "seta g_gametype 1 ; map q3tourney2 ; set nextmap vstr d4"
set d4 "seta g_gametype 4 ; capturelimit 8 ; map q3wctf3 ; set nextmap vstr d5"
set d5 "seta g_gametype 1 ; map ztn3tourney1 ; set nextmap vstr d0"
vstr d0`

func TestMapRead(t *testing.T) {
	var maps Maps
	if err := yaml.Unmarshal([]byte(mapConfig), &maps); err != nil {
		t.Fatal(err)
	}

	for _, m := range maps {
		fmt.Printf("m = %+v\n", m)
	}
	data, err := maps.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(data), expectedMapConfig); diff != "" {
		t.Fatalf(diff)
	}
}
