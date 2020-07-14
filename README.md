# Network Game Simulation

This is a simulation of a very terrible game to experiment with network game programming. In particular, I was interested how this works as network latency changes. [Try it in your browser](https://www.evanjones.ca/network-game-simulation-demo.html). The original version was created with GopherJS, but when I picked it up again I decided to use WASM. See my [blog post for details](https://www.evanjones.ca/network-game-simulation.html).


## Go WASM Resources

* https://github.com/golang/go/wiki/WebAssembly
* https://github.com/markfarnan/go-canvas


## Network Programming Resources

The key one is:
https://gafferongames.com/post/networked_physics_2004/

Other articles:
https://gafferongames.com/post/state_synchronization/

https://www.gabrielgambetta.com/client-server-game-architecture.html

https://fabiensanglard.net/quake3/network.php
https://fabiensanglard.net/quakeSource/quakeSourceNetWork.php

Half Life networking model: https://developer.valvesoftware.com/wiki/Latency_Compensating_Methods_in_Client/Server_In-game_Protocol_Design_and_Optimization

A good explanation of how players are "ahead" of the real time: https://www.ra.is/unlagged/solution.html


## TODO for networking

Terrible model:
* On client frame: send input to server (direction + "should fire"); Assume reliable
* Server tick: process all queued input, send state of work
* Client just displays server ticks


* Client input goes to game server (key down, key up, etc)
* Server streams world ticks to client
* Client updates world at ticks


Quake's original LAN network model:

* client is effectively "dumb".
* The server runs at 20 FPS (50 ms) Source: https://fabiensanglard.net/quakeSource/johnc-log.aug.htm
* "In the shipping version of Quake, some latency was introduced on purpose to keep the displayed frame simulation time between the last two packets from the server so that the variability in arrival time could be smoothed out."
