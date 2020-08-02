# QuakeKube

QuakeKube is a Kubernetes-ified version of [QuakeJS](https://github.com/inolen/quakejs) that runs a dedicated [Quake 3](https://en.wikipedia.org/wiki/Quake_III_Arena) server in a Kubernetes Deployment, and then allow clients to connect via QuakeJS in the browser.

## Quick start

Start an instance of Kubernetes locally using cinder (or kind):

```shell
$ cinder create cluster
```

Deploy the example manifest:

```shell
$ kubectl apply -f example.yaml
```

Finally, navigate to the `http://$(cinder get ip):30001` in the browser.

## How it works

QuakeKube makes use [ioquake](https://www.ioquake.org) for the Quake 3 dedicated server, and [QuakeJS](https://github.com/inolen/quakejs), a port of ioquake to javascript using [Emscripten](http://github.com/kripken/emscripten), to provide an in-browser game client.

### Networking

The client/server protocol of Quake 3 uses UDP to synchronize game state. Browsers do not natively supporting sending UDP packets so QuakeJS wraps the client and dedicated server net code in websockets, allowing the browser-based clients to send messages and enable multiplayer for other clients. This ends up preventing the browser client from using any other Quake 3 dedicated server. In order to use other Quake 3 dedicated servers, a proxy handles websocket traffic coming from browser clients and translates that into UDP to the backend. This gives the flexibility of being able to talk to other existing Quake 3 servers, but also allows using ioquake (instead of the javascript translation of it), which uses *considerably* less CPU and memory.

QuakeKube also uses a cool trick with [cmux](https://github.com/cockroachdb/cmux) to multiplex the client and websocket traffic into the same connection. Having all the traffic go through the same address makes routing a client to its backend much easier (since it can just use its `document.location.host`).

### Quake 3 demo EULA

The Quake 3 dedicated server requires an End-User License Agreement be agreed to by the user before distributing the Quake 3 demo files that are used (maps, textures, etc). To ensure that the installer is aware of, and agrees to, this EULA, the flag `--agree-eula` must be passed to `q3 server` at runtime. This flag is not set by default in the container image and is therefore required for the dedicated server to pass the prompt for EULA. The [example.yaml](example.yaml) manifest demonstrates usage of this flag to agree to the EULA.

## Configuration

### Server configuration

The server configuration is set by ConfigMap that is mounted to the container:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: quake3-server-config
data:
  server.cfg: |
    seta sv_hostname "quakekube"
    seta g_log ""
    seta sv_maxclients 12
    seta g_motd "Welcome to Critical Stack"
    seta g_quadfactor 3
    seta timelimit 15
    seta fraglimit 25
    seta g_weaponrespawn 3
    seta g_inactivity 600
    seta g_forcerespawn 0
    seta rconpassword "changeme"
```

Many of the config values seen [here](http://www.joz3d.net/html/q3console.html) will likely work, but be careful to not create an invalid server configuration (causing the pod to crash).

### Maps

The maps are configured via the server configuration ConfigMap by specifying a maps file that can be mounted into the container:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: quake3-server-config
data:
  ...
  maps.yaml: |
    - name: q3dm7
      type: FreeForAll
    - name: q3dm17
      type: FreeForAll
```

These values are then used when specifying container arguments in the Kubernetes Deployment via `--maps`. For example usage,  be sure to check out the [example.yaml](example.yaml) manifest.

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

## Add new maps

The content server hosts a small upload app to allow uploading `pk3` or `zip` files containing maps. The server must be restarted after maps have been added to be available. Currently, this happens after the map rotates to one of the previously loaded maps, however, in the future the server will respond to changes to the ConfigMap automatically.

## Credits

* [inolen/quakejs](https://github.com/inolen/quakejs) - The really awesome QuakeJS project that makes this possible.
* [ioquake/ioq3](https://github.com/ioquake/ioq3) - The community supported version of Quake 3 used by QuakeJS. It is licensed under the GPLv2.
* [begleysm/quakejs](https://github.com/begleysm/quakejs) - Information in the README.md (very helpful) was used as a guide, as well as, some forked assets of this project (which came from quakejs-web originally) were used.
* [joz3d.net](http://www.joz3d.net/html/q3console.html) - Useful information about configuration values.
