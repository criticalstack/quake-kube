# Due to changes in the priorities, this project is currently not being supported. The project is archived as of 11/17/21 and will be available in a read-only state. Please note, since archival, the project is not maintained or reviewed. #

![Build Status](https://github.com/criticalstack/quake-kube/workflows/Push%20Image/badge.svg)

# QuakeKube

QuakeKube is a Kubernetes-ified version of [QuakeJS](https://github.com/inolen/quakejs) that runs a dedicated [Quake 3](https://en.wikipedia.org/wiki/Quake_III_Arena) server in a Kubernetes Deployment, and then allow clients to connect via QuakeJS in the browser.

## Quick start

### With an existing K8s cluster

Deploy the example manifest:

```shell
$ kubectl apply -f https://raw.githubusercontent.com/criticalstack/quake-kube/master/example.yaml
```


### Without an existing K8s cluster

Start an instance of Kubernetes locally using [cinder](https://docs.crit.sh/cinder-guide/what-is-cinder.html) (or [kind](https://kind.sigs.k8s.io/)):

```shell
$ cinder create cluster
```

Deploy the example manifest:

```shell
$ kubectl apply -f example.yaml
```

Finally, navigate to `http://$(cinder get ip):30001` in the browser.

## How it works

QuakeKube makes use of [ioquake](https://www.ioquake.org) for the Quake 3 dedicated server, and [QuakeJS](https://github.com/inolen/quakejs), a port of ioquake to javascript using [Emscripten](http://github.com/kripken/emscripten), to provide an in-browser game client.

### Networking

The client/server protocol of Quake 3 uses UDP to synchronize game state. Browsers do not natively support sending UDP packets so QuakeJS wraps the client and dedicated server net code in websockets, allowing the browser-based clients to send messages and enable multiplayer for other clients. This ends up preventing the browser client from using any other Quake 3 dedicated server. In order to use other Quake 3 dedicated servers, a proxy handles websocket traffic coming from browser clients and translates that into UDP to the backend. This gives the flexibility of being able to talk to other existing Quake 3 servers, but also allows using ioquake (instead of the javascript translation of it), which uses *considerably* less CPU and memory.

QuakeKube also uses a cool trick with [cmux](https://github.com/cockroachdb/cmux) to multiplex the client and websocket traffic into the same connection. Having all the traffic go through the same address makes routing a client to its backend much easier (since it can just use its `document.location.host`).

### Quake 3 demo EULA

The Quake 3 dedicated server requires an End-User License Agreement be agreed to by the user before distributing the Quake 3 demo files that are used (maps, textures, etc). To ensure that the installer is aware of, and agrees to, this EULA, the flag `--agree-eula` must be passed to `q3 server` at runtime. This flag is not set by default in the container image and is therefore required for the dedicated server to pass the prompt for EULA. The [example.yaml](example.yaml) manifest demonstrates usage of this flag to agree to the EULA.

## Configuration

The server and maps are configured via ConfigMap that is mounted to the container:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: quake3-server-config
data:
  config.yaml: |
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
```

The time limit and frag limit can be specified with each map (it will change it for subsequent maps in the list):

```yaml
- name: q3dm17
  type: FreeForAll
  fragLimit: 30
  timeLimit: 30
```

Capture limit for CTF maps can also be configured:

```yaml
- name: q3wctf3
  type: CaptureTheFlag
  captureLimit: 8
```

Any commands not captured by the config yaml can be specified in the `commands` section:

```yaml
commands:
- seta g_inactivity 600
- seta sv_timeout 120
```

### Add bots

Bots can be added individually to map rotations using the `commands` section of the config:

```yaml
commands:
  - addbot crash 1
  - addbot sarge 2
```

The `addbot` server command requires the name of the bot and skill level (crash and sarge are a couple of the built-in bots).

Another way to add bots is by setting a minimum number of players to allow the server to add bots up to a certain value (removed when human players join):

```yaml
bot:
  minPlayers: 8
game:
  singlePlayerSkill: 2
```

`singlePlayerSkill` can be used to set the skill level of the automatically added bots (2 is the default skill level).

### Setting a password

A password should be set for the server to allow remote administration and is found in the server configuration settings:

```yaml
server:
  password: "changeme"
```

This will allow clients to use `\rcon changeme <cmd>` to remotely administrate the server. To create a password that must be provided by clients to connect:

```yaml
game:
  password: "letmein"
```

This will add an additional dialog to the in-browser client to accept the password. It will only appear if the server indicates it needs a password.

### Add custom maps

The content server hosts a small upload app to allow uploading `pk3` or `zip` files containing maps. The content server in the [example.yaml](example.yaml) shares a volume with the game server, effectively "side-loading" the map content, however, in the future the game server will introspect into the maps and make sure that it can fulfill the users map configuration before starting.

### Development

The easiest way to develop quake-kube is building the binary locally with `make` and running it directly. This only requires that you have the `ioq3ded` binary in your path:

```shell
$ bin/q3 server -c config.yaml --assets-dir $HOME/.q3a --agree-eula
```

### Multi-platform images

Container images are being cross-compiled with [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/) so it can run on hardware with different architectures and operating systems. Currently, it is building for `linux/amd64` and `linux/arm64`. While not specifically compiling to the macOS platform (`darwin/amd64`) QuakeKube should also work on macOS and maybe even Windows. This is due to the fact that they both use a linux VM to provide container support.

Docker Buildx uses [QEMU](https://www.qemu.org/) to virtualize non-native platforms, which has unfortunately had long-running issues running the Go compiler:

* [golang/go#24656](https://github.com/golang/go/issues/24656)
* [https://bugs.launchpad.net/qemu/+bug/1696773](https://bugs.launchpad.net/qemu/+bug/1696773)

This issue is circumvented by ensuring that the Go compiler does not run across multiple hardware threads, which is why the affinity is being limited in the Dockerfile.

## Credits

* [inolen/quakejs](https://github.com/inolen/quakejs) - The really awesome QuakeJS project that makes this possible.
* [ioquake/ioq3](https://github.com/ioquake/ioq3) - The community supported version of Quake 3 used by QuakeJS. It is licensed under the GPLv2.
* [begleysm/quakejs](https://github.com/begleysm/quakejs) - Information in the README.md (very helpful) was used as a guide, as well as, some forked assets of this project (which came from quakejs-web originally) were used.
* [joz3d.net](http://www.joz3d.net/html/q3console.html) - Useful information about configuration values.
